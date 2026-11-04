// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLexicon2MDGeneration(t *testing.T) {
	tempDir := t.TempDir()
	lexiconFile := filepath.Join(tempDir, "lexicon.yaml")
	templateFile := filepath.Join(tempDir, "definitions.md")

	lexiconContent := `title: Test Lexicon
metadata:
  id: test-lexicon
  type: Lexicon
terms:
  - id: control
    title: Control
    definition: a safeguard or countermeasure to mitigate risk | piped detail
    references:
      - citation: Layer 2
      - citation: NIST 800-53
  - id: policy
    title: Policy Document
    definition: organizational rules and mandates
    references:
      - citation: Layer 3
`
	if err := os.WriteFile(lexiconFile, []byte(lexiconContent), 0644); err != nil {
		t.Fatalf("failed to write test lexicon: %v", err)
	}

	templateContent := `---
layout: page
title: Definitions
---

# Common Definitions

| Term | Definition | Applies To |
|------|------------|------------|
{{.Table}}
`
	if err := os.WriteFile(templateFile, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write test template: %v", err)
	}

	// Configure flags for command run
	lexicon2MDFlags.lexiconFile = lexiconFile
	lexicon2MDFlags.outputFile = templateFile

	if err := runLexicon2MD(lexicon2MDCmd, []string{}); err != nil {
		t.Fatalf("runLexicon2MD failed: %v", err)
	}

	result, err := os.ReadFile(templateFile)
	if err != nil {
		t.Fatalf("failed to read generated output: %v", err)
	}

	output := string(result)

	// Verify table entries
	if !strings.Contains(output, `<a id="control"></a>**Control**`) {
		t.Errorf("output missing anchor for Control:\n%s", output)
	}
	// Verify pipe character escaping
	if !strings.Contains(output, `mitigate risk \| piped detail`) {
		t.Errorf("pipe escaping failed in definition:\n%s", output)
	}
	// Verify references joined by <br>
	if !strings.Contains(output, `Layer 2<br>NIST 800-53`) {
		t.Errorf("multiple references not formatted with <br>:\n%s", output)
	}
	if !strings.Contains(output, `<a id="policy-document"></a>**Policy Document**`) {
		t.Errorf("output missing anchor for Policy Document:\n%s", output)
	}
	if !strings.Contains(output, `Layer 3`) {
		t.Errorf("reference Layer 3 missing:\n%s", output)
	}
}

func TestLexicon2MDMissingFiles(t *testing.T) {
	tempDir := t.TempDir()

	lexicon2MDFlags.lexiconFile = filepath.Join(tempDir, "nonexistent-lexicon.yaml")
	lexicon2MDFlags.outputFile = filepath.Join(tempDir, "nonexistent-output.md")

	if err := runLexicon2MD(lexicon2MDCmd, []string{}); err == nil {
		t.Fatal("expected error on missing lexicon file, got nil")
	}

	// Create valid lexicon file but missing template file
	validLex := filepath.Join(tempDir, "valid.yaml")
	if err := os.WriteFile(validLex, []byte("title: Test\nterms: []\n"), 0644); err != nil {
		t.Fatal(err)
	}
	lexicon2MDFlags.lexiconFile = validLex
	lexicon2MDFlags.outputFile = filepath.Join(tempDir, "missing-template.md")

	if err := runLexicon2MD(lexicon2MDCmd, []string{}); err == nil {
		t.Fatal("expected error on missing template file, got nil")
	}
}

func TestLexicon2MDInvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	invalidLex := filepath.Join(tempDir, "invalid.yaml")
	tmplFile := filepath.Join(tempDir, "tmpl.md")

	if err := os.WriteFile(invalidLex, []byte("invalid: yaml: [unclosed"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tmplFile, []byte("{{.Table}}"), 0644); err != nil {
		t.Fatal(err)
	}

	lexicon2MDFlags.lexiconFile = invalidLex
	lexicon2MDFlags.outputFile = tmplFile

	if err := runLexicon2MD(lexicon2MDCmd, []string{}); err == nil {
		t.Fatal("expected error on malformed YAML, got nil")
	}
}

func TestLexicon2MDInvalidTemplate(t *testing.T) {
	tempDir := t.TempDir()
	validLex := filepath.Join(tempDir, "valid.yaml")
	invalidTmpl := filepath.Join(tempDir, "invalid_tmpl.md")

	if err := os.WriteFile(validLex, []byte("title: Test\nterms: []\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(invalidTmpl, []byte("{{ .BadField / syntax error }}"), 0644); err != nil {
		t.Fatal(err)
	}

	lexicon2MDFlags.lexiconFile = validLex
	lexicon2MDFlags.outputFile = invalidTmpl

	if err := runLexicon2MD(lexicon2MDCmd, []string{}); err == nil {
		t.Fatal("expected error on invalid template syntax, got nil")
	}
}
