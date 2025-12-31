package ui

import (
	"testing"
)

func TestRedText(t *testing.T) {
	result := RedText("test")
	if result == "" {
		t.Error("RedText() returned empty string")
	}
}

func TestYellowText(t *testing.T) {
	result := YellowText("test")
	if result == "" {
		t.Error("YellowText() returned empty string")
	}
}

func TestGreenText(t *testing.T) {
	result := GreenText("test")
	if result == "" {
		t.Error("GreenText() returned empty string")
	}
}

func TestCyanText(t *testing.T) {
	result := CyanText("test")
	if result == "" {
		t.Error("CyanText() returned empty string")
	}
}

func TestBoldText(t *testing.T) {
	result := BoldText("test")
	if result == "" {
		t.Error("BoldText() returned empty string")
	}
}

func TestCommand(t *testing.T) {
	result := Command("muxi deploy")
	if result == "" {
		t.Error("Command() returned empty string")
	}
}

func TestDimmedText(t *testing.T) {
	result := DimmedText("test")
	if result == "" {
		t.Error("DimmedText() returned empty string")
	}
}

func TestGoldText(t *testing.T) {
	result := GoldText("test")
	if result == "" {
		t.Error("GoldText() returned empty string")
	}
}

func TestIndent(t *testing.T) {
	result := Indent("test", 2)
	if result == "" {
		t.Error("Indent() returned empty string")
	}
}

func TestIndentString(t *testing.T) {
	result := IndentString("line1\nline2", 2)
	if result == "" {
		t.Error("IndentString() returned empty string")
	}
}

func TestRenderYAML(t *testing.T) {
	yaml := `key: value
nested:
  foo: bar`
	result := RenderYAML(yaml)
	if result == "" {
		t.Error("RenderYAML() returned empty string")
	}
}

func TestRenderMarkdown(t *testing.T) {
	md := "# Header\n\nSome **bold** text"
	result := RenderMarkdown(md)
	if result == "" {
		t.Error("RenderMarkdown() returned empty string")
	}
}

func TestSuccess(t *testing.T) {
	// Just verify no panic
	Success("test message")
}

func TestStep(t *testing.T) {
	Step("test step")
}

func TestError(t *testing.T) {
	Error("test error")
}

func TestWarning(t *testing.T) {
	Warning("test warning")
}

func TestInfo(t *testing.T) {
	Info("test info")
}

func TestSkipped(t *testing.T) {
	Skipped("test skipped")
}

func TestDimmed(t *testing.T) {
	Dimmed("test dimmed")
}

func TestGold(t *testing.T) {
	Gold("test gold")
}

func TestBold(t *testing.T) {
	Bold("test bold")
}

func TestSection(t *testing.T) {
	Section("Test Section")
}

func TestList(t *testing.T) {
	List([]string{"item1", "item2", "item3"})
}

func TestStatusList(t *testing.T) {
	items := []StatusItem{
		{Text: "Test", Status: "success"},
		{Text: "Test2", Status: "warning", Detail: "some detail"},
	}
	StatusList(items)
}

func TestErrorBlock(t *testing.T) {
	ErrorBlock("Error Title", "Error details here", "Try this fix")
}

func TestSuccessBlock(t *testing.T) {
	SuccessBlock("Success!", "Next step: do something")
}

func TestProgressStep(t *testing.T) {
	ProgressStep(1, 5, "Processing...")
}

func TestBanner(t *testing.T) {
	Banner("Test Banner")
}

func TestInfoBanner(t *testing.T) {
	InfoBanner("Test Info Banner")
}

func TestNewSpinner(t *testing.T) {
	s := NewSpinner("Loading...")
	if s == nil {
		t.Error("NewSpinner() returned nil")
	}
}

func TestNewSpinnerWithPadding(t *testing.T) {
	s := NewSpinnerWithPadding("Loading...", 2)
	if s == nil {
		t.Error("NewSpinnerWithPadding() returned nil")
	}
}
