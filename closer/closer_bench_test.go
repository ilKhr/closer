package closer

import (
	"context"
	"testing"
)

func BenchmarkCloser_Close(b *testing.B) {

	var cl Closer

	for j := 0; j < 100; j++ {
		cl.Add(func(ctx context.Context) error {
			return nil
		})
	}

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := cl.Close(ctx)
		cl.reset()

		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}

func BenchmarkCloser_CloseOne(b *testing.B) {
	var cl Closer

	for j := 0; j < 100; j++ {
		cl.Add(func(ctx context.Context) error {
			return nil
		})
	}

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := cl.CloseOne(ctx)
		cl.reset()
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
