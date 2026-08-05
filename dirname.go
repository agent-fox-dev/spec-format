package afspec

// IsSpecDirName validates whether a directory name matches the
// NN_snake_case pattern (two-digit numeric prefix, underscore, then
// one or more snake_case segments).
func IsSpecDirName(name string) bool {
	panic("not implemented")
}

// ParseSpecDirName parses a valid NN_snake_case directory name and
// returns the numeric prefix and the snake_case name portion as
// separate values.
//
// Returns an error if the name does not match the NN_snake_case pattern.
func ParseSpecDirName(name string) (string, string, error) {
	panic("not implemented")
}
