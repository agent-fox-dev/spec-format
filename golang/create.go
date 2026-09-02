package afspec

// Schema URIs for each artifact type.
const (
	requirementsSchemaURI = "https://agent-fox.dev/schemas/requirements.v1.json"
	testSpecSchemaURI     = "https://agent-fox.dev/schemas/test_spec.v1.json"
	tasksSchemaURI        = "https://agent-fox.dev/schemas/tasks.v1.json"
)

// CreateSpec creates a new Spec with status set to "draft", all sub-artifacts
// initialized with valid $schema, spec_id, spec_name, and schema_version
// fields so that Save + LoadSpec round-trips succeed without error.
// Validation of specID and specName is deferred to spec.Validate.
func CreateSpec(specID, specName string) *Spec {
	reqURI := requirementsSchemaURI
	tsURI := testSpecSchemaURI
	taskURI := tasksSchemaURI
	reqSchema := RequirementsV1JsonSchema(&reqURI)
	tsSchema := TestSpecV1JsonSchema(&tsURI)
	taskSchema := TasksV1JsonSchema(&taskURI)

	return &Spec{
		SpecID:        specID,
		SpecName:      specName,
		Status:        "draft",
		SchemaVersion: 1,
		Requirements: &RequirementsV1Json{
			Schema:                reqSchema,
			SpecId:                specID,
			SpecName:              specName,
			SchemaVersion:         1,
			Glossary:              RequirementsV1JsonGlossary{},
			Requirements:          []Requirement{},
			CorrectnessProperties: []CorrectnessProperty{},
			ExecutionPaths:        []ExecutionPath{},
			ErrorHandling:         []ErrorHandlingEntry{},
		},
		TestSpec: &TestSpecV1Json{
			Schema:        tsSchema,
			SpecId:        specID,
			SpecName:      specName,
			SchemaVersion: 1,
			TestCases:     []TestCase{},
			PropertyTests: []PropertyTest{},
			EdgeCaseTests: []EdgeCaseTest{},
			SmokeTests:    []SmokeTest{},
		},
		Tasks: &TasksV1Json{
			Schema:        taskSchema,
			SpecId:        specID,
			SpecName:      specName,
			SchemaVersion: 1,
			TestCommands:  TestCommands{},
			Dependencies:  []TaskDependency{},
			TaskGroups:    []TaskGroup{},
			Traceability:  []TraceabilityEntry{},
		},
	}
}
