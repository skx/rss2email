package template

import "testing"

func TestTemplate(t *testing.T) {

	// content of our template
	content := EmailTemplate()

	// expected template length
	length := 2457

	// check the content is as big as it should be.
	if len(content) != length {
		t.Fatalf("unexpected template size %d != %d", length, len(content))
	}
}
