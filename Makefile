# Makefile for the Gemara website (gemara.openssf.org)
#
# The site content lives at the repository root. Schema reference pages,
# the definitions table, and term cross-links are GENERATED from the Gemara
# specification's OpenAPI projection (openapi.yaml), which the spec repo
# (github.com/gemaraproj/gemara) publishes as a release asset. The markdown
# tooling that renders it lives in this repo under tools/.
#
# openapi.yaml acquisition, in order of precedence:
#   GEMARA_OPENAPI=/path/to/openapi.yaml   use a pre-generated file
#   GEMARA_DIR=../gemara                   generate from a local spec checkout
#                                          (runs its cue2openapi command)
#   GEMARA_REF=v1.2.3 (default: latest)    download the release asset

GEMARA_REPO ?= https://github.com/gemaraproj/gemara
GEMARA_REF  ?= latest

SITE_ABS      := $(abspath .)
TOOLS_DIR     := tools
GENERATED_DIR := generated
OPENAPI_YAML  := $(GENERATED_DIR)/openapi.yaml
SPEC_MD_DIR   := $(GENERATED_DIR)/spec
SCHEMA_DIR    := schema
SCHEMA_NAV    := schema-nav.yml

.PHONY: all fetch-openapi genmd gendocs serve build test test-tools test-links cleanup cleanup-links check-jekyll deps

all: test gendocs test-links cleanup

test-tools:
	@echo "  >  Running Go tests in tools/ ..."
	@cd $(TOOLS_DIR) && go test -v ./...

test: test-tools

deps:
	@echo "  >  Running bundle install..."
	@bundle install

check-jekyll:
	@if ! bundle exec jekyll -v >/dev/null 2>&1; then \
		echo "ERROR: Jekyll not available. Install dependencies: make deps"; \
		exit 1; \
	fi

# File target: if generated/openapi.yaml already exists (e.g. CI downloaded
# or generated it beforehand), acquisition is skipped entirely.
$(OPENAPI_YAML):
	@mkdir -p $(GENERATED_DIR)
	@if [ -n "$(GEMARA_OPENAPI)" ]; then \
		echo "  >  Using local OpenAPI file $(GEMARA_OPENAPI) ..."; \
		cp "$(GEMARA_OPENAPI)" "$(OPENAPI_YAML)"; \
	elif [ -n "$(GEMARA_DIR)" ]; then \
		echo "  >  Generating OpenAPI from local spec checkout $(GEMARA_DIR) ..."; \
		cd "$(abspath $(GEMARA_DIR))/cmd" && go run . cue2openapi \
			--schema .. \
			--output $(SITE_ABS)/$(OPENAPI_YAML); \
	else \
		if [ "$(GEMARA_REF)" = "latest" ]; then \
			url="$(GEMARA_REPO)/releases/latest/download/openapi.yaml"; \
		else \
			url="$(GEMARA_REPO)/releases/download/$(GEMARA_REF)/openapi.yaml"; \
		fi; \
		echo "  >  Downloading $$url ..."; \
		curl --fail --silent --show-error --location "$$url" --output "$(OPENAPI_YAML)" || { \
			rm -f "$(OPENAPI_YAML)"; \
			echo "ERROR: could not download openapi.yaml for spec ref '$(GEMARA_REF)'."; \
			echo "Releases before the asset existed can be built from a checkout instead:"; \
			echo "  make gendocs GEMARA_DIR=/path/to/gemara"; \
			exit 1; \
		}; \
	fi

fetch-openapi: $(OPENAPI_YAML)

genmd: fetch-openapi
	@echo "  >  Generating markdown from OpenAPI ..."
	@mkdir -p $(SPEC_MD_DIR)
	@cd $(TOOLS_DIR) && go run . openapi2md \
		--input $(SITE_ABS)/$(OPENAPI_YAML) \
		--output $(SITE_ABS)/$(SPEC_MD_DIR) \
		--nav $(SITE_ABS)/$(SCHEMA_NAV)
	@echo "  >  Markdown generation complete!"

gendocs: genmd
	@echo "  >  Copying schema pages to $(SCHEMA_DIR)/ for website ..."
	@mkdir -p $(SCHEMA_DIR)
	@sh "$(TOOLS_DIR)/scripts/parse-nav.sh" "$(SCHEMA_NAV)" list-pages | while IFS='|' read -r filename title; do \
		if [ -f "$(SPEC_MD_DIR)/$$filename.md" ]; then \
			{ \
				echo "---"; \
				echo "layout: page"; \
				echo "title: $$title"; \
				echo "---"; \
				echo ""; \
				cat "$(SPEC_MD_DIR)/$$filename.md"; \
			} > "$(SCHEMA_DIR)/$$filename.md"; \
		fi; \
	done
	@echo "  >  Updating schema list in $(SCHEMA_DIR)/index.md ..."
	@if [ -f "$(SCHEMA_DIR)/index.md" ]; then \
		schema_list_file="$(SCHEMA_DIR)/index.md.schema_list.tmp"; \
		sh "$(TOOLS_DIR)/scripts/parse-nav.sh" "$(SCHEMA_NAV)" list-pages | while IFS='|' read -r filename title; do \
			[ -f "$(SCHEMA_DIR)/$$filename.md" ] && echo "- [$$title]($$filename.html)"; \
		done > "$$schema_list_file"; \
		awk -v list_file="$$schema_list_file" ' \
			BEGIN { \
				while ((getline line < list_file) > 0) { \
					schema_list = schema_list line "\n"; \
				} \
				close(list_file); \
			} \
			/<!-- SCHEMA_LIST_START -->/ { \
				print; \
				print ""; \
				printf "%s", schema_list; \
				print ""; \
				skip=1; \
				next \
			} \
			/<!-- SCHEMA_LIST_END -->/ { print; skip=0; next } \
			skip==0 { print } \
		' "$(SCHEMA_DIR)/index.md" > "$(SCHEMA_DIR)/index.md.tmp" && \
		rm -f "$$schema_list_file" && \
		mv "$(SCHEMA_DIR)/index.md.tmp" "$(SCHEMA_DIR)/index.md"; \
	fi
	@echo "  >  Generating definitions table from lexicon ..."
	@if [ -f "model/02-definitions.md.template" ]; then \
		cp "model/02-definitions.md.template" "model/02-definitions.md"; \
	fi
	@cd $(TOOLS_DIR) && go run . lexicon2md \
		--lexicon $(SITE_ABS)/lexicon.yaml \
		--output $(SITE_ABS)/model/02-definitions.md
	@echo "  >  Linking defined terms across documentation ..."
	@cd $(TOOLS_DIR) && go run . termlinker \
		--lexicon $(SITE_ABS)/lexicon.yaml \
		--docs $(SITE_ABS)
	@echo "  >  Documentation generation complete!"

serve: check-jekyll gendocs
	@echo "  >  Starting Jekyll documentation site..."
	@bundle exec jekyll serve --host 0.0.0.0 --livereload

build: check-jekyll gendocs
	@echo "  >  Building Jekyll documentation site..."
	@bundle exec jekyll build

test-links:
	@echo "  >  Validating all site pages and links with html-proofer..."
	@bundle exec htmlproofer _site \
		--allow-hash-href \
		--disable-external \
		--ignore-empty-alt \
		--only-4xx \
		--ignore-files '/model\/02-definitions\.html/' \
		--root-dir "$$(pwd)/_site"

cleanup-links:
	@echo "  >  Removing termlinker-generated links from documentation ..."
	@cd $(TOOLS_DIR) && go run . termlinker \
		--lexicon $(SITE_ABS)/lexicon.yaml \
		--docs $(SITE_ABS) \
		--cleanup
	@echo "  >  Link cleanup complete!"

cleanup: cleanup-links
	@echo "  >  Removing generated documentation files and links..."
	@sh "$(TOOLS_DIR)/scripts/parse-nav.sh" "$(SCHEMA_NAV)" list-pages | while IFS='|' read -r filename title; do \
		rm -f "$(SCHEMA_DIR)/$$filename.md"; \
	done
	@rm -f model/02-definitions.md
	@git checkout -- $(SCHEMA_DIR)/index.md 2>/dev/null || true
	@rm -rf $(GENERATED_DIR) _site .jekyll-cache .jekyll-metadata
	@echo "  >  Cleanup complete!"
