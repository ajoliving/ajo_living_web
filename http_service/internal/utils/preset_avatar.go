/*
 * Preset avatar helpers.
 * 1. Keep a fixed static avatar catalog identified by short keys.
 * 2. Resolve a stored key or derive a stable fallback from a seed.
 */
package utils

import (
	"fmt"
	"strings"
)

const PresetAvatarCount = 16

// 1. NormalizeAvatarKey returns a valid catalog key or an empty string.
func NormalizeAvatarKey(key string) string {
	value := strings.ToLower(strings.TrimSpace(key))
	if len(value) != 6 || !strings.HasPrefix(value, "avo-") {
		return ""
	}

	index := int(value[4]-'0')*10 + int(value[5]-'0')
	if index < 1 || index > PresetAvatarCount {
		return ""
	}

	return value
}

// 2. AvatarKeyFromSeed maps any identifier to a stable catalog key.
func AvatarKeyFromSeed(seed string) string {
	value := strings.TrimSpace(seed)
	if value == "" {
		return "avo-01"
	}

	var hash uint32
	for i := 0; i < len(value); i++ {
		hash = hash*33 + uint32(value[i])
	}
	index := int(hash%uint32(PresetAvatarCount)) + 1
	return fmt.Sprintf("avo-%02d", index)
}

// 3. ResolveAvatarKey prefers a stored catalog key, then a seed fallback.
func ResolveAvatarKey(saved string, seed string) string {
	if key := NormalizeAvatarKey(saved); key != "" {
		return key
	}
	return AvatarKeyFromSeed(seed)
}

// 4. BuildingAvatarKey returns the stable group avatar for one building.
func BuildingAvatarKey(buildingID string) string {
	value := strings.TrimSpace(buildingID)
	if value == "" {
		return AvatarKeyFromSeed("building")
	}
	return AvatarKeyFromSeed("building:" + value)
}
