package afspec

import "embed"

//go:embed schemas/*.json
var schemaFS embed.FS

// Schemas returns the bundled JSON Schema files as a map of schema file names
// to their raw bytes. The schemas are embedded at compile time from the
// schemas/ directory via //go:embed. The returned map is never nil and is
// never empty because schema files are embedded at compile time.
func Schemas() map[string][]byte {
	panic("not implemented")
}
