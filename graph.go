package afspec

// BuildDependencyGraph reads tasks.json for each spec in the provided
// SpecMeta slice, extracts cross-spec dependency declarations, and
// returns a DependencyGraph.
//
// If a tasks.json references a spec ID not present in the provided
// SpecMeta slice, the dangling reference is recorded as an error and
// returned alongside the partial graph.
func BuildDependencyGraph(metas []SpecMeta, root string) (*DependencyGraph, error) {
	panic("not implemented")
}

// Dependencies returns all edges where the given spec ID is the target
// of a dependency (i.e., edges where ToSpec matches specID).
func (g *DependencyGraph) Dependencies(specID string) []DependencyEdge {
	panic("not implemented")
}

// Dependents returns all edges where the given spec ID is depended upon
// (i.e., edges where FromSpec matches specID).
func (g *DependencyGraph) Dependents(specID string) []DependencyEdge {
	panic("not implemented")
}

// TopologicalSort returns spec IDs in topological order using Kahn's
// algorithm such that every spec appears before its dependents.
// Returns an error if a cycle is detected.
func (g *DependencyGraph) TopologicalSort() ([]string, error) {
	panic("not implemented")
}
