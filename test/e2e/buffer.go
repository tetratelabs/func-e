// Copyright 2020 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	"bytes"
	"sync"
)

func newSyncBuffer() *syncBuffer {
	return &syncBuffer{
		cond: sync.NewCond(new(sync.Mutex)),
	}
}

// syncBuffer represents a synchronized version of bytes.Buffer.
type syncBuffer struct {
	cond   *sync.Cond
	buffer bytes.Buffer
}

func (s *syncBuffer) Read(p []byte) (n int, err error) {
	s.cond.L.Lock()
	// bytes.Buffer returns io.EOF error if there are no more unread bytes left.
	// we want to avoid these false EOFs and wait until more data becomes available.
	for s.buffer.Len() == 0 {
		s.cond.Wait()
	}
	defer s.cond.L.Unlock()
	return s.buffer.Read(p)
}

func (s *syncBuffer) Write(p []byte) (n int, err error) {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	defer s.cond.Broadcast()
	return s.buffer.Write(p)
}
