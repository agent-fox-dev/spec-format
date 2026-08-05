package afspec

// UbiquitousCriterion constructs a Criterion with the ubiquitous EARS pattern.
// Sets Id, EarsPattern, System, and Action. No pattern-specific optional fields.
func UbiquitousCriterion(id, system, action string) Criterion {
	panic("not implemented")
}

// EventDrivenCriterion constructs a Criterion with the event_driven EARS pattern.
// Sets Id, EarsPattern, Trigger, System, and Action.
func EventDrivenCriterion(id, trigger, system, action string) Criterion {
	panic("not implemented")
}

// ComplexEventCriterion constructs a Criterion with the complex_event EARS pattern.
// Sets Id, EarsPattern, Trigger, Condition, System, and Action.
func ComplexEventCriterion(id, trigger, condition, system, action string) Criterion {
	panic("not implemented")
}

// StateDrivenCriterion constructs a Criterion with the state_driven EARS pattern.
// Sets Id, EarsPattern, State, System, and Action.
func StateDrivenCriterion(id, state, system, action string) Criterion {
	panic("not implemented")
}

// UnwantedCriterion constructs a Criterion with the unwanted EARS pattern.
// Sets Id, EarsPattern, ErrorCondition, System, and Action.
// Does not set Trigger, State, or Feature (they remain nil).
func UnwantedCriterion(id, errorCondition, system, action string) Criterion {
	panic("not implemented")
}

// OptionalCriterion constructs a Criterion with the optional EARS pattern.
// Sets Id, EarsPattern, Feature, System, and Action.
// Does not set Trigger, State, or ErrorCondition (they remain nil).
func OptionalCriterion(id, feature, system, action string) Criterion {
	panic("not implemented")
}

// RenderEARSSentence renders a single EARS criterion as a natural-language
// sentence following the EARS pattern templates:
//
//   - ubiquitous:    "THE <System> SHALL <Action>"
//   - event_driven:  "WHEN <Trigger>, THE <System> SHALL <Action>"
//   - complex_event: "WHEN <Trigger> IF <Condition>, THE <System> SHALL <Action>"
//   - state_driven:  "WHILE <State>, THE <System> SHALL <Action>"
//   - unwanted:      "IF <ErrorCondition>, THE <System> SHALL <Action>"
//   - optional:      "WHERE <Feature>, THE <System> SHALL <Action>"
func (c Criterion) RenderEARSSentence() string {
	panic("not implemented")
}
