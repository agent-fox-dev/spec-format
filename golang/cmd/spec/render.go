package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// renderArtifactNames maps artifact basenames to their display key names.
var renderArtifactNames = map[string]string{
	"requirements.json": "requirements",
	"test_spec.json":    "test_spec",
	"tasks.json":        "tasks",
}

// renderArtifactOrder is the canonical order for rendering artifacts.
var renderArtifactOrder = []string{"requirements.json", "test_spec.json", "tasks.json"}

// newRenderCmd creates the "spec render" subcommand which renders spec
// artifacts as markdown or JSON.
func newRenderCmd() *cobra.Command {
	var jsonOutput bool
	var combined bool

	cmd := &cobra.Command{
		Use:   "render SPEC",
		Short: "Render spec artifacts as markdown or JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specDir, _ := cmd.Flags().GetString("spec-dir")
			specName := args[0]

			// Auto-enable --json when AF_AGENT=1 is active.
			if isAgentMode() {
				jsonOutput = true
			}

			// Resolve spec to its directory path.
			specPath, err := resolveSpec(specDir, specName)
			if err != nil {
				return err
			}

			// Read available artifacts.
			available, err := readAvailableArtifacts(specPath)
			if err != nil {
				return err
			}

			// Error if no artifacts are available.
			if len(available) == 0 {
				return fmt.Errorf("no renderable artifacts found in %q", specPath)
			}

			w := cmd.OutOrStdout()

			if jsonOutput {
				return renderJSON(w, available, combined)
			}

			return renderMarkdown(w, available, combined)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON envelope")
	cmd.Flags().BoolVar(&combined, "combined", false, "combine all artifacts into a single document")

	return cmd
}

// readAvailableArtifacts reads artifact files from the spec directory,
// returning a map of display-key to content for each file that exists
// and is readable. Returns an error if a file exists but cannot be read
// due to permissions.
func readAvailableArtifacts(specPath string) (map[string]string, error) {
	available := make(map[string]string)

	for _, filename := range renderArtifactOrder {
		p := filepath.Join(specPath, filename)
		info, statErr := os.Stat(p)
		if os.IsNotExist(statErr) {
			continue
		}
		if statErr != nil {
			return nil, fmt.Errorf("cannot stat %s: %w", filename, statErr)
		}
		if info.Mode().Perm()&0400 == 0 {
			// File exists but is not readable.
			data, readErr := os.ReadFile(p)
			if readErr != nil {
				return nil, fmt.Errorf("cannot read %s: %w", filename, readErr)
			}
			key := renderArtifactNames[filename]
			available[key] = string(data)
			continue
		}

		data, readErr := os.ReadFile(p)
		if readErr != nil {
			return nil, fmt.Errorf("cannot read %s: %w", filename, readErr)
		}

		key := renderArtifactNames[filename]
		available[key] = string(data)
	}

	return available, nil
}

// renderJSON outputs artifacts in a JSON envelope.
func renderJSON(w interface{ Write([]byte) (int, error) }, available map[string]string, combined bool) error {
	if combined {
		// Combine all artifact content into a single markdown document.
		content := combineArtifacts(available)
		return emitOKTo(w, "format", "combined", "content", content)
	}

	// Individual format: emit each artifact keyed by name.
	return emitOKTo(w, "format", "individual", "artifacts", available)
}

// renderMarkdown outputs artifacts as raw markdown to the writer.
func renderMarkdown(w interface{ Write([]byte) (int, error) }, available map[string]string, combined bool) error {
	if combined {
		content := combineArtifacts(available)
		_, err := fmt.Fprint(w, content)
		return err
	}

	// Print each artifact separated by '---', with a header for each.
	first := true
	for _, filename := range renderArtifactOrder {
		key := renderArtifactNames[filename]
		content, exists := available[key]
		if !exists {
			continue
		}
		if !first {
			fmt.Fprintln(w, "\n---")
		}
		fmt.Fprintf(w, "# %s\n\n", key)
		fmt.Fprintln(w, strings.TrimRight(content, "\n"))
		first = false
	}

	return nil
}

// combineArtifacts concatenates all available artifact content into a
// single markdown document with headers.
func combineArtifacts(available map[string]string) string {
	var sb strings.Builder
	for _, filename := range renderArtifactOrder {
		key := renderArtifactNames[filename]
		content, exists := available[key]
		if !exists {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(fmt.Sprintf("# %s\n\n", key))
		sb.WriteString(strings.TrimRight(content, "\n"))
	}
	return sb.String()
}
