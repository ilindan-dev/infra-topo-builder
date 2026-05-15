// Package parser provides parsers for ibdiagnet diagnostic outputs and related
// enrichment files (for example, .sharp_an_info). The package converts raw
// textual diagnostics into domain models (nodes, ports, switches, node metadata)
// and persists them using ports.ParserWriter. Parsers are optimized for
// streaming large files with configurable batching to minimize memory usage and
// maximize DB throughput.
package parser
