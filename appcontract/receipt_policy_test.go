package appcontract

import "testing"

func TestHistoricalReceiptPolicyOptIn(t *testing.T) {
	for _, raw := range []string{`{"capabilities":{"model_receipts":true}}`, `{"capabilities":{"model_receipts":true,"jobs":true}}`, `{"capabilities":{"model_receipts":true,"jobs":true,"models":["not-a-model"]}}`, `{"capabilities":{"model_receipts":true,"unknown_receipt_option":true}}`} {
		if _, err := ParsePolicy(raw); err == nil {
			t.Errorf("invalid receipt policy accepted: %s", raw)
		}
	}
	p, err := ParsePolicy(`{"capabilities":{"model_receipts":true,"jobs":true,"models":["text"]}}`)
	if err != nil || !p.Capabilities.ModelReceipts {
		t.Fatalf("valid policy rejected: %+v %v", p, err)
	}
	p, err = ParsePolicy(`{"capabilities":{"jobs":true,"models":["text"]}}`)
	if err != nil || p.Capabilities.ModelReceipts {
		t.Fatal("historical receipt reads must opt in")
	}
}
