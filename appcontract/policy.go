package appcontract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
)

const MaxToolOutputBytes = 64 << 10
const ServicePrincipalPlaceholder = "${SERVICE_PRINCIPAL_ID}"

// Policies are opt-in. Only these versioned capabilities have runtime implementations.
type Policy struct {
	Tool         *Tool         `json:"tool,omitempty" yaml:"tool,omitempty"`
	Transport    *Transport    `json:"transport,omitempty" yaml:"transport,omitempty"`
	Capabilities *Capabilities `json:"capabilities,omitempty" yaml:"capabilities,omitempty"`
}
type Tool struct {
	Version        int            `json:"version" yaml:"version"`
	Name           string         `json:"name" yaml:"name"`
	Description    string         `json:"description" yaml:"description"`
	Risk           string         `json:"risk" yaml:"risk"`
	Approval       bool           `json:"approval" yaml:"approval"`
	Schema         map[string]any `json:"schema" yaml:"schema"`
	TimeoutSeconds int            `json:"timeout_seconds" yaml:"timeout_seconds"`
	Idempotent     bool           `json:"idempotent" yaml:"idempotent"`
	OutputBytes    int            `json:"output_bytes" yaml:"output_bytes"`
}
type Transport struct {
	Mode             string `json:"mode" yaml:"mode"`
	BodyBytes        int64  `json:"body_bytes" yaml:"body_bytes"`
	FirstByteSeconds int    `json:"first_byte_seconds" yaml:"first_byte_seconds"`
	IdleSeconds      int    `json:"idle_seconds" yaml:"idle_seconds"`
	TotalSeconds     int    `json:"total_seconds" yaml:"total_seconds"`
}
type Capabilities struct {
	BackgroundJobs     bool     `json:"background_jobs,omitempty" yaml:"background_jobs,omitempty"`
	ServicePrincipals  []string `json:"service_principals,omitempty" yaml:"service_principals,omitempty"`
	ModelReceipts      bool     `json:"model_receipts,omitempty" yaml:"model_receipts,omitempty"`
	JobInputs          bool     `json:"job_inputs,omitempty" yaml:"job_inputs,omitempty"`
	JobSourcePromotion bool     `json:"job_source_promotion,omitempty" yaml:"job_source_promotion,omitempty"`
	MediaReferences    bool     `json:"media_references" yaml:"media_references"`
	MediaAccess        bool     `json:"media_access" yaml:"media_access"`
	Jobs               bool     `json:"jobs" yaml:"jobs"`
	Media              bool     `json:"media" yaml:"media"`
	Models             []string `json:"models" yaml:"models"`
}

func ParsePolicy(raw string) (Policy, error) {
	var p Policy
	if raw == "" {
		return p, nil
	}
	d := json.NewDecoder(bytes.NewBufferString(raw))
	d.DisallowUnknownFields()
	if e := d.Decode(&p); e != nil {
		return p, e
	}
	if e := d.Decode(new(any)); e != io.EOF {
		return p, fmt.Errorf("trailing policy data")
	}
	return p, p.Validate()
}

// PolicyDigest canonicalizes a validated operation policy so resource-version
// changes can be accepted only when the authorization contract is unchanged.
func PolicyDigest(raw string) (string, error) {
	p, err := ParsePolicy(raw)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return Digest(b), nil
}
func (p Policy) Validate() error {
	if c := p.Capabilities; c != nil {
		if c.BackgroundJobs && (!c.Jobs || !c.Media || !c.MediaReferences || len(c.ServicePrincipals) == 0) {
			return fmt.Errorf("background_jobs requires jobs, media, media_references and registered service principals")
		}
		if c.ModelReceipts && (!c.Jobs || len(c.Models) == 0) {
			return fmt.Errorf("model_receipts requires jobs and model capabilities")
		}
		if c.JobInputs && (!c.Jobs || !c.Media || !c.MediaReferences) {
			return fmt.Errorf("job inputs require jobs, media and media_references")
		}
		if c.JobSourcePromotion && (!c.Jobs || !c.Media || !c.MediaReferences || !c.JobInputs) {
			return fmt.Errorf("job source promotion requires job_inputs, jobs, media and media_references")
		}
		if (c.MediaReferences || c.MediaAccess) && !c.Media {
			return fmt.Errorf("media delegation requires media capability")
		}
		if len(c.Models) > 0 && !c.Jobs {
			return fmt.Errorf("model calls require job authorization capability")
		}
		seen := map[string]bool{}
		for _, v := range c.Models {
			if (v != "text" && v != "image" && v != "video" && v != "transcription" && v != "analysis" && v != "embedding") || seen[v] {
				return fmt.Errorf("unsupported/duplicate model capability")
			}
			if v != "text" && v != "embedding" && !c.Media {
				return fmt.Errorf("media models require private media capability")
			}
			seen[v] = true
		}
	}
	if t := p.Transport; t != nil {
		if t.Mode != "http" && t.Mode != "sse" && t.Mode != "upload" {
			return fmt.Errorf("unsupported transport mode")
		}
		if t.BodyBytes < 1 || t.BodyBytes > 64<<20 || t.FirstByteSeconds < 1 || t.FirstByteSeconds > 30 || t.IdleSeconds < 1 || t.IdleSeconds > 60 || t.TotalSeconds < 1 || t.TotalSeconds > 3600 {
			return fmt.Errorf("transport bounds invalid")
		}
		if t.Mode == "http" && (t.TotalSeconds > 120 || t.BodyBytes > 12<<20) {
			return fmt.Errorf("ordinary HTTP bounds exceeded")
		}
		if t.Mode == "sse" && (t.BodyBytes > 1<<20 || t.TotalSeconds < t.IdleSeconds) {
			return fmt.Errorf("stream bounds invalid")
		}
	}
	if t := p.Tool; t != nil {
		if t.Version < 1 || t.Name == "" || len(t.Name) > 100 || len(t.Description) > 1000 || t.TimeoutSeconds < 1 || t.TimeoutSeconds > 30 || t.OutputBytes < 1 || t.OutputBytes > MaxToolOutputBytes {
			return fmt.Errorf("tool bounds invalid")
		}
		switch t.Risk {
		case "read", "write", "delete", "permission", "export":
		default:
			return fmt.Errorf("unsupported tool risk")
		}
		if !t.Idempotent {
			return fmt.Errorf("application tools require durable application idempotency")
		}
		if t.Schema["type"] != "object" {
			return fmt.Errorf("tool schema must be an object")
		}
		if e := ValidateSchema(t.Schema, 0); e != nil {
			return e
		}
		if p.Transport != nil && p.Transport.Mode != "http" {
			return fmt.Errorf("tools must return promptly over ordinary HTTP")
		}
	}
	return nil
}
func (p Policy) JSON() string { b, _ := json.Marshal(p); return string(b) }
func (p Policy) Route() Transport {
	if p.Transport != nil {
		return *p.Transport
	}
	return Transport{Mode: "http", BodyBytes: 12 << 20, FirstByteSeconds: 10, IdleSeconds: 30, TotalSeconds: 60}
}

var toolID = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func ToolName(app, operation, policy, service string, version int64) string {
	prefix := toolID.ReplaceAllString(app+"_"+operation, "_")
	if len(prefix) > 38 {
		prefix = prefix[:38]
	}
	b, _ := json.Marshal([]any{app, operation, policy, service, version})
	return "app_" + prefix + "_" + Digest(b)[:20]
}
