// Package migrations contains SQL migration scripts applied to the Postgres
// database for the infra-topo-builder. Files follow the conventional naming
// scheme: <version>_description.up.sql and <version>_description.down.sql.
//
// Apply migrations with your preferred tool (for example: golang-migrate),
// ensuring up scripts are run in ascending order. The migrations are the
// source of truth for schema changes and should be reviewed in code reviews.
package migrations
