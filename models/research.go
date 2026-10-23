package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// ResearchSchema identifies the versioned result_json contract for
// transcription, frame analysis, and embedding model receipts.
const ResearchSchema = "research.v1"

const MaxResearchResultBytes = 32 << 20

type TimedText struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

type FrameAnalysis struct {
	TimestampMS int64    `json:"timestamp_ms"`
	Index       int      `json:"index"`
	OCR         []string `json:"ocr"`
	Description string   `json:"description"`
	Coverage    string   `json:"coverage"`
}

type EmbeddingVector struct {
	Index  int       `json:"index"`
	Values []float32 `json:"values"`
}

// ResearchResult is the read-only platform result_json shape. Applications
// still verify kind, model version, dimensions, and their own frozen inputs
// before using a receipt.
type ResearchResult struct {
	Coverage        string            `json:"coverage,omitempty"`
	Schema          string            `json:"schema"`
	Kind            string            `json:"kind"`
	Model           string            `json:"model"`
	ModelVersion    string            `json:"model_version"`
	ResponseModel   string            `json:"response_model,omitempty"`
	Text            string            `json:"text,omitempty"`
	DurationSeconds float64           `json:"duration_seconds,omitempty"`
	Segments        []TimedText       `json:"segments,omitempty"`
	Words           []TimedText       `json:"words,omitempty"`
	Frames          []FrameAnalysis   `json:"frames,omitempty"`
	Dimensions      int               `json:"dimensions,omitempty"`
	Encoding        string            `json:"encoding,omitempty"`
	Vectors         []EmbeddingVector `json:"vectors,omitempty"`
}

// ParseResearchResult reads one bounded result_json document and ignores
// unknown fields so future v1 additions remain compatible with current clients.
func ParseResearchResult(raw []byte) (ResearchResult, error) {
	if len(raw) == 0 || len(raw) > MaxResearchResultBytes {
		return ResearchResult{}, fmt.Errorf("research result size is invalid")
	}
	var result ResearchResult
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&result); err != nil {
		return ResearchResult{}, fmt.Errorf("decode research result: %w", err)
	}
	if result.Schema != ResearchSchema {
		return ResearchResult{}, fmt.Errorf("unsupported research result schema %q", result.Schema)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return ResearchResult{}, fmt.Errorf("trailing research result data")
	}
	return result, nil
}
