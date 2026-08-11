package spec

import (
	"fmt"
)

// resolveSpec resolves a spec argument (either a numeric prefix or an
// exact directory name) within the given spec directory. Returns the
// full path to the matched spec directory.
//
// If the argument is purely numeric, it matches entries whose directory
// name starts with that numeric prefix. Otherwise, it matches entries
// whose directory name equals the argument exactly.
//
// Returns an error if zero or multiple matches are found, or if the
// spec directory cannot be read.
func resolveSpec(specDir, arg string) (string, error) {
	_ = specDir
	_ = arg
	return "", fmt.Errorf("resolveSpec: not implemented")
}
