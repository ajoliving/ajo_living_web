package utils

import "testing"

func TestNormalizeAvatarKey(t *testing.T) {
	if got := NormalizeAvatarKey(" AVO-16 "); got != "avo-16" {
		t.Fatalf("NormalizeAvatarKey() = %q", got)
	}
	if got := NormalizeAvatarKey("avo-00"); got != "" {
		t.Fatalf("NormalizeAvatarKey(invalid) = %q", got)
	}
	if got := NormalizeAvatarKey("custom"); got != "" {
		t.Fatalf("NormalizeAvatarKey(custom) = %q", got)
	}
}

func TestAvatarKeyFromSeedIsStable(t *testing.T) {
	first := AvatarKeyFromSeed("01KUSER001")
	second := AvatarKeyFromSeed("01KUSER001")
	if first != second {
		t.Fatalf("AvatarKeyFromSeed() changed from %q to %q", first, second)
	}
	if NormalizeAvatarKey(first) == "" {
		t.Fatalf("AvatarKeyFromSeed() returned invalid key %q", first)
	}
	if AvatarKeyFromSeed("01KUSER002") == first {
		t.Fatalf("different seeds mapped to the same key %q", first)
	}
}

func TestResolveAndBuildingAvatarKey(t *testing.T) {
	if got := ResolveAvatarKey("avo-07", "seed"); got != "avo-07" {
		t.Fatalf("ResolveAvatarKey() = %q", got)
	}
	if got := ResolveAvatarKey("", "seed"); got != AvatarKeyFromSeed("seed") {
		t.Fatalf("ResolveAvatarKey(fallback) = %q", got)
	}
	if got := BuildingAvatarKey("0999900"); got != AvatarKeyFromSeed("building:0999900") {
		t.Fatalf("BuildingAvatarKey() = %q", got)
	}
}
