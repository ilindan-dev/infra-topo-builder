package utils

import "github.com/google/uuid"

// GenerateDeterministicUUID returns a name-based UUID (version 5) derived from
// the provided logID (used as the namespace) and nodeGUID (used as the name).
// This produces stable, deterministic UUIDs for the same (logID, nodeGUID) pair,
// enabling consistent node identification across parsing runs without keeping an in-memory mapping.
func GenerateDeterministicUUID(logID uuid.UUID, nodeGUID string) uuid.UUID {
	return uuid.NewSHA1(logID, []byte(nodeGUID))
}
