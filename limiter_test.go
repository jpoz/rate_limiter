package rate

import (
	"testing"
)

func BenchmarkInternal(b *testing.B) {
	limiter := NewInternalLimiter(2500)

	for i := 0; i < b.N; i++ {
		limiter.Allow()
	}
}
