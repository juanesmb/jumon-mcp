package observability

import (
	"crypto/sha256"
	"encoding/hex"
)

// UserHashHex returns a deterministic, short hex digest suitable for correlating telemetry without emitting raw Clerk user identifiers.
func UserHashHex(userID, salt string) string {
	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte("|"))
	h.Write([]byte(userID))
	sum := h.Sum(nil)
	return hex.EncodeToString(sum[:8])
}
