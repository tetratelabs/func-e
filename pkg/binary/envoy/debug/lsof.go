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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/shirou/gopsutil/process"
	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/log"
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

// EnableOpenFilesDataCollection is a preset option that registers collection of statistics of files opened by envoy instance(s)
func EnableOpenFilesDataCollection(r *envoy.Runtime) {
	if err := os.Mkdir(filepath.Join(r.DebugStore(), "lsof"), os.ModePerm); err != nil {
		log.Errorf("error in creating a directory to write open file data of envoy to: %v", err)
	}
	r.RegisterPreTermination(retrieveOpenFilesData)
}

// retrieveOpenFilesData writes statistics of open files associated with envoy instance(s) to a json file
// if succeeded, return nil, else return an error instance
func retrieveOpenFilesData(r binary.Runner) error { //nolint:gocyclo
	// get a list of processes
	processes, err := process.Processes()
	if err != nil {
		return fmt.Errorf("error in getting process pids")
	}

	// filter Process instances of envoy
	isEnvoy := func(p *process.Process) bool {
		name, errName := p.Name()
		if errName != nil {
			log.Errorf("error in getting process name for %v", p)
			return false
		}
		return name == "envoy"
	}
	envoys := make([]*process.Process, 0)
	for _, p := range processes {
		if isEnvoy(p) {
			envoys = append(envoys, p)
		}
	}

	f, err := os.Create(filepath.Join(r.DebugStore(), "lsof/lsof.json"))
	if err != nil {
		return fmt.Errorf("unable to create file to write lisof output to: %v", err)
	}
	defer f.Close() //nolint

	result := make([]OpenFileStat, 0)
	// print open file stats for all envoy instances
	for _, envoy := range envoys {
		// relevant fields of the process
		username, _ := envoy.Username()
		name, _ := envoy.Name()
		pid := envoy.Pid

		openFiles, errOpen := envoy.OpenFiles()
		if errOpen != nil {
			log.Debugf("error in getting ofStat for %v\n", envoy)
			continue
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
			if err := syscall.Stat(stat.Path, &fstat); err != nil {
				// continue if the path is invalid
				result = append(result, ofStat)
				continue
			}

			ofStat.Node = fmt.Sprint(fstat.Ino)
			ofStat.Size = fmt.Sprint(fstat.Size)
			result = append(result, ofStat)
		}
	}

	out, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("unable to convert to json representation: %v", err)
	}

	fmt.Fprintln(f, string(out))

	return nil
}
