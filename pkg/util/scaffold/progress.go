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

package scaffold

// ProgressSink is a sink for progress events.
type ProgressSink interface {
	// OnStart is called when scaffolding is about to get started.
	OnStart()
	// OnFile is called for every generated file.
	OnFile(file string)
	// OnComplete is called when scaffolding has finished successfully.
	OnComplete()
}

// NoOpProgressSink returns a no-op implementation of ProgressSink.
func NoOpProgressSink() ProgressSink {
	return ProgressSinkFuncs{}
}

// ProgressSinkFuncs dispatches progress events to individual handler functions.
type ProgressSinkFuncs struct {
	OnStartFunc    func()
	OnFileFunc     func(file string)
	OnCompleteFunc func()
}

// OnStart is called when scaffolding is about to get started.
func (f ProgressSinkFuncs) OnStart() {
	if f.OnStartFunc != nil {
		f.OnStartFunc()
	}
}

// OnFile is called for every generated file.
func (f ProgressSinkFuncs) OnFile(file string) {
	if f.OnFileFunc != nil {
		f.OnFileFunc(file)
	}

}

// OnComplete is called when scaffolding has finished successfully.
func (f ProgressSinkFuncs) OnComplete() {
	if f.OnCompleteFunc != nil {
		f.OnCompleteFunc()
	}
}
