package backend

import (
	"runtime"
	"testing"
)

func BenchmarkListAndExists(b *testing.B) {
	if runtime.GOOS != "windows" {
		b.Skip("requires windows")
	}

	cfg := NewConfigService()
	svc := NewDdevService(cfg)

	// Prime cache
	svc.ListWSLDistros()
	distroExists("DDEV")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.ListWSLDistros()
		distroExists("DDEV")
	}
}
