package analytics

import "testing"

func TestAcceptanceCountsBounces(t *testing.T) {
	// 90 sent, 10 bounced, 0 failed. The old formula (sent/(sent+failed))
	// would have reported 100%. Correct answer is 90%.
	if got := acceptancePct(90, 0, 10); got != 90 {
		t.Errorf("expected 90, got %v", got)
	}
	// A heavy-bounce campaign must not read as near-perfect.
	if got := acceptancePct(50, 0, 50); got != 50 {
		t.Errorf("expected 50, got %v", got)
	}
}

func TestAcceptanceExcludesSuppressed(t *testing.T) {
	// Suppressed addresses are never passed into the denominator, so a run
	// that sent 100 and suppressed some still reads 100% acceptance.
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
