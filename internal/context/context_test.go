package context

import (
	"reflect"
	"testing"
)

// Test types for various scenarios
type SimpleContext struct {
	Value string
}

type ComplexContext struct {
	ID       int
	Name     string
	Settings map[string]any
}

type PointerContext struct {
	Data *string
}

func TestNewContext(t *testing.T) {
	tests := []struct {
		name string
		data any
		want any
	}{
		{
			name: "simple string context",
			data: "test value",
			want: "test value",
		},
		{
			name: "struct context",
			data: SimpleContext{Value: "test"},
			want: SimpleContext{Value: "test"},
		},
		{
			name: "complex struct context",
			data: ComplexContext{
				ID:       123,
				Name:     "test",
				Settings: map[string]any{"key": "value"},
			},
			want: ComplexContext{
				ID:       123,
				Name:     "test",
				Settings: map[string]any{"key": "value"},
			},
		},
		{
			name: "nil pointer context",
			data: (*string)(nil),
			want: (*string)(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewContext(tt.data)
			if ctx == nil {
				t.Fatal("NewContext returned nil")
			}

			got := ctx.Value()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Value() = %v, want %v", got, tt.want)
			}

			if ctx.Type() == nil {
				t.Error("Type() returned nil")
			}
		})
	}
}

func TestEmptyContext(t *testing.T) {
	ctx := EmptyContext()
	if ctx == nil {
		t.Fatal("EmptyContext returned nil")
	}

	val := ctx.Value()
	// val should be of type NoContext (not an interface)
	if reflect.TypeOf(val) != reflect.TypeOf(NoContext{}) {
		t.Errorf("EmptyContext.Value() = %T, want NoContext", val)
	}
}

func TestToAnyContext(t *testing.T) {
	tests := []struct {
		name         string
		ctx          any
		wantTypeName string
		wantIsNil    bool
	}{
		{
			name:         "string context",
			ctx:          NewContext("test"),
			wantTypeName: "string",
			wantIsNil:    false,
		},
		{
			name:         "struct context",
			ctx:          NewContext(SimpleContext{Value: "test"}),
			wantTypeName: "context.SimpleContext",
			wantIsNil:    false,
		},
		{
			name:         "pointer context with value",
			ctx:          NewContext(&SimpleContext{Value: "test"}),
			wantTypeName: "*context.SimpleContext",
			wantIsNil:    false,
		},
		{
			name:         "nil pointer context",
			ctx:          NewContext((*SimpleContext)(nil)),
			wantTypeName: "*context.SimpleContext",
			wantIsNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var anyCtx AnyContext
			switch v := tt.ctx.(type) {
			case Context[string]:
				anyCtx = ToAnyContext(v)
			case Context[SimpleContext]:
				anyCtx = ToAnyContext(v)
			case Context[*SimpleContext]:
				anyCtx = ToAnyContext(v)
			}

			if anyCtx == nil {
				t.Fatal("ToAnyContext returned nil")
			}

			if got := anyCtx.TypeName(); got != tt.wantTypeName {
				t.Errorf("TypeName() = %v, want %v", got, tt.wantTypeName)
			}

			if got := anyCtx.IsNil(); got != tt.wantIsNil {
				t.Errorf("IsNil() = %v, want %v", got, tt.wantIsNil)
			}
		})
	}
}

func TestFromAnyContext(t *testing.T) {
	// Success cases
	t.Run("successful conversion", func(t *testing.T) {
		original := SimpleContext{Value: "test"}
		ctx := NewContext(original)
		anyCtx := ToAnyContext(ctx)

		recovered, err := FromAnyContext[SimpleContext](anyCtx)
		if err != nil {
			t.Fatalf("FromAnyContext failed: %v", err)
		}

		if recovered == nil {
			t.Fatal("FromAnyContext returned nil context")
		}

		if got := recovered.Value(); !reflect.DeepEqual(got, original) {
			t.Errorf("Recovered value = %v, want %v", got, original)
		}
	})

	// Error cases
	t.Run("nil anyContext", func(t *testing.T) {
		_, err := FromAnyContext[SimpleContext](nil)
		if err == nil {
			t.Fatal("FromAnyContext with nil should return error")
		}

		contextErr, ok := err.(*ContextError)
		if !ok {
			t.Fatalf("Expected ContextError, got %T", err)
		}

		if contextErr.Op != "FromAnyContext" {
			t.Errorf("Expected Op = FromAnyContext, got %s", contextErr.Op)
		}

		if contextErr.Err == nil || contextErr.Err.Error() != "anyContext is nil" {
			t.Errorf("Unexpected underlying error: %v", contextErr.Err)
		}
	})

	t.Run("type mismatch", func(t *testing.T) {
		ctx := NewContext("string value")
		anyCtx := ToAnyContext(ctx)

		_, err := FromAnyContext[SimpleContext](anyCtx)
		if err == nil {
			t.Fatal("FromAnyContext with type mismatch should return error")
		}

		contextErr, ok := err.(*ContextError)
		if !ok {
			t.Fatalf("Expected ContextError, got %T", err)
		}

		if contextErr.Op != "FromAnyContext" {
			t.Errorf("Expected Op = FromAnyContext, got %s", contextErr.Op)
		}

		if contextErr.Expected != "context.SimpleContext" {
			t.Errorf("Expected type = context.SimpleContext, got %s", contextErr.Expected)
		}

		if contextErr.Got != "string" {
			t.Errorf("Got type = string, got %s", contextErr.Got)
		}
	})
}

func TestContextWrapper_EdgeCases(t *testing.T) {
	t.Run("nil context handling", func(t *testing.T) {
		var ctx Context[string]
		anyCtx := ToAnyContext(ctx)

		if !anyCtx.IsNil() {
			t.Error("IsNil() should return true for nil context")
		}
	})

	t.Run("empty struct context", func(t *testing.T) {
		type EmptyStruct struct{}
		ctx := NewContext(EmptyStruct{})
		anyCtx := ToAnyContext(ctx)

		if anyCtx.IsNil() {
			t.Error("IsNil() should return false for empty struct")
		}

		recovered, err := FromAnyContext[EmptyStruct](anyCtx)
		if err != nil {
			t.Fatalf("Failed to recover empty struct: %v", err)
		}

		if !reflect.DeepEqual(recovered.Value(), EmptyStruct{}) {
			t.Error("Failed to recover empty struct value")
		}
	})

	t.Run("interface context", func(t *testing.T) {
		var val any = "test"
		ctx := NewContext(val)
		anyCtx := ToAnyContext(ctx)

		if anyCtx.TypeName() != "interface {}" {
			t.Errorf("TypeName() = %v, want interface {}", anyCtx.TypeName())
		}

		recovered, err := FromAnyContext[any](anyCtx)
		if err != nil {
			t.Fatalf("Failed to recover interface: %v", err)
		}

		if recovered.Value() != val {
			t.Errorf("Recovered value = %v, want %v", recovered.Value(), val)
		}
	})
}

func TestContextFactory(t *testing.T) {
	factory := func() SimpleContext {
		return SimpleContext{Value: "factory created"}
	}

	// Create context using factory
	ctx := NewContext(factory())

	if ctx.Value().Value != "factory created" {
		t.Errorf("Factory created unexpected value: %v", ctx.Value())
	}
}

// Basic benchmark test
func BenchmarkNewContext(b *testing.B) {
	data := ComplexContext{
		ID:       123,
		Name:     "benchmark",
		Settings: map[string]any{"key": "value"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewContext(data)
	}
}
