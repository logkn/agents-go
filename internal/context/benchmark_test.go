package context

import (
	"sync"
	"testing"
)

// Complex context type for benchmarking
type BenchmarkContext struct {
	UserID      string
	SessionID   string
	Permissions []string
	Metadata    map[string]any
	Config      struct {
		Theme    string
		Language string
		Features map[string]bool
	}
}

func createComplexContext() BenchmarkContext {
	return BenchmarkContext{
		UserID:      "user123456789",
		SessionID:   "session987654321",
		Permissions: []string{"read", "write", "admin", "delete", "create"},
		Metadata: map[string]any{
			"last_login":  "2024-01-01T00:00:00Z",
			"login_count": 42,
			"preferences": map[string]string{"theme": "dark", "lang": "en"},
			"experiments": []string{"feature_a", "feature_b", "feature_c"},
		},
		Config: struct {
			Theme    string
			Language string
			Features map[string]bool
		}{
			Theme:    "dark",
			Language: "en-US",
			Features: map[string]bool{
				"notifications": true,
				"analytics":     false,
				"beta_features": true,
				"advanced_ui":   true,
			},
		},
	}
}

func BenchmarkContextCreation(b *testing.B) {
	data := createComplexContext()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewContext(data)
	}
}

func BenchmarkToAnyContext(b *testing.B) {
	data := createComplexContext()
	ctx := NewContext(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ToAnyContext(ctx)
	}
}

func BenchmarkFromAnyContext(b *testing.B) {
	data := createComplexContext()
	ctx := NewContext(data)
	anyCtx := ToAnyContext(ctx)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FromAnyContext[BenchmarkContext](anyCtx)
	}
}

func BenchmarkContextChainFind(b *testing.B) {
	// Create a chain with multiple contexts
	chain := NewContextChain(
		ToAnyContext(NewContext("string context")),
		ToAnyContext(NewContext(123)),
		ToAnyContext(NewContext(createComplexContext())),
		ToAnyContext(NewContext(SimpleContext{Value: "simple"})),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Find[BenchmarkContext](chain)
	}
}

func BenchmarkCompositeContextOperations(b *testing.B) {
	b.Run("Add", func(b *testing.B) {
		data := createComplexContext()
		ctx := NewContext(data)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cc := NewCompositeContext()
			AddTyped(cc, ctx)
		}
	})

	b.Run("Get", func(b *testing.B) {
		cc := NewCompositeContext()
		data := createComplexContext()
		ctx := NewContext(data)
		AddTyped(cc, ctx)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = Get[BenchmarkContext](cc)
		}
	})

	b.Run("Has", func(b *testing.B) {
		cc := NewCompositeContext()
		data := createComplexContext()
		ctx := NewContext(data)
		AddTyped(cc, ctx)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Has[BenchmarkContext](cc)
		}
	})
}

func BenchmarkThreadSafeContext(b *testing.B) {
	data := createComplexContext()
	ctx := NewContext(data)
	tsCtx := NewThreadSafeContext(ctx)

	b.Run("Read", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = tsCtx.Value()
			}
		})
	})

	b.Run("Write", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				newCtx := NewContext(createComplexContext())
				tsCtx.Update(newCtx)
			}
		})
	})

	b.Run("ReadWrite", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if pb.Next() {
					_ = tsCtx.Value() // Read
				} else {
					newCtx := NewContext(createComplexContext())
					tsCtx.Update(newCtx) // Write
				}
			}
		})
	})
}

func BenchmarkConcurrentAccess(b *testing.B) {
	data := createComplexContext()
	ctx := NewContext(data)
	anyCtx := ToAnyContext(ctx)

	b.Run("FromAnyContext_Parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = FromAnyContext[BenchmarkContext](anyCtx)
			}
		})
	})

	b.Run("CompositeContext_Parallel", func(b *testing.B) {
		cc := NewCompositeContext()
		AddTyped(cc, ctx)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = Get[BenchmarkContext](cc)
			}
		})
	})
}

func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("BasicContext", func(b *testing.B) {
		data := createComplexContext()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			ctx := NewContext(data)
			anyCtx := ToAnyContext(ctx)
			_, _ = FromAnyContext[BenchmarkContext](anyCtx)
		}
	})

	b.Run("CompositeContext", func(b *testing.B) {
		data := createComplexContext()
		ctx := NewContext(data)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			cc := NewCompositeContext()
			AddTyped(cc, ctx)
			_, _ = Get[BenchmarkContext](cc)
		}
	})

	b.Run("ContextChain", func(b *testing.B) {
		data := createComplexContext()
		ctx := NewContext(data)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			chain := NewContextChain(ToAnyContext(ctx))
			_, _ = Find[BenchmarkContext](chain)
		}
	})
}

// Benchmark comparison: context vs non-context tool execution
func BenchmarkToolExecution(b *testing.B) {
	type MockTool struct {
		Value string
	}

	tool := MockTool{Value: "test"}
	data := createComplexContext()
	ctx := NewContext(data)
	anyCtx := ToAnyContext(ctx)

	b.Run("WithoutContext", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate tool execution without context
			_ = tool.Value + " processed"
		}
	})

	b.Run("WithContext", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate tool execution with context access
			if recoveredCtx, err := FromAnyContext[BenchmarkContext](anyCtx); err == nil {
				contextData := recoveredCtx.Value()
				_ = tool.Value + " processed for user " + contextData.UserID
			}
		}
	})
}

// Stress test with many concurrent operations
func BenchmarkStressTest(b *testing.B) {
	const numGoroutines = 100
	const opsPerGoroutine = 1000

	data := createComplexContext()
	ctx := NewContext(data)
	tsCtx := NewThreadSafeContext(ctx)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup

		// Spawn many goroutines performing different operations
		for g := 0; g < numGoroutines; g++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				for op := 0; op < opsPerGoroutine; op++ {
					switch op % 4 {
					case 0:
						// Read operation
						_ = tsCtx.Value()
					case 1:
						// Write operation
						newCtx := NewContext(createComplexContext())
						tsCtx.Update(newCtx)
					case 2:
						// Composite context operation
						cc := NewCompositeContext()
						AddTyped(cc, ctx)
						_, _ = Get[BenchmarkContext](cc)
					case 3:
						// Chain operation
						chain := NewContextChain(ToAnyContext(ctx))
						_, _ = Find[BenchmarkContext](chain)
					}
				}
			}(g)
		}

		wg.Wait()
	}
}
