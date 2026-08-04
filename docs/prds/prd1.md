# Changes to spec CLI and campaign handling

The way how specs with and without a campaign are created is inconsistent.

## Definitions

<spec_root>: location where spec looks for specs or writes specs into. Default: ".specs/" (relative to current directory).
<context_dir>: creating specs requires eg analysing source code or reading other files. spec looks for these sources in <context_dir>. Default: current directory

## Rules

- Specs in <spec_root>/ are part of the "default" campaign, ie all specs in <spec_root> are treated as one campaign.
- if "spec new" is called the first time (without --spec-dir parameter) and there is no ".spec" folder in the current working directory, a new ".specs" folder is created, with a "campaign.yaml" file in it. This is same as calling

```
spec campaign create -p ".spec" -n "default" --description "default campaign"
```

## Changes

Remove these commands from the CLI:

- spec campaign open
- spec campaign new-spec

Changes to existing CLI commands:

"spec campaign create" simply becomes "spec campaign" with same signature as "spec campaign create"

Every command accepts these two flags:
--spec-dir: sets the <spec_root>
--source: sets the <context_dir>

Spec these changes to the CLI and agentspec. Update the various docs and READMEs accordingly.

.spec/config.toml moves to .specs/config.toml. Same for the global config file: $HOME/.specs/config.toml

on config.toml: 
- remove all configuration options that are not related to LLM provider and model selection, eg stuff like "[theme]"
- remove [spec_tool]
- add [provider] for provider configuration
- add [model] for model selection


## Clarifications

- spec campaign is only for creating campaigns at non-default locations
- --source / <context_dir> is a new concept: Set the working directory for source code analysis during AI operations
- --source applies only AI-driven commands eg "refine"
- What replaces spec campaign open? drop it, replace with "spec list"
- Spec numbering across archive: spec new now also check the archive to avoid prefix reuse
- spec new still accepts --name
- Migration: no need for any migration
