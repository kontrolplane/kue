package styles

import (
	"testing"
)

func TestTableStylesDoNotPanic(t *testing.T) {
	// Ensure TableStyles() doesn't panic and returns valid styles
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("TableStyles() panicked: %v", r)
		}
	}()

	s := TableStyles()
	if s.Header.GetBold() {
		t.Error("Header should not be bold by default")
	}
}

func TestAttributesTableStylesDoNotPanic(t *testing.T) {
	// Ensure AttributesTableStyles() doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AttributesTableStyles() panicked: %v", r)
		}
	}()

	s := AttributesTableStyles()
	if s.Selected.GetBold() {
		t.Error("Selected should not be bold for attributes table")
	}
}

func TestFormThemeDoesNotPanic(t *testing.T) {
	// Ensure FormTheme() doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("FormTheme() panicked: %v", r)
		}
	}()

	theme := FormTheme()
	if theme == nil {
		t.Error("FormTheme() should not return nil")
	}
}

func TestButtonStylesRender(t *testing.T) {
	// Ensure button styles can render without panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Button styles panicked: %v", r)
		}
	}()

	primary := ButtonPrimary.Render("test")
	if primary == "" {
		t.Error("ButtonPrimary should render non-empty string")
	}

	secondary := ButtonSecondary.Render("test")
	if secondary == "" {
		t.Error("ButtonSecondary should render non-empty string")
	}
}

func TestContentWrapperCreatesValidStyle(t *testing.T) {
	style := ContentWrapper(100, 50)
	if style.GetWidth() != 100 {
		t.Errorf("Expected width 100, got %d", style.GetWidth())
	}
	if style.GetHeight() != 50 {
		t.Errorf("Expected height 50, got %d", style.GetHeight())
	}
}

func TestCenteredFormCreatesValidStyle(t *testing.T) {
	style := CenteredForm(100)
	// Width should be 100 - 4 = 96
	if style.GetWidth() != 96 {
		t.Errorf("Expected width 96, got %d", style.GetWidth())
	}
}
