// Copyright 2019 Andrew Bates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"io"
	"testing"
)

type errWriter struct{}

func (*errWriter) Write([]byte) (int, error) { return 0, io.EOF }

type shortWriter struct{}

func (*shortWriter) Write(data []byte) (int, error) { return len(data) - 1, nil }

type closeWriter struct{}

func (*closeWriter) Write(data []byte) (int, error) { return len(data), nil }

func (*closeWriter) Close() error { return io.EOF }

func TestMultiWriter(t *testing.T) {
	tests := []struct {
		name     string
		writer   io.Writer
		wantErr  error
		closeErr error
		got      func(io.Writer) []byte
	}{
		{"error writer", &errWriter{}, io.EOF, nil, nil},
		{"short writer", &shortWriter{}, io.ErrShortWrite, nil, nil},
		{"buffer writer", bytes.NewBuffer(nil), nil, nil, func(w io.Writer) []byte { return w.(*bytes.Buffer).Bytes() }},
		{"close writer", &closeWriter{}, nil, io.EOF, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mw := &multiWriter{}
			mw.add(test.writer)

			_, err := mw.Write([]byte("hello world!"))
			if test.wantErr != err {
				t.Errorf("unexpected error: %v", err)
			} else if err == nil {
				if test.got != nil {
					got := test.got(test.writer)
					if !bytes.Equal(got, []byte("hello world!")) {
						t.Errorf("wanted %q got %q", "hello world!", string(got))
					}
				}
			}

			err = mw.Close()
			if test.closeErr != err {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
