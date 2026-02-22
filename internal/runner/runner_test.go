package runner

import "testing"

func TestAllowlist(t *testing.T) {
	if IsWithinRoots("/tmp", []string{"/Users/none"}) {
		t.Fatal("unexpected allow")
	}
}

func TestJSONParse(t *testing.T) {
	if _, err := ParseOneJSONObject("{\"a\":1}"); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseOneJSONObject("[]"); err == nil {
		t.Fatal("expected error")
	}
}
