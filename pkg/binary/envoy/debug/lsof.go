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
	"syscall"

	"github.com/shirou/gopsutil/process"
	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
)

// Lsof defines the structure of statistics about a single opened file
type Lsof struct {
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
	r.RegisterPreTermination(retrieveOpenFilesData)
}

func retrieveOpenFilesData(r binary.Runner) error { //nolint:gocyclo
	// get a list of processes
	processes, err := process.Processes()
	if err != nil {
		fmt.Println("error in getting pids")
	}

	// filter Process instances of envoy
	isEnvoy := func(p *process.Process) bool {
		name, err := p.Name()
		if err != nil {
			fmt.Println("error in getting process name ")
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
	fmt.Println("------- filtered envoy processes ---------")
	f, err := os.Create("./lsof.json")
	if err != nil {
		fmt.Println("unable to create file to write lisof output to", err)
	}
	defer f.Close() //nolint

	ofStatArr := make([]Lsof, 0)
	// print open file stats for all envoy instances
	for i, envoy := range envoys {
		fmt.Printf("--------open file stat for envoy instance %d: %s------\n", i, envoy)
		// relevant fields of the process
		username, _ := envoy.Username()
		name, _ := envoy.Name()
		pid := envoy.Pid

		ofStatTemp, err := envoy.OpenFiles()
		if err != nil {
			fmt.Printf("error in getting ofStat for %v\n", envoy)
		}
		fmt.Println(ofStatTemp)

		for _, stat := range ofStatTemp {
			statPath := stat.Path

			ofStat := Lsof{
				Command: name,
				Pid:     fmt.Sprint(pid),
				User:    username,
				Fd:      fmt.Sprint(stat.Fd),
				Name:    statPath,
			}

			fmt.Println("-------------current stat path-------------", statPath)
			var fstat syscall.Stat_t
			if err := syscall.Stat(statPath, &fstat); err != nil {
				// continue if the path is invalid
				ofStatArr = append(ofStatArr, ofStat)
				continue
			}
			fmt.Printf("System info: %+v\n\n", fstat)
			fmt.Println("Size in bytes:", fstat.Size)
			fmt.Println("inode number: ", fstat.Ino)
			ofStat.Node = fmt.Sprint(fstat.Ino)
			ofStat.Size = fmt.Sprint(fstat.Size)
			ofStatArr = append(ofStatArr, ofStat)
		}
	}

	out, err := json.Marshal(ofStatArr)
	if err != nil {
		fmt.Println("unable to convert to json representation", err)
	}
	// write to file
	fmt.Fprintln(f, string(out))

	return nil
}
