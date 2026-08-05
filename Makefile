.PHONY: clean test test-fast lint format check json-go install-skills uninstall-skills

SKILLS_TEMPLATES_DIR := $(CURDIR)/skills
CLAUDE_SKILLS_DIR := $(HOME)/.claude/skills

clean:
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name '*.pyc' -delete 2>/dev/null || true
	find . -type f -name '*.pyo' -delete 2>/dev/null || true
	rm -rf .pytest_cache/ *.egg-info/ dist/ .ruff_cache/ .mypy_cache/ .hypothesis/
	rm -rf packages/*/.pytest_cache packages/*/.mypy_cache packages/*/.ruff_cache
	rm -rf packages/*/build packages/*/dist packages/*/*.egg-info

test:
	uv run pytest -q

test-fast:
	uv run pytest -m "not slow" -q

lint:
	uv run ruff check packages/ && uv run ruff format --check packages/

format:
	uv run ruff format packages/

check: lint test

json-go-gen:
	go get github.com/atombender/go-jsonschema/...
	go install github.com/atombender/go-jsonschema@latest
	rm -rf schemas && mkdir -p schemas
	cp packages/afspec/afspec/schemas/*.json schemas/
	go-jsonschema -p afspec schemas/tasks.v1.json > tasks.v1.go
	go-jsonschema -p afspec schemas/requirements.v1.json > requirements.v1.go
	go-jsonschema -p afspec schemas/test_spec.v1.json > test_spec.v1.go
	go-jsonschema -p afspec schemas/prd-frontmatter.v1.json > prd-frontmatter.v1.go
	 

install-skills:
	@for skill in $(SKILLS_TEMPLATES_DIR)/*; do \
		name=$$(basename "$$skill"); \
		target="$(CLAUDE_SKILLS_DIR)/$$name"; \
		mkdir -p "$$target"; \
		cp "$$skill" "$$target/SKILL.md"; \
		echo "installed: $$name -> $$target/SKILL.md"; \
	done

uninstall-skills:
	@for skill in $(SKILLS_TEMPLATES_DIR)/*; do \
		name=$$(basename "$$skill"); \
		if [ -d "$(CLAUDE_SKILLS_DIR)/$$name" ]; then \
			rm -rf "$(CLAUDE_SKILLS_DIR)/$$name"; \
			echo "removed: $$name"; \
		fi; \
	done