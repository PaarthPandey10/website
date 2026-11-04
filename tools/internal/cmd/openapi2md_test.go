package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
)

func loadAllOfFixture(t *testing.T) OpenAPISpec {
	t.Helper()
	data, err := os.ReadFile("testdata/allof.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var spec OpenAPISpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	return spec
}

func TestResolveSchemaComposition(t *testing.T) {
	spec := loadAllOfFixture(t)
	schema, err := resolveSchemaByName("Derived", spec, make(map[string]bool))
	if err != nil {
		t.Fatal(err)
	}

	if got := schema.Properties["base"].(map[string]interface{})["type"]; got != "integer" {
		t.Fatalf("base override type = %v, want integer", got)
	}
	for _, name := range []string{"base", "middle", "own"} {
		if _, ok := schema.Properties[name]; !ok {
			t.Errorf("missing inherited property %q", name)
		}
	}
	for _, name := range []string{"base", "middle", "own"} {
		if !contains(schema.Required, name) {
			t.Errorf("missing required field %q", name)
		}
	}
}

func TestResolveSchemaCompositionCycle(t *testing.T) {
	spec := loadAllOfFixture(t)
	schema, err := resolveSchemaByName("CycleA", spec, make(map[string]bool))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a", "b"} {
		if _, ok := schema.Properties[name]; !ok {
			t.Errorf("missing cyclic property %q", name)
		}
	}
}

func TestGenerateRootSectionPreservesWrappedDescription(t *testing.T) {
	spec := loadAllOfFixture(t)
	schema, err := resolveSchemaByName("Wrapped", spec, make(map[string]bool))
	if err != nil {
		t.Fatal(err)
	}
	output := generateRootSection("Wrapped", schema, spec, map[string]string{})
	if strings.Count(output, "`metadata`") != 1 {
		t.Fatalf("metadata rendered more than once:\n%s", output)
	}
	if !strings.Contains(output, "metadata description") {
		t.Fatalf("wrapped description missing:\n%s", output)
	}
}

func TestResolveSchemaCompositionPreservesXStatus(t *testing.T) {
	spec := loadAllOfFixture(t)
	schema, err := resolveSchemaByName("StatusDerived", spec, make(map[string]bool))
	if err != nil {
		t.Fatal(err)
	}
	if schema.XStatus != "experimental" {
		t.Fatalf("x-status = %q, want experimental", schema.XStatus)
	}

	output := generateRootSection("StatusDerived", schema, spec, map[string]string{})
	if !strings.Contains(output, "badge-experimental") {
		t.Fatalf("status badge missing from generated output:\n%s", output)
	}
}

func TestFormatFieldTypeResolvesItemsAllOf(t *testing.T) {
	spec := loadAllOfFixture(t)
	schema, err := resolveSchemaByName("ItemsAllOfArray", spec, make(map[string]bool))
	if err != nil {
		t.Fatal(err)
	}

	itemsBytes, err := yaml.Marshal(schema.Properties["entries"])
	if err != nil {
		t.Fatal(err)
	}
	var entries Schema
	if err := yaml.Unmarshal(itemsBytes, &entries); err != nil {
		t.Fatal(err)
	}
	itemsBytes, err = yaml.Marshal(entries.Items)
	if err != nil {
		t.Fatal(err)
	}
	var items Schema
	if err := yaml.Unmarshal(itemsBytes, &items); err != nil {
		t.Fatal(err)
	}
	firstPart, err := parseSchema(items.AllOf[0])
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := resolveSchemaComposition(items, spec, make(map[string]bool))
	if err != nil {
		t.Fatal(err)
	}
	resolved.Ref = firstPart.Ref
	if !contains(resolved.Required, "id") {
		t.Fatal("resolved items lost inline required field id")
	}
	if _, ok := resolved.Properties["extra"]; !ok {
		t.Fatal("resolved items lost inline property extra")
	}

	output := generateRootSection("ItemsAllOfArray", schema, spec, map[string]string{})
	if !strings.Contains(output, "array[ItemTarget]") {
		t.Fatalf("resolved item type missing from generated output:\n%s", output)
	}
}

func TestStripCUEProjectionNotes(t *testing.T) {
	if got := stripCUEProjectionNotes("description\n\n(Enforced by the CUE schema; not by this OpenAPI projection.)"); got != "description" {
		t.Fatalf("stripped description = %q, want description", got)
	}
	if got := stripCUEProjectionNotes("description"); got != "description" {
		t.Fatalf("clean description changed to %q", got)
	}

	spec := loadAllOfFixture(t)
	schema, err := resolveSchemaByName("CUENoteSchema", spec, make(map[string]bool))
	if err != nil {
		t.Fatal(err)
	}
	output := generateRootSection("CUENoteSchema", schema, spec, map[string]string{})
	if strings.Contains(output, "Enforced by the CUE schema") {
		t.Fatalf("CUE projection note leaked into generated output:\n%s", output)
	}
	if !strings.Contains(output, "items is a list") || !strings.Contains(output, "name of the item") {
		t.Fatalf("description text missing from generated output:\n%s", output)
	}
}

func TestPropertyAllOfPreservesRef(t *testing.T) {
	spec := loadAllOfFixture(t)
	schema, err := resolveSchemaByName("PropRefAllOf", spec, make(map[string]bool))
	if err != nil {
		t.Fatal(err)
	}

	output := generateRootSection("PropRefAllOf", schema, spec, map[string]string{})
	if !strings.Contains(output, "**RefTarget**") {
		t.Fatalf("property allOf $ref lost — expected **RefTarget** in output:\n%s", output)
	}
	if strings.Contains(output, "**object**") {
		t.Fatalf("property allOf rendered as bare object instead of linked type:\n%s", output)
	}
	if !strings.Contains(output, "log maps to the entry") {
		t.Fatalf("property description lost:\n%s", output)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestSplitRoots(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"   ", nil},
		{"Root1", []string{"Root1"}},
		{"Root1, Root2, Root3", []string{"Root1", "Root2", "Root3"}},
		{"Root1,,  Root2 , ", []string{"Root1", "Root2"}},
	}

	for _, tt := range tests {
		got := splitRoots(tt.input)
		if len(got) != len(tt.want) {
			t.Fatalf("splitRoots(%q) length = %d, want %d", tt.input, len(got), len(tt.want))
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitRoots(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Control Catalog", "control-catalog"},
		{"Risk-Catalog", "risk-catalog"},
		{"Security Policy (v1.0)", "security-policy-v10"},
		{"Simple", "simple"},
	}

	for _, tt := range tests {
		if got := slugify(tt.input); got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatStatusBadge(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"experimental", "<span class=\"badge badge-experimental\">Experimental</span>"},
		{"stable", "<span class=\"badge badge-stable\">Stable</span>"},
		{"deprecated", "<span class=\"badge badge-deprecated\">Deprecated</span>"},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		if got := formatStatusBadge(tt.status); got != tt.want {
			t.Errorf("formatStatusBadge(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestLoadManifest(t *testing.T) {
	tempDir := t.TempDir()

	// Missing file error
	_, err := loadManifest(filepath.Join(tempDir, "nonexistent.json"))
	if err == nil {
		t.Error("expected error for missing manifest file, got nil")
	}

	// Invalid JSON error
	badJSON := filepath.Join(tempDir, "bad.json")
	if err := os.WriteFile(badJSON, []byte("invalid json"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = loadManifest(badJSON)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}

	// Valid manifest
	goodJSON := filepath.Join(tempDir, "good.json")
	content := `{"schema1.cue": ["ControlCatalog", "Policy"], "schema2.cue": ["Risk"]}`
	if err := os.WriteFile(goodJSON, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	m, err := loadManifest(goodJSON)
	if err != nil {
		t.Fatalf("unexpected error for valid manifest: %v", err)
	}
	if len(m) != 2 || len(m["schema1.cue"]) != 2 {
		t.Errorf("manifest parsed incorrectly: %+v", m)
	}
}

func TestLoadNavFile(t *testing.T) {
	tempDir := t.TempDir()

	// Missing file error
	_, err := loadNavFile(filepath.Join(tempDir, "nonexistent.yml"))
	if err == nil {
		t.Error("expected error for missing nav file, got nil")
	}

	// Invalid YAML error
	badYAML := filepath.Join(tempDir, "bad.yml")
	if err := os.WriteFile(badYAML, []byte("invalid: [unclosed"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err = loadNavFile(badYAML)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}

	// Valid nav file
	goodYAML := filepath.Join(tempDir, "good.yml")
	navContent := `
pages:
  - title: Control Catalog
    filename: control-catalog.md
    schemas:
      - ControlCatalog
      - Control
`
	if err := os.WriteFile(goodYAML, []byte(navContent), 0644); err != nil {
		t.Fatal(err)
	}
	nav, err := loadNavFile(goodYAML)
	if err != nil {
		t.Fatalf("unexpected error for valid nav file: %v", err)
	}
	if len(nav.Pages) != 1 || nav.Pages[0].Title != "Control Catalog" || len(nav.Pages[0].Schemas) != 2 {
		t.Errorf("nav parsed incorrectly: %+v", nav)
	}
}
