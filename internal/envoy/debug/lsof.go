// Copyright 2019 Tetrate
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

package debug

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/shirou/gopsutil/v3/process"

	"github.com/tetratelabs/getenvoy/internal/envoy"
)

// OpenFileStat defines the structure of statistics about a single opened file
type OpenFileStat struct {
	// string enclosed in `` are known as struct tags
	// this particular tag is used by json.Marshal() to encode Command field in a specific manner
	// process dependent information
	Command string `json:"command"`
	Pid     string `json:"pid"`
	User    string `json:"user"`
	Fd      string `json:"fd"`
	// non process dependent information
	Type   string `json:"type"`   // type of node associated with the file / FIFO for FIFO special file
	Device string `json:"device"` // device numbers associated with the file, separated by commas
	Size   string `json:"size"`
	Node   string `json:"node"` // inode number of a local file/ Internet protocol type
	Name   string `json:"name"` // name of the mount point and file system on which the file resides
}

// enableOpenFilesDataCollection is a preset option that registers collection of statistics of files opened by envoy
// instance(s).
func enableOpenFilesDataCollection(r *envoy.Runtime) error {
	lsofDir := filepath.Join(r.GetRunDir(), "lsof")
	if err := os.MkdirAll(lsofDir, 0750); err != nil {
		return fmt.Errorf("unable to create directory %q, so lsof will not be captured: %w", lsofDir, err)
	}
	o := openFilesDataCollection{r.GetEnvoyPid, lsofDir}
	r.RegisterPreTermination(o.retrieveOpenFilesData)
	return nil
}

type openFilesDataCollection struct {
	getPid  func() (int, error)
	lsofDir string
}

// retrieveOpenFilesData writes statistics of open files associated with envoy instance(s) to a json file
func (o *openFilesDataCollection) retrieveOpenFilesData() error {
	// get pid of envoy instance
	p, err := o.getPid()
	if err != nil {
		return fmt.Errorf("error in getting pid of envoy instance: %w", err)
	}
	pid := int32(p)

	ctx, cancel := context.WithTimeout(context.Background(), processTimeout)
	defer cancel()

	envoyProcess, err := process.NewProcessWithContext(ctx, pid)
	if w := wrapError(ctx, err, "process", pid); w != nil {
		return w
	}

	result := make([]OpenFileStat, 0)
	// print open file stats for all envoy relevant fields of the process
	username, err := envoyProcess.UsernameWithContext(ctx)
	if w := wrapError(ctx, err, "username", pid); w != nil {
		return w
	}

	name, err := envoyProcess.NameWithContext(ctx)
	if w := wrapError(ctx, err, "name", pid); w != nil {
		return w
	}

	openFiles, err := envoyProcess.OpenFilesWithContext(ctx)
	if w := wrapError(ctx, err, "open files", pid); w != nil {
		return w
	}

	for _, stat := range openFiles {
		ofStat := OpenFileStat{
			Command: name,
			Pid:     fmt.Sprint(pid),
			User:    username,
			Fd:      fmt.Sprint(stat.Fd),
			Name:    stat.Path,
		}

		var fstat syscall.Stat_t
		if errSyscall := syscall.Stat(stat.Path, &fstat); errSyscall != nil {
			// continue if the path is invalid
			result = append(result, ofStat)
			continue
		}

		ofStat.Node = fmt.Sprint(fstat.Ino)
		ofStat.Size = fmt.Sprint(fstat.Size)
		result = append(result, ofStat)
	}

	if len(result) == 0 {
		return nil // don't write a file on an unsupported platform
	}
	return writeJSON(result, filepath.Join(o.lsofDir, "lsof.json"))
}
