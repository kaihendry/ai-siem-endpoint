package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// mockPutter captures the last PutItemInput for inspection by tests.
type mockPutter struct {
	lastInput *dynamodb.PutItemInput
	err       error
}

func (m *mockPutter) PutItem(_ context.Context, params *dynamodb.PutItemInput, _ ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	m.lastInput = params
	return &dynamodb.PutItemOutput{}, m.err
}

// TestMain installs the mock putter before any test runs so that handler tests
// do not attempt real DynamoDB calls.
func TestMain(m *testing.M) {
	eventPutter = &mockPutter{}
	m.Run()
}

// TestPutEventAttributes verifies that putEvent builds the correct DynamoDB
// attribute map for a known AuditRun value.
func TestPutEventAttributes(t *testing.T) {
	mock := &mockPutter{}
	eventPutter = mock

	run := AuditRun{
		RunID:      "test-run-1",
		Timestamp:  time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Host:       "testhost",
		Score:      90,
		ExitCode:   0,
		DurationMs: 500,
		Findings: []Finding{
			{Type: "policy", Severity: "high", Module: "guardrails"},
		},
	}

	sk, err := putEvent(context.Background(), run, "test-agent/1.0")
	if err != nil {
		t.Fatalf("putEvent returned unexpected error: %v", err)
	}

	const wantSKPrefix = "2024-01-15T10:00:00Z#test-run-1"
	if !strings.HasPrefix(sk, wantSKPrefix) {
		t.Errorf("sk = %q, want prefix %q", sk, wantSKPrefix)
	}

	if mock.lastInput == nil {
		t.Fatal("PutItem was not called")
	}
	item := mock.lastInput.Item

	// pk must be "all"
	pkAttr, ok := item["pk"].(*types.AttributeValueMemberS)
	if !ok || pkAttr.Value != "all" {
		t.Errorf("pk = %v, want \"all\"", item["pk"])
	}

	// sk must start with expected prefix
	skAttr, ok := item["sk"].(*types.AttributeValueMemberS)
	if !ok || !strings.HasPrefix(skAttr.Value, wantSKPrefix) {
		t.Errorf("sk attribute = %v, want prefix %q", item["sk"], wantSKPrefix)
	}

	// run_id
	runIDAttr, ok := item["run_id"].(*types.AttributeValueMemberS)
	if !ok || runIDAttr.Value != "test-run-1" {
		t.Errorf("run_id = %v, want \"test-run-1\"", item["run_id"])
	}

	// host
	hostAttr, ok := item["host"].(*types.AttributeValueMemberS)
	if !ok || hostAttr.Value != "testhost" {
		t.Errorf("host = %v, want \"testhost\"", item["host"])
	}

	// score must be numeric attribute "90"
	scoreAttr, ok := item["score"].(*types.AttributeValueMemberN)
	if !ok || scoreAttr.Value != "90" {
		t.Errorf("score = %v, want N(\"90\")", item["score"])
	}

	// findings must be a non-empty JSON string containing "guardrails"
	findingsAttr, ok := item["findings"].(*types.AttributeValueMemberS)
	if !ok || findingsAttr.Value == "" {
		t.Errorf("findings attribute missing or empty")
	} else if !strings.Contains(findingsAttr.Value, "guardrails") {
		t.Errorf("findings JSON %q does not contain \"guardrails\"", findingsAttr.Value)
	}
}

// TestHandlePost covers the POST / handler with a happy path, missing required
// fields, and an oversized body.
func TestHandlePost(t *testing.T) {
	// Install a fresh mock for each sub-test group.
	mock := &mockPutter{}
	eventPutter = mock

	// Helper: build a minimal valid AuditRun payload.
	validPayload := func() map[string]any {
		return map[string]any{
			"run_id":    "happy-run",
			"timestamp": "2024-06-01T12:00:00Z",
			"host":      "host1",
		}
	}

	marshal := func(v any) []byte {
		b, _ := json.Marshal(v)
		return b
	}

	postJSON := func(body []byte) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handlePost(rr, req)
		return rr
	}

	t.Run("happy path returns 201 with run_id", func(t *testing.T) {
		rr := postJSON(marshal(validPayload()))
		if rr.Code != http.StatusCreated {
			t.Errorf("status = %d, want 201; body: %s", rr.Code, rr.Body)
		}
		if !strings.Contains(rr.Body.String(), "run_id") {
			t.Errorf("response body %q does not contain \"run_id\"", rr.Body.String())
		}
	})

	t.Run("missing run_id returns 400", func(t *testing.T) {
		payload := validPayload()
		payload["run_id"] = ""
		rr := postJSON(marshal(payload))
		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400; body: %s", rr.Code, rr.Body)
		}
		if !strings.Contains(rr.Body.String(), "error") {
			t.Errorf("response body %q does not contain \"error\"", rr.Body.String())
		}
	})

	t.Run("missing timestamp returns 400", func(t *testing.T) {
		payload := map[string]any{
			"run_id": "some-run",
			"host":   "host1",
			// timestamp deliberately omitted
		}
		rr := postJSON(marshal(payload))
		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400; body: %s", rr.Code, rr.Body)
		}
		if !strings.Contains(rr.Body.String(), "error") {
			t.Errorf("response body %q does not contain \"error\"", rr.Body.String())
		}
	})

	t.Run("missing host returns 400", func(t *testing.T) {
		payload := validPayload()
		payload["host"] = ""
		rr := postJSON(marshal(payload))
		if rr.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400; body: %s", rr.Code, rr.Body)
		}
		if !strings.Contains(rr.Body.String(), "error") {
			t.Errorf("response body %q does not contain \"error\"", rr.Body.String())
		}
	})

	t.Run("body too large returns 413", func(t *testing.T) {
		// Build a JSON object whose value field exceeds the 1 MiB limit so
		// that MaxBytesReader triggers before the JSON decoder can finish.
		// Format: {"run_id":"<1 MiB + 1 bytes of 'a'>"}
		const limit = 1 << 20
		var buf bytes.Buffer
		buf.WriteString(`{"run_id":"`)
		buf.Write(bytes.Repeat([]byte("a"), limit+1))
		buf.WriteString(`"}`)
		req := httptest.NewRequest(http.MethodPost, "/", &buf)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		handlePost(rr, req)
		if rr.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("status = %d, want 413; body: %s", rr.Code, rr.Body)
		}
	})

}
