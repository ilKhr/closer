package closer_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/ilKhr/closer"
	"github.com/stretchr/testify/require"
)

type mockCloseFunc struct {
	calledCount int
	mu          sync.Mutex
}

func (m *mockCloseFunc) close(ctx context.Context) error {
	m.mu.Lock()
	m.calledCount++
	m.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

type sizeTestCase struct {
	mocks []*mockCloseFunc
}

func getTestCases() []sizeTestCase {
	return []sizeTestCase{
		{mocks: []*mockCloseFunc{}},
		{mocks: []*mockCloseFunc{{}}},
		{mocks: []*mockCloseFunc{{}, {}}},
		{mocks: []*mockCloseFunc{{}, {}, {}}},
	}
}

func Test_Size_HappyPath(t *testing.T) {
	sizeTestCases := getTestCases()

	for i, test := range sizeTestCases {
		t.Run(fmt.Sprintf("Close_function_count_%d", i), func(t *testing.T) {
			var cl closer.Closer

			for _, mcf := range test.mocks {
				cl.Add(mcf.close)
			}

			require.Equal(t, len(test.mocks), cl.Size())
		})

	}
}

func Test_CancelOne_CancelWithCtxPath(t *testing.T) {
	sizeTestCases := getTestCases()

	for i, test := range sizeTestCases {
		t.Run(fmt.Sprintf("Close_function_count_%d", i), func(t *testing.T) {

			var cl closer.Closer

			for _, mcf := range test.mocks {
				cl.Add(mcf.close)
			}

			ttlContext, cancel := context.WithCancel(context.Background())
			cancel()

			for _, mcf := range test.mocks {
				err := cl.CloseOne(ttlContext)
				require.ErrorContains(t, err, context.Canceled.Error())
				require.Equal(t, 1, mcf.calledCount)
			}
		})
	}
}

func Test_CancelOne_HappyPath(t *testing.T) {
	sizeTestCases := getTestCases()

	for i, test := range sizeTestCases {
		t.Run(fmt.Sprintf("Close_function_count_%d", i), func(t *testing.T) {

			var cl closer.Closer

			for _, mcf := range test.mocks {
				cl.Add(mcf.close)
			}

			for _, mcf := range test.mocks {
				err := cl.CloseOne(context.Background())
				require.NoError(t, err)
				require.Equal(t, 1, mcf.calledCount)
			}
		})
	}
}

func Test_CancelOne_CallMoreThanHasFuncsPath(t *testing.T) {
	sizeTestCases := getTestCases()

	for i, test := range sizeTestCases {
		t.Run(fmt.Sprintf("Close_function_count_%d", i), func(t *testing.T) {

			var cl closer.Closer

			for _, mcf := range test.mocks {
				cl.Add(mcf.close)
			}

			for range test.mocks {
				cl.CloseOne(context.Background())
			}

			err := cl.CloseOne(context.Background())

			require.ErrorContains(t, err, closer.ErrAllServicesClosed)
		})
	}
}

func Test_Cancel_HappyPath(t *testing.T) {
	sizeTestCases := getTestCases()

	for i, test := range sizeTestCases {
		t.Run(fmt.Sprintf("Close_function_count_%d", i), func(t *testing.T) {

			var cl closer.Closer

			for _, mcf := range test.mocks {
				cl.Add(mcf.close)
			}

			err := cl.Close(context.Background())

			for _, mcf := range test.mocks {
				require.Equal(t, 1, mcf.calledCount)
			}

			if len(test.mocks) == 0 {
				require.ErrorContains(t, err, closer.ErrAllServicesClosed)
			} else {
				require.NoError(t, err)
			}

			errCloseOne := cl.CloseOne(context.Background())

			require.ErrorContains(t, errCloseOne, closer.ErrAllServicesClosed)
		})
	}
}

func Test_Cancel_CallMoreThanHasFuncsPath(t *testing.T) {
	sizeTestCases := getTestCases()

	for i, test := range sizeTestCases {
		t.Run(fmt.Sprintf("Close_function_count_%d", i), func(t *testing.T) {

			var cl closer.Closer

			for _, mcf := range test.mocks {
				cl.Add(mcf.close)
			}

			cl.Close(context.Background())

			for _, mcf := range test.mocks {
				require.Equal(t, 1, mcf.calledCount)
			}

			err := cl.Close(context.Background())

			errCloseOne := cl.CloseOne(context.Background())

			require.ErrorContains(t, err, closer.ErrAllServicesClosed)

			require.ErrorContains(t, err, closer.ErrAllServicesClosed)

			require.ErrorContains(t, errCloseOne, closer.ErrAllServicesClosed)
		})
	}
}

func Test_Cancel_CancelWithCtxPath(t *testing.T) {
	sizeTestCases := getTestCases()

	for i, test := range sizeTestCases {
		t.Run(fmt.Sprintf("Close_function_count_%d", i), func(t *testing.T) {

			var cl closer.Closer

			for _, mcf := range test.mocks {
				cl.Add(mcf.close)
			}

			ttlContext, cancel := context.WithCancel(context.Background())
			cancel()

			err := cl.Close(ttlContext)

			errCloseOne := cl.CloseOne(context.Background())

			for _, mcf := range test.mocks {
				require.Equal(t, 1, mcf.calledCount)
			}

			if len(test.mocks) == 0 {
				require.ErrorContains(t, err, closer.ErrAllServicesClosed)
			} else {
				require.ErrorContains(t, err, context.Canceled.Error())
			}

			require.ErrorContains(t, errCloseOne, closer.ErrAllServicesClosed)
		})
	}
}

func Test_CloseOne_MultiThreadedPath(t *testing.T) {
	var cl closer.Closer
	mocks := []*mockCloseFunc{{}, {}, {}}

	for _, mcf := range mocks {
		cl.Add(mcf.close)
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, mcf := range mocks {
		wg.Add(1)
		go func(m *mockCloseFunc) {
			defer wg.Done()
			err := cl.CloseOne(ctx)
			require.NoError(t, err)
		}(mcf)
	}

	wg.Wait()
}

func Test_CloseOne_MultiThreaded_CancelWithCtxPath(t *testing.T) {
	var cl closer.Closer
	mocks := []*mockCloseFunc{{}, {}, {}}

	for _, mcf := range mocks {
		cl.Add(mcf.close)
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	for _, mcf := range mocks {
		wg.Add(1)
		go func(m *mockCloseFunc) {
			defer wg.Done()
			err := cl.CloseOne(ctx)

			require.ErrorContains(t, err, context.Canceled.Error())
		}(mcf)
	}

	wg.Wait()
}
