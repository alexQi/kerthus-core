package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestResearchResultWireShapes(t *testing.T) {
	for _, raw := range []string{
		`{"schema":"research.v1","kind":"transcription","model":"asr","model_version":"v1","response_model":"asr-snapshot","coverage":"transcribed","text":"hello","duration_seconds":1.5,"segments":[{"start":0,"end":1.5,"text":"hello"}],"words":[{"start":0,"end":1.5,"text":"hello"}]}`,
		`{"schema":"research.v1","kind":"analysis","model":"vision","model_version":"v1","frames":[{"timestamp_ms":1250,"index":0,"ocr":["text"],"description":"scene","coverage":"observed"}]}`,
		`{"schema":"research.v1","kind":"embedding","model":"embed","model_version":"v1","response_model":"snapshot","dimensions":2,"encoding":"float32","vectors":[{"index":0,"values":[0.25,0.5]}],"future_field":true}`,
	} {
		result, err := ParseResearchResult([]byte(raw))
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(result)
		if err != nil {
			t.Fatal(err)
		}
		var decoded ResearchResult
		if err := json.Unmarshal(encoded, &decoded); err != nil || decoded.Schema != ResearchSchema || decoded.Kind != result.Kind || decoded.ModelVersion != result.ModelVersion || decoded.ResponseModel != result.ResponseModel {
			t.Fatalf("research result did not round trip: %#v, %v", decoded, err)
		}
	}
}

func TestParseResearchResultRejectsInvalidDocument(t *testing.T) {
	for _, raw := range []string{"", "null", `{"schema":"research.v2"}`, `{"schema":"research.v1"}{}`, strings.Repeat(" ", MaxResearchResultBytes+1)} {
		if _, err := ParseResearchResult([]byte(raw)); err == nil {
			t.Fatalf("invalid research result accepted: %q", raw[:min(len(raw), 80)])
		}
	}
}
