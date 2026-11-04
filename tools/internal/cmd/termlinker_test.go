// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestProcessContentPreservesLiquidOutput(t *testing.T) {
	termInfos := []TermInfo{{
		OriginalTerm: "organization",
		Slug:         "organization",
		Regex:        regexp.MustCompile(`(?i)\borganization\b`),
	}}

	for _, content := range []string{
		"{{ maintainer.organization }}",
		"{{ organization | append: \"}\" }}",
		"{% assign organization = \"100%\" %}",
	} {
		if got := processContent(content, termInfos, "../model/02-definitions.html"); got != content {
			t.Fatalf("processContent() = %q, want %q", got, content)
		}
	}
}

func TestBuildTermInfos(t *testing.T) {
	terms := []Term{
		{Title: "Risk"},
		{Title: "Risk Assessment"},
		{Title: "Special & Characters (Test)"},
	}

	infos := buildTermInfos(terms)
	if len(infos) != 3 {
		t.Fatalf("expected 3 term infos, got %d", len(infos))
	}

	// Verify sorting: longest term first to prevent partial shadowing
	if infos[0].OriginalTerm != "Special & Characters (Test)" {
		t.Errorf("expected longest term first, got %q", infos[0].OriginalTerm)
	}
	if infos[1].OriginalTerm != "Risk Assessment" {
		t.Errorf("expected 'Risk Assessment' second, got %q", infos[1].OriginalTerm)
	}
	if infos[2].OriginalTerm != "Risk" {
		t.Errorf("expected 'Risk' third, got %q", infos[2].OriginalTerm)
	}

	// Verify slugs
	if infos[1].Slug != "risk-assessment" {
		t.Errorf("expected slug 'risk-assessment', got %q", infos[1].Slug)
	}
	if infos[2].Slug != "risk" {
		t.Errorf("expected slug 'risk', got %q", infos[2].Slug)
	}

	// Test regex match
	if !infos[1].Regex.MatchString("Conducting a risk assessment today.") {
		t.Error("regex failed to match 'risk assessment' in sentence")
	}
}

func TestTermToSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Risk Assessment", "risk-assessment"},
		{"Access Control (AC)", "access-control-ac"},
		{"Multi-Factor Authentication", "multi-factor-authentication"},
		{"Simple", "simple"},
	}

	for _, tt := range tests {
		got := termToSlug(tt.input)
		if got != tt.want {
			t.Errorf("termToSlug(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCalculateRelativePath(t *testing.T) {
	docsDir := filepath.Join("website", "docs")

	// File in subfolder docs/schema/controls.md
	filePath := filepath.Join(docsDir, "schema", "controls.md")
	rel := calculateRelativePath(filePath, docsDir)
	rel = filepath.ToSlash(rel)
	if !strings.HasSuffix(rel, "model/02-definitions.html") {
		t.Errorf("calculateRelativePath() = %q, want suffix model/02-definitions.html", rel)
	}

	// File in root docs/index.md
	rootFile := filepath.Join(docsDir, "index.md")
	relRoot := calculateRelativePath(rootFile, docsDir)
	relRoot = filepath.ToSlash(relRoot)
	if relRoot != "model/02-definitions.html" {
		t.Errorf("calculateRelativePath() for root file = %q, want 'model/02-definitions.html'", relRoot)
	}
}

func TestCleanupLine(t *testing.T) {
	termInfos := []TermInfo{
		{
			OriginalTerm: "Risk Assessment",
			Slug:         "risk-assessment",
		},
		{
			OriginalTerm: "Control",
			Slug:         "control",
		},
	}

	defPath := "../model/02-definitions.html"

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "replaces definition link with plain text",
			input: "Perform a [Risk Assessment](../model/02-definitions.html#risk-assessment) on the system.",
			want:  "Perform a Risk Assessment on the system.",
		},
		{
			name:  "preserves external markdown links",
			input: "See [GitHub](https://github.com/gemaraproj/gemara) for details.",
			want:  "See [GitHub](https://github.com/gemaraproj/gemara) for details.",
		},
		{
			name:  "preserves other internal links not pointing to 02-definitions",
			input: "Refer to [Architecture](../model/01-architecture.html#overview) guide.",
			want:  "Refer to [Architecture](../model/01-architecture.html#overview) guide.",
		},
		{
			name:  "multiple definition links in single line",
			input: "[Control](../model/02-definitions.html#control) and [Risk Assessment](../model/02-definitions.html#risk-assessment).",
			want:  "Control and Risk Assessment.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanupLine(tt.input, termInfos, defPath)
			if got != tt.want {
				t.Errorf("cleanupLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContentStateSkipLine(t *testing.T) {
	state := &contentState{}

	// Normal line should not skip
	if state.skipLine("This is regular text.") {
		t.Error("expected regular text not to be skipped")
	}

	// Heading line should skip
	if !state.skipLine("# Header 1") {
		t.Error("expected heading to be skipped")
	}
	if !state.skipLine("### Subheader") {
		t.Error("expected subheader to be skipped")
	}

	// In front matter
	state.inFrontMatter = true
	if !state.skipLine("title: My Page") {
		t.Error("expected front matter line to be skipped")
	}
	state.inFrontMatter = false

	// In code block
	state.inCodeBlock = true
	if !state.skipLine("val x = 10") {
		t.Error("expected code block line to be skipped")
	}
	state.inCodeBlock = false

	// In Jekyll include
	state.inJekyllInclude = true
	if !state.skipLine("{% include alert.html %}") {
		t.Error("expected jekyll include line to be skipped")
	}
	state.inJekyllInclude = false

	// In HTML block
	state.inHTMLBlock = true
	if !state.skipLine("<div>some text</div>") {
		t.Error("expected html block line to be skipped")
	}
	state.inHTMLBlock = false
}

func TestContentStateUpdate(t *testing.T) {
	state := &contentState{}

	// Front matter toggle
	state.update("---")
	if !state.inFrontMatter {
		t.Error("expected inFrontMatter to be true after first ---")
	}
	state.update("---")
	if state.inFrontMatter {
		t.Error("expected inFrontMatter to be false after second ---")
	}

	// Code block toggle
	state.update("```go")
	if !state.inCodeBlock {
		t.Error("expected inCodeBlock to be true after ```go")
	}
	state.update("```")
	if state.inCodeBlock {
		t.Error("expected inCodeBlock to be false after closing ```")
	}

	// Jekyll include multiline
	state.update("{% include figure.html")
	if !state.inJekyllInclude {
		t.Error("expected inJekyllInclude to be true after open tag")
	}
	state.update("caption=\"test\" %}")
	if state.inJekyllInclude {
		t.Error("expected inJekyllInclude to be false after closing tag")
	}
}

func TestProcessLineTermLinking(t *testing.T) {
	termInfos := []TermInfo{
		{
			OriginalTerm: "Risk Assessment",
			Slug:         "risk-assessment",
			Regex:        regexp.MustCompile(`(?i)\bRisk Assessment\b`),
		},
	}

	defPath := "../model/02-definitions.html"

	input := "Every project requires a Risk Assessment before launch."
	got := processLine(input, termInfos, defPath)
	want := "Every project requires a [Risk Assessment](../model/02-definitions.html#risk-assessment) before launch."
	if got != want {
		t.Errorf("processLine() = %q, want %q", got, want)
	}

	// Do not link when already linked
	existing := "Every project requires a [Risk Assessment](../model/02-definitions.html#risk-assessment) before launch."
	gotExisting := processLine(existing, termInfos, defPath)
	if gotExisting != existing {
		t.Errorf("processLine() double linked = %q, want %q", gotExisting, existing)
	}

	// Do not link within inline code
	inlineCode := "The function `performRisk Assessment()` is called."
	gotCode := processLine(inlineCode, termInfos, defPath)
	if gotCode != inlineCode {
		t.Errorf("processLine() linked inside code = %q, want %q", gotCode, inlineCode)
	}
}

func TestProcessFileAndCleanupFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")

	initialContent := "This document discusses Risk Assessment in detail.\n"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	termInfos := []TermInfo{
		{
			OriginalTerm: "Risk Assessment",
			Slug:         "risk-assessment",
			Regex:        regexp.MustCompile(`(?i)\bRisk Assessment\b`),
		},
	}

	// Process file
	if err := processFile(testFile, termInfos, tempDir); err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	processed, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(processed), "[Risk Assessment](") {
		t.Fatalf("expected linked term in processed content, got:\n%s", string(processed))
	}

	// Cleanup file
	if err := cleanupFile(testFile, termInfos, tempDir); err != nil {
		t.Fatalf("cleanupFile failed: %v", err)
	}

	cleaned, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(cleaned), "[Risk Assessment](") {
		t.Fatalf("expected plain term after cleanup, got:\n%s", string(cleaned))
	}
	if !strings.Contains(string(cleaned), "Risk Assessment") {
		t.Fatalf("expected original text preserved, got:\n%s", string(cleaned))
	}
}

func TestFindMarkdownFiles(t *testing.T) {
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	_ = os.WriteFile(filepath.Join(tempDir, "one.md"), []byte("# One"), 0644)
	_ = os.WriteFile(filepath.Join(subDir, "two.md"), []byte("# Two"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "ignore.txt"), []byte("Ignore"), 0644)

	files, err := findMarkdownFiles(tempDir)
	if err != nil {
		t.Fatalf("findMarkdownFiles failed: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 markdown files, found %d", len(files))
	}
}

func TestFindHTMLContentRanges(t *testing.T) {
	// Empty tag ranges
	empty := findHTMLContentRanges("plain text", nil)
	if len(empty) != 0 {
		t.Errorf("expected 0 ranges for empty tagRanges, got %d", len(empty))
	}

	// Simple tags
	text := "<div>hello world</div>"
	tagRanges := [][]int{{0, 5}, {16, 22}}
	ranges := findHTMLContentRanges(text, tagRanges)
	if len(ranges) != 1 || ranges[0].start != 5 || ranges[0].end != 16 {
		t.Errorf("expected range (5, 16), got %+v", ranges)
	}

	// Nested tags
	nested := "<div><span>inner</span> outside</div>"
	// <div> is (0,5), <span> is (5,11), </span> is (16,23), </div> is (31,37)
	nestedTagRanges := [][]int{{0, 5}, {5, 11}, {16, 23}, {31, 37}}
	nestedRanges := findHTMLContentRanges(nested, nestedTagRanges)
	if len(nestedRanges) < 1 {
		t.Errorf("expected nested content ranges, got %+v", nestedRanges)
	}

	// Self-closing tags should not produce content range
	selfClosing := "<img src='test.png' />"
	scTagRanges := [][]int{{0, len(selfClosing)}}
	scRanges := findHTMLContentRanges(selfClosing, scTagRanges)
	if len(scRanges) != 0 {
		t.Errorf("expected 0 content ranges for self-closing tag, got %d", len(scRanges))
	}

	// Unclosed tag
	unclosed := "<article>some unclosed text"
	unclosedTagRanges := [][]int{{0, 9}}
	unclosedRanges := findHTMLContentRanges(unclosed, unclosedTagRanges)
	if len(unclosedRanges) != 1 || unclosedRanges[0].start != 9 || unclosedRanges[0].end != len(unclosed) {
		t.Errorf("expected range (9, %d) for unclosed tag, got %+v", len(unclosed), unclosedRanges)
	}
}
