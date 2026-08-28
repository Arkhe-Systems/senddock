package service

import (
	"strings"
	"testing"
)

func TestNormalizeColumns_NoColumnsUntouched(t *testing.T) {
	in := `<div style="width:100%"><p>plain email</p></div>`
	if got := normalizeColumns(in); got != in {
		t.Fatalf("expected input returned verbatim, got %q", got)
	}
}

func TestNormalizeColumns_ConvertsDivColumnsToTable(t *testing.T) {
	in := `<div data-columns="2" style="width:100%;display: table;">` +
		`<div style="display: table-cell; width:50%;">Left</div>` +
		`<div style="display: table-cell; width:50%;">Right</div>` +
		`</div>`
	got := normalizeColumns(in)

	for _, want := range []string{
		`<table role="presentation" cellpadding="0" cellspacing="0"`,
		`style="width:100%;"`,
		`<td style="width:50%;">Left</td>`,
		`<td style="width:50%;">Right</td>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
	for _, bad := range []string{
		"data-columns",
		"display: table",
		"display:table",
		"display: table-cell",
		"display:table-cell",
	} {
		if strings.Contains(got, bad) {
			t.Errorf("unexpected %q in %q", bad, got)
		}
	}
}

func TestNormalizeColumns_StripsTableDisplayWithoutSpace(t *testing.T) {
	in := `<div data-columns="2" style="display:table;background:#fff;">` +
		`<div style="display:table-cell;">A</div>` +
		`<div style="display:table-cell;">B</div>` +
		`</div>`
	got := normalizeColumns(in)

	if strings.Contains(got, "display:table") {
		t.Errorf("display:table variant survived: %q", got)
	}
	if !strings.Contains(got, `style="background:#fff;"`) {
		t.Errorf("wrapper background lost: %q", got)
	}
	if !strings.Contains(got, `<td>A</td>`) || !strings.Contains(got, `<td>B</td>`) {
		t.Errorf("cells missing: %q", got)
	}
}

func TestNormalizeColumns_PreservesInnerMarkup(t *testing.T) {
	in := `<div data-columns="2" style="display: table;">` +
		`<div style="display: table-cell;"><h2>Title</h2><a href="https://x.example">go</a></div>` +
		`</div>`
	got := normalizeColumns(in)

	if !strings.Contains(got, `<h2>Title</h2>`) {
		t.Errorf("nested heading lost: %q", got)
	}
	if !strings.Contains(got, `href="https://x.example"`) {
		t.Errorf("nested link lost: %q", got)
	}
}

func TestNormalizeColumns_MultipleWrappers(t *testing.T) {
	in := `<div data-columns="2" style="display: table;"><div style="display: table-cell;">A</div></div>` +
		`<div data-columns="3" style="display: table;"><div style="display: table-cell;">B</div></div>`
	got := normalizeColumns(in)

	if n := strings.Count(got, `role="presentation"`); n != 2 {
		t.Errorf("expected 2 tables, got %d in %q", n, got)
	}
}

func TestNormalizeColumns_WrapperWithoutStyle(t *testing.T) {
	in := `<div data-columns="2"><div style="display: table-cell;">A</div></div>`
	got := normalizeColumns(in)

	if !strings.Contains(got, `<table role="presentation" cellpadding="0" cellspacing="0">`) {
		t.Errorf("styleless wrapper missing table: %q", got)
	}
	if !strings.Contains(got, `<td>A</td>`) {
		t.Errorf("styleless wrapper missing cell: %q", got)
	}
	if strings.Contains(got, `style="`) {
		t.Errorf("styleless wrapper should produce no style attribute: %q", got)
	}
}

func TestInlineCSS_InlinesStyleBlock(t *testing.T) {
	in := `<html><head><style>.x{color:red}</style></head><body><div class="x">hi</div></body></html>`
	got := inlineCSS(in)

	if strings.Contains(strings.ToLower(got), "<style") {
		t.Fatalf("style block not removed: %q", got)
	}
	if !strings.Contains(got, "color") {
		t.Fatalf("expected color to be inlined, got %q", got)
	}
}

func TestInlineCSS_PassthroughOnEmpty(t *testing.T) {
	in := `<div>no styles here</div>`
	if got := inlineCSS(in); !strings.Contains(got, "no styles here") {
		t.Fatalf("expected content to survive, got %q", got)
	}
}
