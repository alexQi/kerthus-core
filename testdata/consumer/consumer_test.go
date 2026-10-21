package consumer

import "testing"

func TestPublicModuleConsumerCompiles(t *testing.T) {
	if err := Compile(nil, ""); err == nil {
		t.Fatal("nil platform unexpectedly accepted")
	}
}
