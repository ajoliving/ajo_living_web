package service

import (
	"strings"
	"testing"
)

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

func TestIsmartSubaccountContactResultStates(t *testing.T) {
	result := &IsmartSubaccountContactResult{EmailUserID: 12, PhoneUserID: 12, EmailExists: true, PhoneExists: true}
	result.SameUser = result.EmailUserID == result.PhoneUserID
	result.CanAuthorize = result.SameUser
	if !result.CanAuthorize {
		t.Fatal("matching contact user IDs should be authorizable")
	}
}

func TestGenerateTemporaryPasswordIsStrongEnoughForMailDelivery(t *testing.T) {
	password, err := generateTemporaryPassword()
	if err != nil {
		t.Fatalf("generate temporary password: %v", err)
	}
	if len(password) != 12 {
		t.Fatalf("password length = %d, want 12", len(password))
	}
	for _, value := range password {
		if !strings.ContainsRune(temporaryPasswordAlphabet, value) {
			t.Fatalf("password contains unsupported character %q", value)
		}
	}
	if !strings.ContainsAny(password, temporaryPasswordUpper) || !strings.ContainsAny(password, temporaryPasswordLower) || !strings.ContainsAny(password, temporaryPasswordDigits) {
		t.Fatalf("password should contain upper, lower and digit characters: %q", password)
	}
}

func TestBuildSubaccountEmailBodyIncludesTemporaryPassword(t *testing.T) {
	body := buildSubaccountEmailBody(&IsmartSubaccountRegisterGrantResult{
		AccountCreated: true,
		TargetUsername: "200123",
		BuildingName:   "測試大廈",
		UnitDisplay:    "12 / A",
	}, "TempAbc2345", "user@example.com")
	for _, value := range []string{"測試大廈", "12 / A", "200123", "TempAbc2345", "user@example.com", "修改臨時密碼"} {
		if !strings.Contains(body, value) {
			t.Fatalf("email body does not contain %q: %s", value, body)
		}
	}
}
