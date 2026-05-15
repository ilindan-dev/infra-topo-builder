// Package domain defines the core business entities and value objects
// used by the infra-topo-builder. This package contains pure domain
// models, canonical errors, and small aggregates that represent the
// parsed topology and its elements (logs, nodes, ports, switches, and
// node metadata). It intentionally has no dependencies on persistence,
// transport, or CLI layers so it can be reused across application
// components and tested in isolation.
//
// Typical usage:
//   - A parser translates raw log lines into domain.Node, domain.Port, and
//     domain.Switch instances and groups them into a domain.Topology.
//   - Business services operate on these models and map them to persistence
//     adapters or API DTOs.
package domain
