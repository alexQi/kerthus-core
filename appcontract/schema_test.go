package appcontract

import (
	"encoding/json"
	"testing"
)

func TestStrictSchemaAndValues(t *testing.T) {
	var s map[string]any
	json.Unmarshal([]byte(`{"type":"object","additionalProperties":false,"required":["count"],"properties":{"count":{"type":"integer","minimum":1,"maximum":10}}}`), &s)
	if e := ValidateSchema(s, 0); e != nil {
		t.Fatal(e)
	}
	for _, value := range []string{`{}`, `{"count":0}`, `{"count":1.5}`, `{"count":2,"tenant_id":99}`, `{"count":"2"}`} {
		var v any
		json.Unmarshal([]byte(value), &v)
		if e := ValidateValue(s, v); e == nil {
			t.Fatal("invalid value accepted", value)
		}
	}
	for key, value := range map[string]any{"$ref": "http://example.com/schema", "patternProperties": map[string]any{}, "unevaluatedProperties": false} {
		clone := map[string]any{}
		for k, v := range s {
			clone[k] = v
		}
		clone[key] = value
		if e := ValidateSchema(clone, 0); e == nil {
			t.Fatal("unsupported keyword ignored", key)
		}
	}
	if _, e := ParsePolicy(`{"capabilities":{"provisioning":true}}`); e == nil {
		t.Fatal("unimplemented capability accepted")
	}
	if _, e := ParsePolicy(`{"transport":{"mode":"sse","body_bytes":1048576,"first_byte_seconds":10,"idle_seconds":30,"total_seconds":0}}`); e == nil {
		t.Fatal("unlimited timeout accepted")
	}
}

func TestToolResultLimitMatchesRuntimeContract(t *testing.T) {
	p := Policy{Tool: &Tool{Version: 1, Name: "read", Risk: "read", TimeoutSeconds: 5, Idempotent: true, OutputBytes: MaxToolOutputBytes, Schema: map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{}}}}
	if e := p.Validate(); e != nil {
		t.Fatal(e)
	}
	p.Tool.OutputBytes++
	if e := p.Validate(); e == nil {
		t.Fatal("manifest advertises more bytes than runtime can accept")
	}
}
