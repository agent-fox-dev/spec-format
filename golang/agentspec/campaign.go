package agentspec

import "time"

// CampaignMetadata holds the metadata stored in campaign.yaml.
type CampaignMetadata struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	CreatedAt   time.Time `yaml:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at"`
}

// Campaign represents a campaign directory with its metadata.
type Campaign struct {
	Path     string
	Metadata CampaignMetadata
}

// CreateCampaign initializes a new campaign directory by writing
// campaign.yaml atomically using the atomic temp-file-and-rename pattern.
// Returns a CampaignError if the parent directory does not exist,
// campaign.yaml already exists, or name/description is empty.
func CreateCampaign(path, name, description string) (*Campaign, error) {
	// TODO: implement
	return nil, nil
}

// OpenCampaign reads and parses an existing campaign.yaml from path
// to reconstruct a Campaign. Returns a CampaignError if the file is
// missing, malformed, or missing required fields.
func OpenCampaign(path string) (*Campaign, error) {
	// TODO: implement
	return nil, nil
}

// Specs returns the names of all subdirectories in the campaign directory
// that match the {NN}_{snake_case} spec directory naming pattern, sorted
// by numeric prefix ascending. The archive/ subdirectory is excluded.
func (c *Campaign) Specs() ([]string, error) {
	// TODO: implement
	return nil, nil
}

// NewSpec provisions a new spec directory within the campaign from a PRD
// file. It validates the spec name, computes the next numeric prefix,
// creates the directory, writes prd.md with YAML frontmatter, and
// initializes a SpecSession via CreateSession.
func (c *Campaign) NewSpec(specName, prdPath, mode, source string) (*SpecSession, error) {
	// TODO: implement
	return nil, nil
}
