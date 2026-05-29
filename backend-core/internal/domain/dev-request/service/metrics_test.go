package service

import (
	"testing"
)

func TestMetrics_ObserveHelpers(t *testing.T) {
	// 1. Initialize
	InitMetrics()
	
	// Double initialization guard (sync.Once check)
	InitMetrics()

	// 2. Observe helpers
	ObserveExpiringSoon(5)
	ObserveStale(12)
	
	// Positive count check
	ObserveAutoRevoked(3)
	
	// Negative/Zero count branch check
	ObserveAutoRevoked(0)
	ObserveAutoRevoked(-1)
}
