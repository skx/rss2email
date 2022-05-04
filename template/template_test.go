package template

import "testing"

func TestTemplate(t *testing.T) {
	content := EmailTemplate()
	if len(content) != 2292 {
		t.Fatalf("unexpected template size 2265 != %d", len(content))
	}
}
