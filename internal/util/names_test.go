package util

import (
	"regexp"
	"testing"
)

func TestRandomName(t *testing.T) {
	pattern := regexp.MustCompile(`^[a-z]+-[a-z]+-[0-9a-f]+$`)

	for i := 0; i < 10; i++ {
		name := RandomName()
		if !pattern.MatchString(name) {
			t.Errorf("RandomName() = %q, does not match expected format", name)
		}
	}
}
