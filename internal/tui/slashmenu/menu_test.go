package slashmenu

import "testing"

func TestFilterAndComplete(t *testing.T) {
	m := New()
	m.UpdateQuery("/ac")
	if !m.Active {
		t.Fatal("expected active")
	}
	if len(m.Filtered) == 0 {
		t.Fatal("expected matches")
	}
	// /act should rank high
	found := false
	for _, it := range m.Filtered {
		if it.Command == "/act" {
			found = true
		}
	}
	if !found {
		t.Fatalf("%v", m.Filtered)
	}
	m.Cursor = 0
	c := m.Complete()
	if c == "" || c[0] != '/' {
		t.Fatalf("%q", c)
	}
}

func TestInactiveWithoutSlash(t *testing.T) {
	m := New()
	m.UpdateQuery("hello")
	if m.Active {
		t.Fatal("should be inactive")
	}
}
