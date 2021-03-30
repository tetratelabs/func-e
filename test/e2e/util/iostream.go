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

package util

import (
	"bufio"
	"io"
	"regexp"
	"time"

	"github.com/pkg/errors"
)

// LineOrError represents an element in the stream -
// either a line of text or an I/O error.
type LineOrError interface {
	Line() string
	Err() error
}

// StreamLine represents a line of text.
type StreamLine string

// Line returns a line of text.
func (l StreamLine) Line() string {
	return string(l)
}

// Err returns an I/O error occurred while reading a line of text.
func (StreamLine) Err() error {
	return nil
}

// StreamError represents a line of text.
type StreamError struct {
	err error
}

// Line returns a line of text.
func (StreamError) Line() string {
	return ""
}

// Err returns an I/O error occurred while reading a line of text.
func (e StreamError) Err() error {
	return e.err
}

// Stream represents a stream of text lines.
type Stream <-chan LineOrError

// StreamLines returns a stream of lines of text read from a given source.
func StreamLines(r io.Reader) Stream {
	buf := bufio.NewReader(r)
	lines := make(chan LineOrError, 100)
	go func() {
		defer close(lines)
		for {
			line, err := buf.ReadString('\n')
			if err != nil && err != io.EOF {
				lines <- StreamError{errors.Errorf("failed to read the next line: %v", err)}
			}
			lines <- StreamLine(line)
			if err == io.EOF {
				return
			}
		}
	}()
	return lines
}

// NamedStream represents a stream of text lines, such as "stderr" or "stdout".
type NamedStream struct {
	Name string
	Stream
}

// Named returns a new NamedStream.
func (s Stream) Named(name string) *NamedStream {
	return &NamedStream{name, s}
}

// Single represents a stream that can only emit a single text lines.
type Single <-chan LineOrError

// FirstMatch returns a stream that emits the first line matching a given pattern.
func (s *NamedStream) FirstMatch(pattern *regexp.Regexp) Single {
	match := make(chan LineOrError, 1)
	go func() {
		defer close(match)
		for element := range s.Stream {
			if element.Err() != nil {
				match <- element
				return
			}
			if pattern.MatchString(element.Line()) {
				match <- element
				return
			}
		}
		match <- StreamError{errors.Errorf("%q didn't have a line that would match %q", s.Name, pattern)}
	}()
	return match
}

// Wait waits until a stream emit a line of text or fails if timeout has been exceeded.
func (s Single) Wait(timeout time.Duration) (string, error) {
	select {
	case element := <-s:
		return element.Line(), element.Err()
	case <-time.After(timeout):
		return "", errors.Errorf("reached timeout %s", timeout)
	}
}
