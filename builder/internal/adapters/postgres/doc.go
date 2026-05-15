// Package postgres provides a Postgres adapter implementing the core persistence ports.
// It coordinates schema migrations, uses sqlc-generated code in the gen package
// for type-safe SQL access, and exposes CQRS-friendly interfaces implemented in
// parserwriter.go, topologyreader.go and logmanager.go. The adapter prefers bulk
// operations (pgx Batch/COPY) for import performance and translates driver errors
// into canonical domain errors using translateError.
//
// Directory layout:
//   - gen/: sqlc-generated Go code (DO NOT EDIT)
//   - migrations/: SQL migration scripts (apply with your migration tool)
//   - queries/: SQL source files used by sqlc to generate gen/
package postgres
