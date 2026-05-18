package audit

import "time"

type SeverityASFF struct {
	Label    string `json:"Label,omitempty"`
	Original string `json:"Original,omitempty"`
}

type ResourceASFF struct {
	ID        string `json:"Id"`
	Type      string `json:"Type"`
	Region    string `json:"Region,omitempty"`
	Partition string `json:"Partition,omitempty"`
}

type Finding struct {
	Type        string   `json:"type"`
	Severity    string   `json:"severity"`
	Module      string   `json:"module"`
	Resource    string   `json:"resource"`
	Description string   `json:"description"`
	Remediation string   `json:"remediation"`
	Confidence  *float64 `json:"confidence,omitempty"`
	// ASFF fields — all omitempty for backwards compatibility
	ID           string         `json:"Id,omitempty"`
	Title        string         `json:"title,omitempty"`
	GeneratorId  string         `json:"generator_id,omitempty"`
	ASFFTypes    []string       `json:"asff_types,omitempty"`
	ASFFSeverity *SeverityASFF  `json:"asff_severity,omitempty"`
	Resources    []ResourceASFF `json:"resources,omitempty"`
	CreatedAt    string         `json:"created_at,omitempty"`
	UpdatedAt    string         `json:"updated_at,omitempty"`
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
	// ASFF run-level identity fields — omitempty for backwards compatibility
	ProductArn   string `json:"product_arn,omitempty"`
	AwsAccountId string `json:"aws_account_id,omitempty"`
}
