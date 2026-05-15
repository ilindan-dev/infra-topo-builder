package domain

import "github.com/google/uuid"

// Topology represents the complete parsed network graph for a specific parsing session.
type Topology struct {
	LogID uuid.UUID
	Nodes []Node
}
