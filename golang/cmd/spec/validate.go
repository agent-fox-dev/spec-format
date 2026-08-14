package spec

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	afspec "github.com/agent-fox-dev/spec-format"
	"github.com/spf13/cobra"
)

// requiredFiles lists the files that every spec must contain.
var requiredFiles = []string{"prd.md", "requirements.json", "test_spec.json", "tasks.json"}

// validationError represents a single validation finding.
type validationError struct {
	File     string `json:"file"`
	Severity string `json:"severity"` // "error" or "warning"
	Message  string `json:"message"`
}

// validationResult holds the aggregated validation output for one or more specs.
type validationResult struct {
	Valid        bool              `json:"valid"`
	ErrorCount   int               `json:"error_count"`
	WarningCount int               `json:"warning_count"`
	Errors       []validationError `json:"errors,omitempty"`
}

// mergeResult merges src into dst in place.
func mergeResult(dst, src *validationResult) {
	if !src.Valid {
		dst.Valid = false
	}
	dst.ErrorCount += src.ErrorCount
	dst.WarningCount += src.WarningCount
	dst.Errors = append(dst.Errors, src.Errors...)
}

// newValidateCmd creates the "spec validate" subcommand which validates
// one or all specs for structural correctness.
func newValidateCmd() *cobra.Command {
	var cross bool
	var short bool

	cmd := &cobra.Command{
		Use:   "validate [SPEC]",
		Short: "Validate one or all specs for structural correctness",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			specDir, _ := cmd.Flags().GetString("spec-dir")
			quiet, _ := cmd.Flags().GetBool("quiet")
			if isAgentMode() {
				quiet = true
			}
			w := cmd.OutOrStdout()

			if len(args) == 1 {
				// Single-spec mode.
				return runValidateSingle(specDir, args[0], short, w)
			}

			// Multi-spec or cross-spec mode.
			if cross {
				return runValidateCross(specDir, short, quiet, w, cmd)
			}
			return runValidateAll(specDir, short, quiet, w, cmd)
		},
	}

	cmd.Flags().BoolVar(&cross, "cross", false, "run cross-spec interface consistency checks")
	cmd.Flags().BoolVar(&short, "short", false, "emit condensed output with only valid/error_count/warning_count")

	return cmd
}

// runValidateSingle validates a single spec by name.
func runValidateSingle(specDir, specName string, short bool, w interface{ Write([]byte) (int, error) }) error {
	specPath, err := resolveSpec(specDir, specName)
	if err != nil {
		return err
	}

	result := validateSpec(specPath)
	return emitValidationResult(w, result, short)
}

// runValidateAll discovers all specs and validates each, aggregating results.
func runValidateAll(specDir string, short, quiet bool, w interface{ Write([]byte) (int, error) }, cmd *cobra.Command) error {
	specs, err := discoverSpecDirs(specDir)
	if err != nil {
		return fmt.Errorf("cannot discover specs: %w", err)
	}

	spinner := NewStatusSpinner("Validating specs...", cmd.ErrOrStderr(), quiet, false)
	spinner.Start()
	defer spinner.Stop()

	agg := &validationResult{Valid: true}
	for _, sp := range specs {
		r := validateSpec(sp)
		mergeResult(agg, r)
	}

	return emitValidationResult(w, agg, short)
}

// runValidateCross discovers all specs, loads them via the library,
// builds a dependency graph, runs cross-spec validation via
// afspec.ValidateCrossSpec, and merges results.
func runValidateCross(specDir string, short, quiet bool, w interface{ Write([]byte) (int, error) }, cmd *cobra.Command) error {
	specPaths, err := discoverSpecDirs(specDir)
	if err != nil {
		return fmt.Errorf("cannot discover specs: %w", err)
	}

	spinner := NewStatusSpinner("Running cross-spec validation...", cmd.ErrOrStderr(), quiet, false)
	spinner.Start()
	defer spinner.Stop()

	// First run per-spec structural validation.
	agg := &validationResult{Valid: true}
	for _, sp := range specPaths {
		r := validateSpec(sp)
		mergeResult(agg, r)
	}

	// Load all specs via the library and build a dependency graph.
	var loadedSpecs []*afspec.Spec
	var metas []afspec.SpecMeta
	for _, sp := range specPaths {
		spec, loadErr := afspec.LoadSpec(sp)
		if loadErr != nil {
			// Surface load failure as a validation error, not a crash.
			agg.Errors = append(agg.Errors, validationError{
				File:     filepath.Base(sp),
				Severity: "error",
				Message:  fmt.Sprintf("cannot load spec %s: %s", filepath.Base(sp), loadErr),
			})
			agg.ErrorCount++
			agg.Valid = false
			continue
		}
		loadedSpecs = append(loadedSpecs, spec)
		metas = append(metas, afspec.SpecMeta{
			SpecID:   spec.SpecID,
			SpecName: spec.SpecName,
			Status:   spec.Status,
			Dir:      sp,
		})
	}

	// Build dependency graph and run library cross-spec validation.
	if len(loadedSpecs) > 0 {
		graph, _ := afspec.BuildDependencyGraph(metas, specDir)
		libResult := afspec.ValidateCrossSpec(loadedSpecs, graph)
		// Map library ValidationEntry errors to CLI validationError structs.
		for _, entry := range libResult.Errors {
			agg.Errors = append(agg.Errors, validationError{
				File:     entry.Artifact,
				Severity: "error",
				Message:  entry.Message,
			})
			agg.ErrorCount++
		}
		if !libResult.Valid {
			agg.Valid = false
		}
	}

	// Run CLI-level cross-spec check: duplicate requirement IDs.
	crossResult := validateCrossSpecs(specPaths)
	mergeResult(agg, crossResult)

	return emitValidationResult(w, agg, short)
}

// validateSpec checks a single spec directory for required files,
// JSON readability, and structural validity.
func validateSpec(specPath string) *validationResult {
	result := &validationResult{Valid: true}

	for _, f := range requiredFiles {
		p := filepath.Join(specPath, f)
		data, err := os.ReadFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				result.Errors = append(result.Errors, validationError{
					File:     f,
					Severity: "error",
					Message:  fmt.Sprintf("required file %s is missing", f),
				})
				result.ErrorCount++
				result.Valid = false
				continue
			}
			result.Errors = append(result.Errors, validationError{
				File:     f,
				Severity: "error",
				Message:  fmt.Sprintf("cannot read %s: %v", f, err),
			})
			result.ErrorCount++
			result.Valid = false
			continue
		}

		// prd.md is not JSON — skip JSON parse check.
		if f == "prd.md" {
			if len(data) == 0 {
				result.Errors = append(result.Errors, validationError{
					File:     f,
					Severity: "warning",
					Message:  "prd.md is empty",
				})
				result.WarningCount++
			}
			continue
		}

		// Verify the JSON artifact is valid JSON.
		if !json.Valid(data) {
			result.Errors = append(result.Errors, validationError{
				File:     f,
				Severity: "error",
				Message:  fmt.Sprintf("%s contains malformed JSON", f),
			})
			result.ErrorCount++
			result.Valid = false
		}
	}

	return result
}

// validateCrossSpecs runs CLI-level cross-spec consistency checks.
// This is a supplementary check that detects duplicate requirement IDs
// across specs, complementing the library's ValidateCrossSpec checks.
func validateCrossSpecs(specPaths []string) *validationResult {
	result := &validationResult{Valid: true}

	// Cross-spec validation: check for duplicate requirement IDs.
	seenReqIDs := make(map[string]string) // reqID -> specPath
	for _, sp := range specPaths {
		reqPath := filepath.Join(sp, "requirements.json")
		data, err := os.ReadFile(reqPath)
		if err != nil {
			continue
		}

		var doc struct {
			Requirements []struct {
				ID string `json:"id"`
			} `json:"requirements"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			continue
		}

		for _, req := range doc.Requirements {
			if req.ID == "" {
				continue
			}
			if prevSpec, exists := seenReqIDs[req.ID]; exists {
				result.Errors = append(result.Errors, validationError{
					File:     "requirements.json",
					Severity: "error",
					Message:  fmt.Sprintf("duplicate requirement ID %q: found in %s and %s", req.ID, filepath.Base(prevSpec), filepath.Base(sp)),
				})
				result.ErrorCount++
				result.Valid = false
			} else {
				seenReqIDs[req.ID] = sp
			}
		}
	}

	return result
}

// discoverSpecDirs scans the spec directory for subdirectories matching the
// spec naming pattern (NN_name) and returns their full paths.
func discoverSpecDirs(specDir string) ([]string, error) {
	entries, err := os.ReadDir(specDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read spec directory %q: %w", specDir, err)
	}

	var specs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if specDirPattern.MatchString(e.Name()) {
			specs = append(specs, filepath.Join(specDir, e.Name()))
		}
	}
	return specs, nil
}

// emitValidationResult writes the validation result to the writer.
// If short is true, only the condensed fields are emitted.
// Returns a non-nil error (for cobra exit code 1) if the result contains errors.
func emitValidationResult(w interface{ Write([]byte) (int, error) }, result *validationResult, short bool) error {
	if short {
		condensed := map[string]any{
			"ok":            result.Valid,
			"valid":         result.Valid,
			"error_count":   result.ErrorCount,
			"warning_count": result.WarningCount,
		}
		if err := emitTo(w, condensed); err != nil {
			return err
		}
	} else {
		full := map[string]any{
			"ok":            result.Valid,
			"valid":         result.Valid,
			"error_count":   result.ErrorCount,
			"warning_count": result.WarningCount,
			"errors":        result.Errors,
		}
		// Ensure errors is always an array, never null.
		if result.Errors == nil {
			full["errors"] = []validationError{}
		}
		if err := emitTo(w, full); err != nil {
			return err
		}
	}

	// Exit code 1 (return error) if there are validation errors.
	if result.ErrorCount > 0 {
		return fmt.Errorf("validation failed: %d error(s) found", result.ErrorCount)
	}
	return nil
}
