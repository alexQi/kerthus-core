// Package appcontract contains transport-neutral public application contracts.
package appcontract

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
)

const TicketHeader = "X-Kerthus-Delegation"
const RequestIDHeader = "X-Kerthus-Request-ID"
const MaxRequestBytes int64 = 64 << 20

func Digest(b []byte) string { v := sha256.Sum256(b); return hex.EncodeToString(v[:]) }

// Fingerprint binds exact wire bytes; query ordering and JSON whitespace are intentional.
func Fingerprint(r *http.Request, body []byte) string {
	b, _ := json.Marshal([]string{r.Method, r.URL.EscapedPath(), r.URL.RawQuery, r.Header.Get("Content-Type"), r.Header.Get("Idempotency-Key"), r.Header.Get("Range"), r.Header.Get("If-Range"), r.Header.Get("Last-Event-ID"), Digest(body)})
	return Digest(b)
}
