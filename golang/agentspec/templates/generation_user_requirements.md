---
name: generation_user_requirements
description: Additional instructions for requirements artifact generation
---
Generate the requirements artifact following the EARS syntax format. Include acceptance criteria, execution paths, error handling, and external APIs.

## EARS Pattern Field Rules

Each acceptance criterion uses exactly one EARS pattern. The schema enforces a discriminated `oneOf` — using wrong or extra fields causes validation failure. Follow these rules precisely:

### ubiquitous
- **Required fields:** `id`, `ears_pattern`, `system`, `action`, `return_contract`
- **Forbidden fields:** `trigger`, `condition`, `error_condition`, `state`, `feature`
- Template: THE {system} SHALL {action}

### event_driven
- **Required fields:** `id`, `ears_pattern`, `trigger`, `system`, `action`, `return_contract`
- **Forbidden fields:** `condition`, `error_condition`, `state`, `feature`
- Template: WHEN {trigger}, THE {system} SHALL {action}

### complex_event
- **Required fields:** `id`, `ears_pattern`, `trigger`, `condition`, `system`, `action`, `return_contract`
- **Forbidden fields:** `error_condition`, `state`, `feature`
- Template: WHEN {trigger} AND {condition}, THE {system} SHALL {action}

### state_driven
- **Required fields:** `id`, `ears_pattern`, `state`, `system`, `action`, `return_contract`
- **Forbidden fields:** `trigger`, `condition`, `error_condition`, `feature`
- Template: WHILE {state}, THE {system} SHALL {action}

### unwanted
- **Required fields:** `id`, `ears_pattern`, `error_condition`, `system`, `action`, `return_contract`
- **Forbidden fields:** `trigger`, `condition`, `state`, `feature`
- Template: IF {error_condition}, THEN THE {system} SHALL {action}

### optional
- **Required fields:** `id`, `ears_pattern`, `feature`, `system`, `action`, `return_contract`
- **Forbidden fields:** `trigger`, `condition`, `error_condition`, `state`
- Template: WHERE {feature}, THE {system} SHALL {action}

## Glossary Backtick Convention

In the `action`, `trigger`, `condition`, `error_condition`, `state`, `feature`, `for_any`, and `invariant` fields: any domain-specific term must be wrapped in backticks (e.g. `` `SpaceManager` ``, `` `LivingSpec` ``). Every backtick-wrapped token must have a corresponding entry in the top-level `glossary` object. Unquoted natural-language words are not checked — only backtick-wrapped tokens are treated as domain terms requiring glossary entries.
