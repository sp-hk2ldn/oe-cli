package cli

import "testing"

func TestSafeDisplayURL(t *testing.T) {
	t.Parallel()

	raw := "https://example.apple.com/path/to/report.csv?token=abc123&sig=zzz#frag"
	got := safeDisplayURL(raw)
	want := "https://example.apple.com/path/to/report.csv"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}

	if safeDisplayURL("not-a-url") != "not-a-url" {
		t.Fatalf("expected non-url input to pass through")
	}
}
