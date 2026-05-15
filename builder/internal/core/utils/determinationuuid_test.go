// Tests for functions in the utils package focused on deterministic UUID generation.
// Ensures identical inputs produce identical UUIDs and different inputs produce different UUIDs.
package utils

import (
	"testing"

	"github.com/google/uuid"
)

// TestGenerateDeterministicUUID checks that generateDeterministicUUID is stable
// for the same inputs and differs for different logIDs or node GUIDs.
func TestGenerateDeterministicUUID(t *testing.T) {
	logID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	guid1 := "0xswitch1"
	guid2 := "0xswitch2"

	uuid1 := GenerateDeterministicUUID(logID, guid1)
	uuid2 := GenerateDeterministicUUID(logID, guid1)
	if uuid1 != uuid2 {
		t.Errorf("Expected identical UUIDs for same inputs, got %s and %s", uuid1, uuid2)
	}

	uuid3 := GenerateDeterministicUUID(logID, guid2)
	if uuid1 == uuid3 {
		t.Errorf("Expected different UUIDs for different GUIDs, got identical %s", uuid1)
	}

	diffLogID := uuid.MustParse("987fcdeb-51a2-43d7-9012-345678901234")
	uuid4 := GenerateDeterministicUUID(diffLogID, guid1)
	if uuid1 == uuid4 {
		t.Errorf("Expected different UUIDs for different LogIDs, got identical %s", uuid1)
	}
}
