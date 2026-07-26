package backend

import (
	"testing"
	"time"
)

func BenchmarkConfigService_ReadHeavy(b *testing.B) {
	cs := NewConfigService()
	cs.Set("testKey", "testValue")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cs.Get("testKey")
			// simulate some tiny work
			time.Sleep(1 * time.Nanosecond)
		}
	})
}
