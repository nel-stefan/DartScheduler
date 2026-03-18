package domain_test

import (
	"testing"

	"DartScheduler/domain"
)

func TestFormatDisplayName_Standard(t *testing.T) {
	got := domain.FormatDisplayName("Jansen, Jan")
	if got != "Jan Jansen" {
		t.Errorf("got %q, want %q", got, "Jan Jansen")
	}
}

func TestFormatDisplayName_LeadingTrailingSpacesInParts(t *testing.T) {
	// TrimSpace is applied to each part after split on ", "
	got := domain.FormatDisplayName("Jansen,  Jan ")
	// splits on first ", " → parts[0]="Jansen", parts[1]=" Jan " → trimmed → "Jan Jansen"
	if got != "Jan Jansen" {
		t.Errorf("got %q, want %q", got, "Jan Jansen")
	}
}

func TestFormatDisplayName_NoComma(t *testing.T) {
	got := domain.FormatDisplayName("Pietersen")
	if got != "Pietersen" {
		t.Errorf("got %q, want %q", got, "Pietersen")
	}
}

func TestFormatDisplayName_Empty(t *testing.T) {
	got := domain.FormatDisplayName("")
	if got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

func TestFormatDisplayName_MultipleCommas(t *testing.T) {
	// Only the first ", " is used as separator; rest becomes part of the first-name segment.
	got := domain.FormatDisplayName("Bakker, Jan, Jr.")
	if got != "Jan, Jr. Bakker" {
		t.Errorf("got %q, want %q", got, "Jan, Jr. Bakker")
	}
}

func TestFormatDisplayName_FullName(t *testing.T) {
	got := domain.FormatDisplayName("van den Berg, Pieter")
	if got != "Pieter van den Berg" {
		t.Errorf("got %q, want %q", got, "Pieter van den Berg")
	}
}
