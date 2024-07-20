package datetime

import "testing"

func BenchmarkAmsterdam(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Amsterdam()
	}
}

func BenchmarkTomorrow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Tomorrow()
	}
}
