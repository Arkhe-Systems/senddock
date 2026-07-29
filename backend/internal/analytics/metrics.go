package analytics

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
