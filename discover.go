package afspec

// SpecMeta is a lightweight struct containing metadata parsed from a spec's
// PRD frontmatter, used for discovery and landscape views.
type SpecMeta struct {
	SpecID   string
	SpecName string
	Status   string
	Dir      string
}

// DiscoverSpecs scans the root directory for subdirectories matching the
// NN_snake_case pattern, parses the PRD frontmatter of each, and returns
// a slice of SpecMeta entries with SpecID, SpecName, Status, and Dir
// fields populated.
//
// If a subdirectory matches the pattern but its prd.md is missing or
// malformed, the entry is skipped and the error is collected. If the
// root directory does not exist or is not readable, a LoadError is returned.
func DiscoverSpecs(root string) ([]SpecMeta, error) {
	panic("not implemented")
}

// LoadSpecLandscape collects metadata for all specs for landscape views.
// It scans spec directories, optionally including the archive subdirectory,
// and excludes the currentSpecID entry from results.
//
// When includeArchive is true, archived specs from {root}/archive/ are
// included. When false, only non-archived spec directories are returned.
func LoadSpecLandscape(root string, includeArchive bool, currentSpecID string) ([]SpecMeta, error) {
	panic("not implemented")
}
