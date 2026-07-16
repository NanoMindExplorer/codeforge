package headless

import "testing"

func TestTrimUniq(t *testing.T) {
	if trim("hello world", 5) != "hello…" {
		t.Fatal(trim("hello world", 5))
	}
	u := uniq([]string{"a", "a", "b"})
	if len(u) != 2 {
		t.Fatal(u)
	}
}
