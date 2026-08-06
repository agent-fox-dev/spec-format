package afspec

// CreateSpec creates a new Spec with status set to "draft", all sub-artifacts
// initialized to their zero/empty values, and the given specID and specName.
// Validation of specID and specName is deferred to spec.Validate.
func CreateSpec(specID, specName string) *Spec {
	return &Spec{
		SpecID:        specID,
		SpecName:      specName,
		Status:        "draft",
		SchemaVersion: 1,
		Requirements:  &RequirementsV1Json{},
		TestSpec:      &TestSpecV1Json{},
		Tasks:         &TasksV1Json{},
	}
}
