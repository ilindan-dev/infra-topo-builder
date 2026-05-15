// Tests for parseNodeKind which maps numeric node-kind codes from logs
// into the domain.NodeKind enumeration.
package utils

import (
	"testing"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
)

// TestParseNodeKind verifies valid codes map to expected NodeKind values
// and that invalid codes return an error and NodeInvalidKind.
func TestParseNodeKind(t *testing.T) {
	tests := []struct {
		name    string
		input   int32
		want    domain.NodeKind
		wantErr bool
	}{
		{"Host", 1, domain.NodeHost, false},
		{"Switch", 2, domain.NodeSwitch, false},
		{"Unknown", 99, domain.NodeInvalidKind, true},
		{"Zero", 0, domain.NodeInvalidKind, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNodeKind(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNodeKind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseNodeKind() got = %v, want %v", got, tt.want)
			}
		})
	}
}
