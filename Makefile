.PHONY: clean test test-fast lint format check json-go install-skills uninstall-skills

GOLANG_DIR := $(CURDIR)/golang
PYTHON_DIR := $(CURDIR)/packages/afspec

SCHEMAS_DIR := $(CURDIR)/schemas
GO_SCHEMAS_DIR := $(GOLANG_DIR)/schemas
PYTHON_SCHEMAS_DIR := $(PYTHON_DIR)/afspec/schemas

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
	cd $(GOLANG_DIR) && go test ./... -count=1

test-fast:
	uv run pytest -m "not slow" -q
	cd $(GOLANG_DIR) && go test ./... -count=1

lint:
	uv run ruff check packages/ && uv run ruff format --check packages/
	cd $(GOLANG_DIR) && test -z "$$(gofmt -l .)"
	cd $(GOLANG_DIR) && go vet ./...

format:
	uv run ruff format packages/
	cd $(GOLANG_DIR) && gofmt -w .

check: lint test

json-gen:
	cd golang && go get github.com/atombender/go-jsonschema/...
	cd golang && go install github.com/atombender/go-jsonschema@latest
	rm -rf $(GO_SCHEMAS_DIR)/*.json && cp $(SCHEMAS_DIR)/*.json $(GO_SCHEMAS_DIR)/
	rm -rf $(PYTHON_SCHEMAS_DIR)/*.json && cp $(SCHEMAS_DIR)/*.json $(PYTHON_SCHEMAS_DIR)/
	cd golang && go-jsonschema -p afspec $(GO_SCHEMAS_DIR)/tasks.v1.json > $(GOLANG_DIR)/tasks.v1.go
	cd golang && go-jsonschema -p afspec $(GO_SCHEMAS_DIR)/requirements.v1.json > $(GOLANG_DIR)/requirements.v1.go
	cd golang && go-jsonschema -p afspec $(GO_SCHEMAS_DIR)/test_spec.v1.json > $(GOLANG_DIR)/test_spec.v1.go
	cd golang && go-jsonschema -p afspec $(GO_SCHEMAS_DIR)/prd-frontmatter.v1.json > $(GOLANG_DIR)/prd-frontmatter.v1.go

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