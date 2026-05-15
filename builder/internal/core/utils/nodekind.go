package utils

import (
	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
)

// ParseNodeKind maps an integer code from CSV logs into a domain.NodeKind.
// According to ibdiagnet: 1 -> Host, 2 -> Switch. Returns NodeInvalidKind and
// domain.ErrInvalidNodeKind for any unsupported or unknown code.
func ParseNodeKind(kind int32) (domain.NodeKind, error) {
	switch kind {
	case 1:
		return domain.NodeHost, nil
	case 2:
		return domain.NodeSwitch, nil
	default:
		return domain.NodeInvalidKind, domain.ErrInvalidNodeKind
	}
}
