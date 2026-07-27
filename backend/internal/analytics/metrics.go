package analytics

// The email log reaches one terminal status: sent (accepted by the relay),
// bounced (rejected, inline or async), failed (send error), or suppressed
// (never attempted because the address was on the suppression list).
//
// Acceptance and bounce rate share a denominator of *attempted* messages —
// sent + failed + bounced — so suppressed addresses, which were never sent, do
// not distort them. This is the fix for the old formula, which used only
// sent / (sent + failed) and so ignored bounces entirely.

func attempted(sent, failed, bounced int64) int64 {
	return sent + failed + bounced
}

func acceptancePct(sent, failed, bounced int64) float64 {
	return ratePct(sent, attempted(sent, failed, bounced))
}

func bounceRatePct(sent, failed, bounced int64) float64 {
	return ratePct(bounced, attempted(sent, failed, bounced))
}

func ratePct(numerator, denominator int64) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator) * 100
}
