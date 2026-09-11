package service

import "testing"

func TestIsmartSubaccountPhoneCandidatesIncludesLegacyStoredFormats(t *testing.T) {
	got := ismartSubaccountPhoneCandidates("+8615666823185")
	if len(got) == 0 || got[0] != "+8615666823185" {
		t.Fatalf("first candidate = %v, want submitted phone first", got)
	}

	seen := map[string]bool{}
	for _, item := range got {
		if seen[item] {
			t.Fatalf("duplicate candidate %q in %v", item, got)
		}
		seen[item] = true
	}
	for _, item := range []string{"+8615666823185", "15666823185", "+15666823185", "8615666823185"} {
		if !seen[item] {
			t.Fatalf("missing candidate %q in %v", item, got)
		}
	}
}

func TestIsmartSubaccountNationalNumberStripsKnownPrefixes(t *testing.T) {
	if got := ismartSubaccountNationalNumber("8615666823185"); got != "15666823185" {
		t.Fatalf("cn prefix = %q", got)
	}
	if got := ismartSubaccountNationalNumber("15666823185"); got != "15666823185" {
		t.Fatalf("cn mobile = %q", got)
	}
	if got := ismartSubaccountNationalNumber("115666823185"); got != "15666823185" {
		t.Fatalf("plus-one prefix = %q", got)
	}
}
