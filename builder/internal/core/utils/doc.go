// Package utils contains small, dependency-free helper functions used by the
// builder core. It provides deterministic UUID generation for nodes, safe
// parsing helpers for nullable CSV/database fields (string, int32, int64),
// and utilities for mapping raw log values into domain types (for example,
// converting numeric node-kind codes to domain.NodeKind). Functions in this
// package are intentionally minimal and well-tested so they can be reused
// across parsing and import code without side effects.
package utils
