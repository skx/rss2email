package template

import "testing"

func TestTemplate(t *testing.T) {

	// content and expected length
	content := EmailTemplate()
	length := 2645

	if len(content) != length {
		t.Fatalf("unexpected template size %d != %d", length, len(content))
	}
}
