package template

import "testing"

func TestTemplate(t *testing.T) {
	content := EmailTemplate()
	if len(content) != 2062 {
		t.Fatalf("unexpected template size 2062 != %d", len(content))
	}
}
