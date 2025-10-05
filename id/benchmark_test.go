package id

import (
	"testing"
)

// Benchmark tests for comparing performance of different ID generators

func BenchmarkGenerate(b *testing.B) {
	for b.Loop() {
		_ = Generate()
	}
}

func BenchmarkGenerateUuid(b *testing.B) {
	for b.Loop() {
		_ = GenerateUuid()
	}
}

func BenchmarkSnowflakeIdGenerator(b *testing.B) {
	generator, err := NewSnowflakeIdGenerator(1)
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		_ = generator.Generate()
	}
}

func BenchmarkSnowflakeIdGenerator_Parallel(b *testing.B) {
	generator, err := NewSnowflakeIdGenerator(1)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = generator.Generate()
		}
	})
}

func BenchmarkXidIdGenerator(b *testing.B) {
	generator := NewXidIdGenerator()

	for b.Loop() {
		_ = generator.Generate()
	}
}

func BenchmarkXidIdGenerator_Parallel(b *testing.B) {
	generator := NewXidIdGenerator()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = generator.Generate()
		}
	})
}

func BenchmarkUuidIdGenerator(b *testing.B) {
	generator := NewUuidIdGenerator()

	for b.Loop() {
		_ = generator.Generate()
	}
}

func BenchmarkUuidIdGenerator_Parallel(b *testing.B) {
	generator := NewUuidIdGenerator()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = generator.Generate()
		}
	})
}

func BenchmarkRandomIdGenerator_Short(b *testing.B) {
	generator := NewRandomIdGenerator("0123456789abcdef", 8)

	for b.Loop() {
		_ = generator.Generate()
	}
}

func BenchmarkRandomIdGenerator_Medium(b *testing.B) {
	generator := NewRandomIdGenerator("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", 21)

	for b.Loop() {
		_ = generator.Generate()
	}
}

func BenchmarkRandomIdGenerator_Long(b *testing.B) {
	generator := NewRandomIdGenerator("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-", 64)

	for b.Loop() {
		_ = generator.Generate()
	}
}

func BenchmarkRandomIdGenerator_Parallel(b *testing.B) {
	generator := NewRandomIdGenerator("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", 21)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = generator.Generate()
		}
	})
}

// Benchmark default generators for comparison.
func BenchmarkDefaultGenerators(b *testing.B) {
	b.Run("DefaultXidIdGenerator", func(b *testing.B) {
		b.ResetTimer()

		for b.Loop() {
			_ = DefaultXidIdGenerator.Generate()
		}
	})

	b.Run("DefaultUuidIdGenerator", func(b *testing.B) {
		b.ResetTimer()

		for b.Loop() {
			_ = DefaultUuidIdGenerator.Generate()
		}
	})

	b.Run("DefaultSnowflakeIdGenerator", func(b *testing.B) {
		b.ResetTimer()

		for b.Loop() {
			_ = DefaultSnowflakeIdGenerator.Generate()
		}
	})
}

// Memory allocation benchmarks.
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("Snowflake", func(b *testing.B) {
		generator, _ := NewSnowflakeIdGenerator(1)

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			_ = generator.Generate()
		}
	})

	b.Run("XID", func(b *testing.B) {
		generator := NewXidIdGenerator()

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			_ = generator.Generate()
		}
	})

	b.Run("UUID", func(b *testing.B) {
		generator := NewUuidIdGenerator()

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			_ = generator.Generate()
		}
	})

	b.Run("Random", func(b *testing.B) {
		generator := NewRandomIdGenerator("0123456789abcdefghijklmnopqrstuvwxyz", 21)

		b.ResetTimer()
		b.ReportAllocs()

		for b.Loop() {
			_ = generator.Generate()
		}
	})
}

// Concurrent performance comparison.
func BenchmarkConcurrentPerformance(b *testing.B) {
	b.Run("Snowflake_Concurrent", func(b *testing.B) {
		generator, _ := NewSnowflakeIdGenerator(1)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = generator.Generate()
			}
		})
	})

	b.Run("XID_Concurrent", func(b *testing.B) {
		generator := NewXidIdGenerator()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = generator.Generate()
			}
		})
	})

	b.Run("UUID_Concurrent", func(b *testing.B) {
		generator := NewUuidIdGenerator()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = generator.Generate()
			}
		})
	})

	b.Run("Random_Concurrent", func(b *testing.B) {
		generator := NewRandomIdGenerator("0123456789abcdefghijklmnopqrstuvwxyz", 21)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = generator.Generate()
			}
		})
	})
}
