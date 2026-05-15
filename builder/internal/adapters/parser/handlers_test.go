package parser

import (
	"testing"

	"github.com/google/uuid"

	"github.com/ilindan-dev/infra-topo-builder/builder/internal/core/domain"
)

func TestCleanGUID(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{" 00112233aabbCC ", "0x00112233aabbcc"},
		{"0xABC123", "0xabc123"},
		{"", ""},
		{"   ", ""},
	}

	for _, tt := range tests {
		got := cleanGUID(tt.input)
		if got != tt.expected {
			t.Errorf("cleanGUID(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseNodeLine(t *testing.T) {
	logID := uuid.New()
	// desc, num_ports, type, class_ver, base_ver, sys_guid, node_guid, port_guid
	line := `"Switch-1",36,2,10,20,001122,AABBCC,DDEEFF`

	node, err := parseNodeLine(logID, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if node.NodeDesc != "Switch-1" {
		t.Errorf("expected Switch-1, got %s", node.NodeDesc)
	}
	if node.NodeGUID != "0xaabbcc" {
		t.Errorf("expected 0xaabbcc, got %s", node.NodeGUID)
	}
	if node.NodeType != domain.NodeSwitch {
		t.Errorf("expected Switch kind, got %v", node.NodeType)
	}

	if _, err := parseNodeLine(logID, "invalid,line"); err == nil {
		t.Error("expected error for insufficient columns")
	}
}

func TestParseSystemInfoLine(t *testing.T) {
	logID := uuid.New()
	// guid, serial, part, rev, name
	line := `AABBCC,SN123,PN456,R1,Quantum-Switch`

	info, err := parseSystemInfoLine(logID, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.SerialNumber == nil || *info.SerialNumber != "SN123" {
		t.Errorf("expected SN123, got %v", info.SerialNumber)
	}

	if info.ProductName == nil || *info.ProductName != "Quantum-Switch" {
		t.Errorf("expected Quantum-Switch, got %v", info.ProductName)
	}

	if info.NodeID == uuid.Nil {
		t.Error("expected non-nil NodeID")
	}
}

func TestParsePortLine_Invalid(t *testing.T) {
	logID := uuid.New()
	_, err := parsePortLine(logID, "node,port,1,2,3")
	if err == nil {
		t.Error("expected error for short port line, got nil")
	}
}
