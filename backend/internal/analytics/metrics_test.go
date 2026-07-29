package analytics

import "testing"

func TestAcceptanceCountsBounces(t *testing.T) {

	if got := acceptancePct(90, 0, 10); got != 90 {
		t.Errorf("expected 90, got %v", got)
	}

	if got := acceptancePct(50, 0, 50); got != 50 {
		t.Errorf("expected 50, got %v", got)
	}
}

func TestAcceptanceExcludesSuppressed(t *testing.T) {

	if got := acceptancePct(100, 0, 0); got != 100 {
		t.Errorf("expected 100, got %v", got)
	}
}

func TestBounceRate(t *testing.T) {
	if got := bounceRatePct(90, 0, 10); got != 10 {
		t.Errorf("expected 10, got %v", got)
	}
}

func TestRatesHandleZero(t *testing.T) {
	if got := acceptancePct(0, 0, 0); got != 0 {
		t.Errorf("no attempts must be 0, got %v", got)
	}
	if got := ratePct(5, 0); got != 0 {
		t.Errorf("zero denominator must be 0, got %v", got)
	}
}
