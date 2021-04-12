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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/tetratelabs/log"

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
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

// EnableOpenFilesDataCollection is a preset option that registers collection of statistics of files opened by envoy
// instance(s). This is unsupported on macOS/Darwin because it does not support process.OpenFiles
func EnableOpenFilesDataCollection(r *envoy.Runtime) {
	if err := os.Mkdir(filepath.Join(r.DebugStore(), "lsof"), os.ModePerm); err != nil {
		log.Errorf("error in creating a directory to write open file data of envoy to: %v", err)
	}
	r.RegisterPreTermination(retrieveOpenFilesData)
}

// retrieveOpenFilesData writes statistics of open files associated with envoy instance(s) to a json file
// Errors from platform-specific libraries log to debug instead of raising an error or logging in an error category.
// This avoids filling logs for unresolvable reasons.
func retrieveOpenFilesData(r binary.Runner) error {
	f, err := os.Create(filepath.Join(r.DebugStore(), "lsof/lsof.json"))
	if err != nil {
		return fmt.Errorf("error in creating a file to write open file statistics to: %w", err)
	}
	defer f.Close() //nolint

	// get pid of envoy instance
	p, err := r.GetPid()
	if err != nil {
		return fmt.Errorf("error in getting pid of envoy instance: %w", err)
	}
	pid := int32(p)

	ctx, cancel := context.WithTimeout(context.Background(), processTimeout)
	defer cancel()

	envoyProcess, err := process.NewProcessWithContext(ctx, pid)
	if logDebugOnError(ctx, err, "process", pid) {
		return nil
	}

	result := make([]OpenFileStat, 0)
	// print open file stats for all envoy instances
	// relevant fields of the process
	username, err := envoyProcess.UsernameWithContext(ctx)
	if logDebugOnError(ctx, err, "username", pid) {
		return nil
	}

	name, err := envoyProcess.NameWithContext(ctx)
	if logDebugOnError(ctx, err, "name", pid) {
		return nil
	}

	openFiles, err := envoyProcess.OpenFilesWithContext(ctx)
	if logDebugOnError(ctx, err, "open files", pid) {
		return nil
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

	if err = json.NewEncoder(f).Encode(result); err != nil {
		return fmt.Errorf("error writing JSON to file %v: %w", f, err)
	}
	return nil
}
