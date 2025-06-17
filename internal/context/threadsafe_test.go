package context

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestThreadSafeContext_ConcurrentAccess(t *testing.T) {
	// Create a context with mutable data
	type MutableContext struct {
		Counter int
		Data    map[string]string
	}
	
	initialCtx := NewContext(MutableContext{
		Counter: 0,
		Data:    map[string]string{"key": "value"},
	})
	
	tsCtx := NewThreadSafeContext(initialCtx)
	
	// Test concurrent reads
	t.Run("concurrent reads", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, 100)
		
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				
				// Perform multiple reads
				for j := 0; j < 10; j++ {
					val := tsCtx.Value()
					if val.Counter != 0 {
						errors <- fmt.Errorf("unexpected counter value: %d", val.Counter)
						return
					}
					if val.Data["key"] != "value" {
						errors <- fmt.Errorf("unexpected data value: %s", val.Data["key"])
						return
					}
					
					typ := tsCtx.Type()
					if typ == nil {
						errors <- fmt.Errorf("Type() returned nil")
						return
					}
				}
			}()
		}
		
		wg.Wait()
		close(errors)
		
		for err := range errors {
			t.Error(err)
		}
	})
	
	// Test concurrent reads and writes
	t.Run("concurrent reads and writes", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, 100)
		
		// Writers
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				
				newCtx := NewContext(MutableContext{
					Counter: id,
					Data:    map[string]string{"key": fmt.Sprintf("value%d", id)},
				})
				tsCtx.Update(newCtx)
			}(i)
		}
		
		// Readers
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				
				for j := 0; j < 10; j++ {
					val := tsCtx.Value()
					// Just verify we can read without panic
					_ = val.Counter
					_ = val.Data
					
					time.Sleep(time.Microsecond)
				}
			}()
		}
		
		wg.Wait()
		close(errors)
		
		for err := range errors {
			t.Error(err)
		}
	})
}

func TestThreadSafeAnyContext_ConcurrentAccess(t *testing.T) {
	ctx := NewContext("test value")
	anyCtx := ToAnyContext(ctx)
	tsAnyCtx := NewThreadSafeAnyContext(anyCtx)
	
	t.Run("concurrent operations", func(t *testing.T) {
		var wg sync.WaitGroup
		
		// Multiple readers
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				
				for j := 0; j < 10; j++ {
					typeName := tsAnyCtx.TypeName()
					if typeName != "string" {
						t.Errorf("unexpected type name: %s", typeName)
					}
					
					isNil := tsAnyCtx.IsNil()
					if isNil {
						t.Error("IsNil returned true for non-nil context")
					}
					
					ctx := tsAnyCtx.Get()
					if ctx == nil {
						t.Error("Get returned nil")
					}
				}
			}()
		}
		
		// Writer
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for i := 0; i < 10; i++ {
				newCtx := NewContext(fmt.Sprintf("value %d", i))
				newAnyCtx := ToAnyContext(newCtx)
				tsAnyCtx.Update(newAnyCtx)
				time.Sleep(time.Millisecond)
			}
		}()
		
		wg.Wait()
	})
}

func TestThreadSafeContext_NilHandling(t *testing.T) {
	var tsCtx ThreadSafeContext[string]
	
	// Test operations on zero-value ThreadSafeContext
	val := tsCtx.Value()
	if val != "" {
		t.Errorf("Expected empty string, got %s", val)
	}
	
	typ := tsCtx.Type()
	if typ != reflect.TypeOf("") {
		t.Errorf("Unexpected type: %v", typ)
	}
	
	// Update with nil
	tsCtx.Update(nil)
	val = tsCtx.Value()
	if val != "" {
		t.Errorf("Expected empty string after nil update, got %s", val)
	}
}

func BenchmarkThreadSafeContext_Read(b *testing.B) {
	ctx := NewContext("test value")
	tsCtx := NewThreadSafeContext(ctx)
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = tsCtx.Value()
		}
	})
}

func BenchmarkThreadSafeContext_Write(b *testing.B) {
	ctx := NewContext("test value")
	tsCtx := NewThreadSafeContext(ctx)
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tsCtx.Update(ctx)
		}
	})
}