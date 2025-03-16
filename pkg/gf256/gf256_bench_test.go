package gf256

import "testing"

func BenchmarkMultiplyDo(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MultiplyDo(0x57, 0x83)
	}
}

func BenchmarkMultiply(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Multiply(0x57, 0x83)
	}
}

func BenchmarkDivide(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Divide(0x57, 0x83)
	}
}

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Add(0x57, 0x83)
	}
}
