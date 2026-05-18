package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/apex/gateway/v2"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func init() {
	tableName = os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		tableName = "mock-siem-events"
	}
	dynamoClient = newDynamoClient()
}

// T006: domain types matching internal/audit/AuditRun and internal/modules/Finding

type Finding struct {
	Type        string   `json:"type"`
	Severity    string   `json:"severity"`
	Module      string   `json:"module"`
	Resource    string   `json:"resource"`
	Description string   `json:"description"`
	Remediation string   `json:"remediation"`
	Confidence  *float64 `json:"confidence,omitempty"`
}

type AuditRun struct {
	SchemaVersion string    `json:"schema_version"`
	RunID         string    `json:"run_id"`
	Timestamp     time.Time `json:"timestamp"`
	Host          string    `json:"host"`
	User          string    `json:"user"`
	Mode          string    `json:"mode"`
	Version       string    `json:"version"`
	Findings      []Finding `json:"findings"`
	Score         int       `json:"score"`
	ExitCode      int       `json:"exit_code"`
	DurationMs    int64     `json:"duration_ms"`
}

// T013: summary view type
type SummaryRow struct {
	SK           string
	RunID        string
	Timestamp    string
	Host         string
	User         string
	Score        int
	FindingCount int
	Mode         string
	UserAgent    string
}

// T007: DynamoDB client setup
var (
	dynamoClient *dynamodb.Client
	tableName    string
)

func newDynamoClient() *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		slog.Error("loading AWS config", "err", err)
		os.Exit(1)
	}
	return dynamodb.NewFromConfig(cfg)
}

// T008: dual-mode entry point
func main() {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
	}

	http.HandleFunc("POST /", handlePost)
	http.HandleFunc("GET /event/", handleDetail)
	http.HandleFunc("GET /", handleGet)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		gateway.ListenAndServe("", nil)
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	slog.Info("listening", "addr", "http://localhost:"+port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

// T010–T012: POST / handler

func handlePost(w http.ResponseWriter, r *http.Request) {
	var run AuditRun
	if err := json.NewDecoder(r.Body).Decode(&run); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload: " + err.Error()})
		return
	}
	if run.RunID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload: run_id is required"})
		return
	}

	sk, err := putEvent(r.Context(), run, r.Header.Get("User-Agent"))
	if err != nil {
		slog.Error("putEvent failed", "err", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "storage error"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"run_id": run.RunID, "sk": sk})
}

func putEvent(ctx context.Context, run AuditRun, userAgent string) (string, error) {
	sk := run.Timestamp.UTC().Format(time.RFC3339) + "#" + run.RunID

	findingsJSON, err := json.Marshal(run.Findings)
	if err != nil {
		return "", fmt.Errorf("marshalling findings: %w", err)
	}

	item := map[string]types.AttributeValue{
		"pk":             &types.AttributeValueMemberS{Value: "all"},
		"sk":             &types.AttributeValueMemberS{Value: sk},
		"run_id":         &types.AttributeValueMemberS{Value: run.RunID},
		"schema_version": &types.AttributeValueMemberS{Value: run.SchemaVersion},
		"timestamp":      &types.AttributeValueMemberS{Value: run.Timestamp.UTC().Format(time.RFC3339)},
		"host":           &types.AttributeValueMemberS{Value: run.Host},
		"user":           &types.AttributeValueMemberS{Value: run.User},
		"mode":           &types.AttributeValueMemberS{Value: run.Mode},
		"version":        &types.AttributeValueMemberS{Value: run.Version},
		"score":          &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", run.Score)},
		"exit_code":      &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", run.ExitCode)},
		"duration_ms":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", run.DurationMs)},
		"findings":       &types.AttributeValueMemberS{Value: string(findingsJSON)},
		"finding_count":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", len(run.Findings))},
		"user_agent":     &types.AttributeValueMemberS{Value: userAgent},
	}

	_, err = dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	return sk, err
}

// T014–T016: GET / summary handler

const summaryTmpl = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>SIEM Mock Backend</title>
<style>
body{font-family:monospace;max-width:960px;margin:2rem auto;padding:0 1rem}
h1{margin-bottom:.5rem}
p.count{color:#666;margin-top:0}
table{border-collapse:collapse;width:100%}
th,td{text-align:left;padding:.4rem .75rem;border-bottom:1px solid #ddd}
th{background:#f5f5f5}
tr:hover td{background:#f9f9f9}
a{color:#0070f3}
.score-high{color:#d32f2f;font-weight:bold}
.score-med{color:#f57c00;font-weight:bold}
.score-ok{color:#388e3c;font-weight:bold}
</style>
</head>
<body>
<h1>SIEM Mock Backend</h1>
{{if .Rows}}
<p class="count">{{len .Rows}} most recent submission(s)</p>
<table>
<thead><tr><th>Received</th><th>Host</th><th>User</th><th>Score</th><th>Findings</th><th>Mode</th><th>User-Agent</th></tr></thead>
<tbody>
{{range .Rows}}
<tr>
  <td><a href="/event/{{.SKEncoded}}">{{.Timestamp}}</a></td>
  <td>{{.Host}}</td>
  <td>{{.User}}</td>
  <td class="{{scoreClass .Score}}">{{.Score}}</td>
  <td>{{.FindingCount}}</td>
  <td>{{.Mode}}</td>
  <td>{{.UserAgent}}</td>
</tr>
{{end}}
</tbody>
</table>
{{else}}
<p>No submissions yet. Run <code>ai-check-guardrails</code> with <code>AI_GUARDRAILS_SIEM_ENDPOINT</code> set to post data here.</p>
{{end}}
</body>
</html>`

type summaryTemplateData struct {
	Rows []summaryRowView
}

type summaryRowView struct {
	SummaryRow
	SKEncoded string
}

var summaryTemplate = template.Must(template.New("summary").Funcs(template.FuncMap{
	"scoreClass": func(score int) string {
		switch {
		case score < 50:
			return "score-high"
		case score < 80:
			return "score-med"
		default:
			return "score-ok"
		}
	},
}).Parse(summaryTmpl))

func listEvents(ctx context.Context) ([]SummaryRow, error) {
	out, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("pk = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "all"},
		},
		ScanIndexForward: aws.Bool(false),
		Limit:            aws.Int32(50),
	})
	if err != nil {
		return nil, err
	}

	rows := make([]SummaryRow, 0, len(out.Items))
	for _, item := range out.Items {
		var row SummaryRow
		if v, ok := item["sk"].(*types.AttributeValueMemberS); ok {
			row.SK = v.Value
		}
		if v, ok := item["run_id"].(*types.AttributeValueMemberS); ok {
			row.RunID = v.Value
		}
		if v, ok := item["timestamp"].(*types.AttributeValueMemberS); ok {
			row.Timestamp = v.Value
		}
		if v, ok := item["host"].(*types.AttributeValueMemberS); ok {
			row.Host = v.Value
		}
		if v, ok := item["user"].(*types.AttributeValueMemberS); ok {
			row.User = v.Value
		}
		if v, ok := item["mode"].(*types.AttributeValueMemberS); ok {
			row.Mode = v.Value
		}
		if v, ok := item["user_agent"].(*types.AttributeValueMemberS); ok {
			row.UserAgent = v.Value
		}
		var scoreAttr struct{ N string }
		if v, ok := item["score"].(*types.AttributeValueMemberN); ok {
			scoreAttr.N = v.Value
		}
		fmt.Sscanf(scoreAttr.N, "%d", &row.Score)
		if v, ok := item["finding_count"].(*types.AttributeValueMemberN); ok {
			fmt.Sscanf(v.Value, "%d", &row.FindingCount)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	rows, err := listEvents(r.Context())
	if err != nil {
		slog.Error("listEvents failed", "err", err)
		http.Error(w, "failed to load submissions", http.StatusInternalServerError)
		return
	}

	views := make([]summaryRowView, len(rows))
	for i, row := range rows {
		views[i] = summaryRowView{
			SummaryRow: row,
			SKEncoded:  base64.URLEncoding.EncodeToString([]byte(row.SK)),
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := summaryTemplate.Execute(w, summaryTemplateData{Rows: views}); err != nil {
		slog.Error("template execute failed", "err", err)
	}
}

// T017–T019: GET /event/{sk} detail handler

const detailTmpl = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<title>Run {{.RunID}}</title>
<style>
body{font-family:monospace;max-width:960px;margin:2rem auto;padding:0 1rem}
h1{margin-bottom:.5rem}
dl{display:grid;grid-template-columns:max-content 1fr;gap:.2rem 1rem;margin-bottom:2rem}
dt{font-weight:bold;color:#555}
dd{margin:0}
table{border-collapse:collapse;width:100%}
th,td{text-align:left;padding:.4rem .75rem;border-bottom:1px solid #ddd;vertical-align:top}
th{background:#f5f5f5}
.sev-CRITICAL{color:#d32f2f;font-weight:bold}
.sev-HIGH{color:#e65100;font-weight:bold}
.sev-WARN{color:#f57c00}
.sev-INFO{color:#0277bd}
a{color:#0070f3}
</style>
</head>
<body>
<p><a href="/">← Back to summary</a></p>
<h1>Run {{.RunID}}</h1>
<dl>
  <dt>Timestamp</dt><dd>{{.Timestamp}}</dd>
  <dt>Host</dt><dd>{{.Host}}</dd>
  <dt>User</dt><dd>{{.User}}</dd>
  <dt>Mode</dt><dd>{{.Mode}}</dd>
  <dt>Version</dt><dd>{{.Version}}</dd>
  <dt>Score</dt><dd>{{.Score}}</dd>
  <dt>Exit Code</dt><dd>{{.ExitCode}}</dd>
  <dt>Duration</dt><dd>{{.DurationMs}}ms</dd>
  <dt>Schema</dt><dd>{{.SchemaVersion}}</dd>
  <dt>User-Agent</dt><dd>{{.UserAgent}}</dd>
</dl>
{{if .Findings}}
<h2>Findings ({{len .Findings}})</h2>
<table>
<thead><tr><th>Module</th><th>Severity</th><th>Type</th><th>Resource</th><th>Description</th><th>Remediation</th><th>Confidence</th></tr></thead>
<tbody>
{{range .Findings}}
<tr>
  <td>{{.Module}}</td>
  <td class="sev-{{.Severity}}">{{.Severity}}</td>
  <td>{{.Type}}</td>
  <td>{{.Resource}}</td>
  <td>{{.Description}}</td>
  <td>{{.Remediation}}</td>
  <td>{{if .Confidence}}{{printf "%.0f%%" (mul .Confidence 100)}}{{else}}—{{end}}</td>
</tr>
{{end}}
</tbody>
</table>
{{else}}
<p>No findings — clean run.</p>
{{end}}
</body>
</html>`

type detailTemplateData struct {
	AuditRun
	UserAgent string
}

var detailTemplate = template.Must(template.New("detail").Funcs(template.FuncMap{
	"mul": func(p *float64, n float64) float64 {
		if p == nil {
			return 0
		}
		return *p * n
	},
}).Parse(detailTmpl))

func getEvent(ctx context.Context, sk string) (*AuditRun, string, error) {
	out, err := dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "all"},
			"sk": &types.AttributeValueMemberS{Value: sk},
		},
	})
	if err != nil {
		return nil, "", err
	}
	if out.Item == nil {
		return nil, "", nil
	}

	item := out.Item
	run := &AuditRun{}

	strAttr := func(key string) string {
		if v, ok := item[key].(*types.AttributeValueMemberS); ok {
			return v.Value
		}
		return ""
	}
	numAttr := func(key string) string {
		if v, ok := item[key].(*types.AttributeValueMemberN); ok {
			return v.Value
		}
		return "0"
	}

	run.RunID = strAttr("run_id")
	run.SchemaVersion = strAttr("schema_version")
	run.Host = strAttr("host")
	run.User = strAttr("user")
	run.Mode = strAttr("mode")
	run.Version = strAttr("version")

	if ts := strAttr("timestamp"); ts != "" {
		run.Timestamp, _ = time.Parse(time.RFC3339, ts)
	}

	fmt.Sscanf(numAttr("score"), "%d", &run.Score)
	fmt.Sscanf(numAttr("exit_code"), "%d", &run.ExitCode)
	fmt.Sscanf(numAttr("duration_ms"), "%d", &run.DurationMs)

	if findingsRaw := strAttr("findings"); findingsRaw != "" {
		if err := json.Unmarshal([]byte(findingsRaw), &run.Findings); err != nil {
			return nil, "", fmt.Errorf("unmarshalling findings: %w", err)
		}
	}

	return run, strAttr("user_agent"), nil
}

func handleDetail(w http.ResponseWriter, r *http.Request) {
	encoded := strings.TrimPrefix(r.URL.Path, "/event/")
	if encoded == "" {
		http.Error(w, "missing event id", http.StatusBadRequest)
		return
	}

	skBytes, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		http.Error(w, "invalid event id", http.StatusBadRequest)
		return
	}

	run, userAgent, err := getEvent(r.Context(), string(skBytes))
	if err != nil {
		slog.Error("getEvent failed", "err", err)
		http.Error(w, "failed to load event", http.StatusInternalServerError)
		return
	}
	if run == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := detailTemplate.Execute(w, detailTemplateData{AuditRun: *run, UserAgent: userAgent}); err != nil {
		slog.Error("template execute failed", "err", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
