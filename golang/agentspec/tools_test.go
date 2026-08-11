package agentspec

import (
	"testing"
)

// --- TS-06-36: AssessmentTools returns correct tool definition ---

// TestTS06_36_AssessmentToolsSchema verifies that AssessmentTools() returns a
// slice with exactly one entry for submit_assessment with the correct input
// schema fields: quality (enum), summary (string), gaps (array), questions
// (array of objects).
// Test Spec: TS-06-36, Requirement: 06-REQ-11.1
func TestTS06_36_AssessmentToolsSchema(t *testing.T) {
	tools := AssessmentTools()
	if len(tools) != 1 {
		t.Fatalf("len(AssessmentTools()) = %d; want 1", len(tools))
	}

	tool := tools[0]
	name, ok := tool["name"].(string)
	if !ok || name != "submit_assessment" {
		t.Errorf("tool[\"name\"] = %v; want %q", tool["name"], "submit_assessment")
	}

	// Verify input_schema exists and has properties.
	inputSchema, ok := tool["input_schema"].(map[string]any)
	if !ok {
		t.Fatalf("tool[\"input_schema\"] is %T; want map[string]any", tool["input_schema"])
	}

	props, ok := inputSchema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("input_schema[\"properties\"] is %T; want map[string]any", inputSchema["properties"])
	}

	// Check quality field has enum.
	qualityProp, ok := props["quality"].(map[string]any)
	if !ok {
		t.Fatalf("properties[\"quality\"] is %T; want map[string]any", props["quality"])
	}
	qualityEnum, ok := qualityProp["enum"].([]any)
	if !ok {
		t.Fatalf("quality[\"enum\"] is %T; want []any", qualityProp["enum"])
	}
	expectedEnums := map[string]bool{
		"ready":            false,
		"needs_refinement": false,
		"incomplete":       false,
	}
	for _, v := range qualityEnum {
		s, ok := v.(string)
		if !ok {
			continue
		}
		expectedEnums[s] = true
	}
	for k, found := range expectedEnums {
		if !found {
			t.Errorf("quality enum missing value %q", k)
		}
	}

	// Check summary field is string type.
	summaryProp, ok := props["summary"].(map[string]any)
	if !ok {
		t.Fatalf("properties[\"summary\"] is %T; want map[string]any", props["summary"])
	}
	if summaryProp["type"] != "string" {
		t.Errorf("summary[\"type\"] = %v; want %q", summaryProp["type"], "string")
	}

	// Check gaps field is array type.
	gapsProp, ok := props["gaps"].(map[string]any)
	if !ok {
		t.Fatalf("properties[\"gaps\"] is %T; want map[string]any", props["gaps"])
	}
	if gapsProp["type"] != "array" {
		t.Errorf("gaps[\"type\"] = %v; want %q", gapsProp["type"], "array")
	}

	// Check questions field is array type.
	questionsProp, ok := props["questions"].(map[string]any)
	if !ok {
		t.Fatalf("properties[\"questions\"] is %T; want map[string]any", props["questions"])
	}
	if questionsProp["type"] != "array" {
		t.Errorf("questions[\"type\"] = %v; want %q", questionsProp["type"], "array")
	}
}

// TestAssessmentTools_ReturnsNewSlice verifies that each call to
// AssessmentTools() returns a distinct slice value, preventing callers
// from mutating the shared definition.
// Edge Case: 06-REQ-11.E1
func TestAssessmentTools_ReturnsNewSlice(t *testing.T) {
	tools1 := AssessmentTools()
	tools2 := AssessmentTools()

	if len(tools1) == 0 || len(tools2) == 0 {
		t.Fatalf("AssessmentTools() returned empty slice; want non-empty")
	}

	// Mutate tools1 and verify tools2 is unaffected.
	tools1[0]["name"] = "mutated"

	tools3 := AssessmentTools()
	if len(tools3) == 0 {
		t.Fatal("AssessmentTools() returned empty slice after mutation")
	}

	name, ok := tools3[0]["name"].(string)
	if !ok || name != "submit_assessment" {
		t.Errorf("after mutation, AssessmentTools()[0][\"name\"] = %v; want %q (new slice expected)", tools3[0]["name"], "submit_assessment")
	}
}

// TestAssessmentTools_QuestionsObjectSchema verifies that the questions
// field contains an items schema with object properties including id,
// text, context, options, and required.
// Test Spec: TS-06-36 (detail), Requirement: 06-REQ-11.1
func TestAssessmentTools_QuestionsObjectSchema(t *testing.T) {
	tools := AssessmentTools()
	if len(tools) != 1 {
		t.Fatalf("len(AssessmentTools()) = %d; want 1", len(tools))
	}

	inputSchema, ok := tools[0]["input_schema"].(map[string]any)
	if !ok {
		t.Fatalf("input_schema not map[string]any")
	}
	props, ok := inputSchema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties not map[string]any")
	}
	questionsProp, ok := props["questions"].(map[string]any)
	if !ok {
		t.Fatalf("questions not map[string]any")
	}

	// The items should be an object with its own properties.
	items, ok := questionsProp["items"].(map[string]any)
	if !ok {
		t.Fatalf("questions[\"items\"] is %T; want map[string]any", questionsProp["items"])
	}
	itemProps, ok := items["properties"].(map[string]any)
	if !ok {
		t.Fatalf("items[\"properties\"] is %T; want map[string]any", items["properties"])
	}

	requiredFields := []string{"id", "text", "context", "options", "required"}
	for _, field := range requiredFields {
		if _, exists := itemProps[field]; !exists {
			t.Errorf("questions items properties missing field %q", field)
		}
	}
}

// --- TS-06-37: RefinementTools returns correct tool definitions ---

// TestTS06_37_RefinementToolsSchema verifies that RefinementTools() returns
// a slice with exactly two entries: submit_prd_update at index 0 and
// submit_assessment at index 1, with submit_prd_update having an
// updated_prd string field.
// Test Spec: TS-06-37, Requirement: 06-REQ-12.1
func TestTS06_37_RefinementToolsSchema(t *testing.T) {
	tools := RefinementTools()
	if len(tools) != 2 {
		t.Fatalf("len(RefinementTools()) = %d; want 2", len(tools))
	}

	// Check first tool is submit_prd_update.
	tool0 := tools[0]
	name0, ok := tool0["name"].(string)
	if !ok || name0 != "submit_prd_update" {
		t.Errorf("tools[0][\"name\"] = %v; want %q", tool0["name"], "submit_prd_update")
	}

	// Check submit_prd_update has updated_prd field.
	schema0, ok := tool0["input_schema"].(map[string]any)
	if !ok {
		t.Fatalf("tools[0][\"input_schema\"] is %T; want map[string]any", tool0["input_schema"])
	}
	props0, ok := schema0["properties"].(map[string]any)
	if !ok {
		t.Fatalf("input_schema[\"properties\"] is %T; want map[string]any", schema0["properties"])
	}
	updatedPRD, ok := props0["updated_prd"].(map[string]any)
	if !ok {
		t.Fatalf("properties[\"updated_prd\"] is %T; want map[string]any", props0["updated_prd"])
	}
	if updatedPRD["type"] != "string" {
		t.Errorf("updated_prd[\"type\"] = %v; want %q", updatedPRD["type"], "string")
	}

	// Check second tool is submit_assessment.
	tool1 := tools[1]
	name1, ok := tool1["name"].(string)
	if !ok || name1 != "submit_assessment" {
		t.Errorf("tools[1][\"name\"] = %v; want %q", tool1["name"], "submit_assessment")
	}
}

// TestRefinementTools_ReturnsNewSlice verifies that each call to
// RefinementTools() returns a distinct slice value.
// Edge Case: 06-REQ-12.E1
func TestRefinementTools_ReturnsNewSlice(t *testing.T) {
	tools1 := RefinementTools()
	tools2 := RefinementTools()

	if len(tools1) == 0 || len(tools2) == 0 {
		t.Fatalf("RefinementTools() returned empty slice; want non-empty")
	}

	// Mutate tools1 and verify a new call is unaffected.
	tools1[0]["name"] = "mutated"

	tools3 := RefinementTools()
	if len(tools3) == 0 {
		t.Fatal("RefinementTools() returned empty slice after mutation")
	}

	name, ok := tools3[0]["name"].(string)
	if !ok || name != "submit_prd_update" {
		t.Errorf("after mutation, RefinementTools()[0][\"name\"] = %v; want %q", tools3[0]["name"], "submit_prd_update")
	}
}
