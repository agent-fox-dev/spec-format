## Documentation Freshness

After implementing any spec, you **must** update all affected documentation
before the session is considered complete. Outdated docs are treated as
regressions.

**When to update:** Every time a spec implementation adds, changes, or removes
any of the items listed below, update the corresponding document in the same
session — not as a follow-up task.

| What changed | Update |
|---|---|
| CLI commands, subcommands, or flags added/changed/removed | `docs/cli.md` |
| Setup, quickstart, or project overview changed | `README.md` |
| package afspec | `packages/afspec/README.md` |
| package agentspec | `packages/agentspec/README.md` |
| package spec | `packages/spec/README.md` |

**Instructions:**

1. Review the spec you just implemented and identify every user-facing or
   developer-facing surface that changed (API routes, request/response
   schemas, CLI parameters, environment variables, architectural decisions).
2. Open each affected doc and update it to reflect the new state. Do not
   leave placeholder text like "TODO" or "TBD" — write the actual content.
3. If a doc file listed above does not exist yet, create it with the correct
   content rather than skipping the update.
4. Run `make check` after doc updates to ensure nothing is broken.

IMPORTANT: spec-format.md is a source-of-truth, never modify it.
