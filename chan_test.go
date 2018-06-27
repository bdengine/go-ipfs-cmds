package cmds

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
)

func TestChanResponsePair(t *testing.T) {
	type testcase struct {
		values   []interface{}
		closeErr error
	}

	mkTest := func(tc testcase) func(*testing.T) {
		return func(t *testing.T) {
			cmd := &Command{}
			req, err := NewRequest(context.TODO(), nil, nil, nil, nil, cmd)
			if err != nil {
				t.Fatal("error building request", err)
			}
			re, res := NewChanResponsePair(req)

			var wg sync.WaitGroup

			wg.Add(1)
			go func() {
				for _, v := range tc.values {
					v_, err := res.Next()
					if err != nil {
						t.Error("Next returned unexpected error:", err)
					}
					if v != v_ {
						t.Errorf("Next returned unexpected value %q, expected %q", v_, v)
					}
				}

				_, err := res.Next()
				if tc.closeErr == nil || tc.closeErr == io.EOF {
					if err == nil {
						t.Error("Next returned nil error, expecting io.EOF")
					} else if err != io.EOF {
						t.Errorf("Next returned error %q, expecting io.EOF", err)
					}
				} else {
					if err != tc.closeErr {
						t.Errorf("Next returned error %q, expecting %q", err, tc.closeErr)
					}
				}

				wg.Done()
			}()

			for _, v := range tc.values {
				err := re.Emit(v)
				if err != nil {
					t.Error("Emit returned unexpected error:", err)
				}
			}

			re.CloseWithError(tc.closeErr)

			wg.Wait()
		}
	}

	tcs := []testcase{
		{values: []interface{}{1, 2, 3}},
		{values: []interface{}{1, 2, 3}, closeErr: io.EOF},
		{values: []interface{}{1, 2, 3}, closeErr: errors.New("an error occured")},
	}

	for i, tc := range tcs {
		t.Run(fmt.Sprint(i), mkTest(tc))
	}
}
