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
	"io"
	"sync"
)

type multiWriter struct {
	writers []io.Writer
	mu      sync.Mutex
}

func (mw *multiWriter) add(writer io.Writer) {
	mw.mu.Lock()
	mw.writers = append(mw.writers, writer)
	mw.mu.Unlock()
}

func (mw *multiWriter) copy(reader io.Reader) error {
	io.Copy(mw, reader)
	return mw.close()
}

func (mw *multiWriter) close() (err error) {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	for _, writer := range mw.writers {
		if closer, ok := writer.(io.Closer); ok {
			err = closer.Close()
			if err != nil {
				break
			}
		}
	}
	return
}

func (mw *multiWriter) Close() (err error) {
	return mw.close()
}
func (mw *multiWriter) Write(p []byte) (n int, err error) {
	mw.mu.Lock()
	defer mw.mu.Unlock()
	for _, writer := range mw.writers {
		n, err = writer.Write(p)
		if err != nil {
			return
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return
		}
	}
	return len(p), nil
}
