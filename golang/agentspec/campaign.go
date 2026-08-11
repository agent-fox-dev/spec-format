package agentspec

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-yaml"
)

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
	// Validate name and description are non-empty.
	if name == "" {
		return nil, &CampaignError{Msg: "campaign name must not be empty"}
	}
	if description == "" {
		return nil, &CampaignError{Msg: "campaign description must not be empty"}
	}

	// Validate parent directory exists.
	parentDir := filepath.Dir(path)
	if _, err := os.Stat(parentDir); err != nil {
		return nil, &CampaignError{
			Msg:   fmt.Sprintf("parent directory does not exist: %s", parentDir),
			Cause: err,
		}
	}

	// Check that campaign.yaml does not already exist at path.
	yamlPath := filepath.Join(path, "campaign.yaml")
	if _, err := os.Stat(yamlPath); err == nil {
		return nil, &CampaignError{
			Msg: fmt.Sprintf("campaign.yaml already exists at %s", path),
		}
	}

	// Create the campaign directory if it does not exist.
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, &CampaignError{
			Msg:   fmt.Sprintf("failed to create campaign directory %s: %v", path, err),
			Cause: err,
		}
	}

	now := time.Now().UTC()
	meta := CampaignMetadata{
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Marshal metadata to YAML.
	data, err := yaml.Marshal(meta)
	if err != nil {
		return nil, &CampaignError{
			Msg:   fmt.Sprintf("failed to marshal campaign metadata: %v", err),
			Cause: err,
		}
	}

	// Atomic write: write to temp file in same directory, then rename.
	tmpFile, err := os.CreateTemp(path, "campaign.yaml.tmp.*")
	if err != nil {
		return nil, &CampaignError{
			Msg:   fmt.Sprintf("failed to create temporary file for campaign.yaml: %v", err),
			Cause: err,
		}
	}
	tmpPath := tmpFile.Name()

	// Ensure cleanup of temp file on any failure.
	success := false
	defer func() {
		if !success {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return nil, &CampaignError{
			Msg:   fmt.Sprintf("failed to write campaign.yaml temp file: %v", err),
			Cause: err,
		}
	}
	if err := tmpFile.Close(); err != nil {
		return nil, &CampaignError{
			Msg:   fmt.Sprintf("failed to close campaign.yaml temp file: %v", err),
			Cause: err,
		}
	}

	// Rename temp file to final path atomically.
	if err := os.Rename(tmpPath, yamlPath); err != nil {
		return nil, &CampaignError{
			Msg:   fmt.Sprintf("failed to rename temp file to campaign.yaml: %v", err),
			Cause: err,
		}
	}

	success = true

	return &Campaign{
		Path:     path,
		Metadata: meta,
	}, nil
}

// OpenCampaign reads and parses an existing campaign.yaml from path
// to reconstruct a Campaign. Returns a CampaignError if the file is
// missing, malformed, or missing required fields.
func OpenCampaign(path string) (*Campaign, error) {
	// Validate path is non-empty.
	if path == "" {
		return nil, &CampaignError{Msg: "campaign path must not be empty"}
	}

	yamlPath := filepath.Join(path, "campaign.yaml")

	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, &CampaignError{
			Msg:   fmt.Sprintf("failed to read campaign.yaml at %s: %v", path, err),
			Cause: err,
		}
	}

	var meta CampaignMetadata
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, &CampaignError{
			Msg:   fmt.Sprintf("failed to parse campaign.yaml at %s: %v", path, err),
			Cause: err,
		}
	}

	// Validate required fields.
	if meta.Name == "" {
		return nil, &CampaignError{
			Msg: fmt.Sprintf("campaign.yaml at %s is missing required field: name", path),
		}
	}
	if meta.CreatedAt.IsZero() {
		return nil, &CampaignError{
			Msg: fmt.Sprintf("campaign.yaml at %s is missing required field: created_at", path),
		}
	}

	return &Campaign{
		Path:     path,
		Metadata: meta,
	}, nil
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
