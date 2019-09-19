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
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
	"github.com/tetratelabs/log"
)

// EnableNodeCollection is a preset option that registers collection of node level information for debugging
func EnableNodeCollection(r *envoy.Runtime) {
	if err := os.MkdirAll(filepath.Join(r.DebugStore(), "node"), os.ModePerm); err != nil {
		log.Errorf("unable to create directory to write node data to: %v", err)
		return
	}
	r.RegisterPreTermination(ps)
	r.RegisterPreTermination(networkInterfaces)
	r.RegisterPreTermination(writeIOStats)
}

func ps(r binary.Runner) error {
	f, err := os.Create(filepath.Join(r.DebugStore(), "node/ps.txt"))
	if err != nil {
		return fmt.Errorf("unable to create file to write ps output to: %v", err)
	}
	defer f.Close() //nolint

	processes, err := process.Processes()
	if err != nil {
		return fmt.Errorf("unable to get list of running processes: %v", err)
	}
	return processPrinter(f, processes)
}

func processPrinter(out io.Writer, processes []*process.Process) error {
	w := tabwriter.NewWriter(out, 0, 8, 5, ' ', 0)
	fmt.Fprintln(w, "PID\tUSERNAME\tSTATUS\tRSS\tVSZ\tMINFLT\tMAJFLT\tPCPU\tPMEM\tARGS")
	for _, p := range processes {
		proc := safeProc(p)
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%.2f\t%.2f\t%v\n", proc.pid, proc.username, proc.status, proc.rss, proc.vms, proc.minflt,
			proc.majflt, proc.pCPU, proc.pMem, proc.cmd)
	}
	return w.Flush()
}

type proc struct {
	username, status, cmd    string
	rss, vms, minflt, majflt uint64
	pCPU                     float64
	pid                      int32
	pMem                     float32
}

func safeProc(p *process.Process) *proc {
	// These are onloy debug logs as on certain OSs these features are not supported
	// If we errorf we spam stderr with errors for every single process
	user, err := p.Username()
	if err != nil {
		log.Debugf("unable to retrieve username of %v: %v", p.Pid, err)
	}
	status, err := p.Status()
	if err != nil {
		log.Debugf("unable to retrieve status of %v: %v", p.Pid, err)
	}
	mem, err := p.MemoryInfo()
	if err != nil {
		log.Debugf("unable to retrieve memory information of %v: %v", p.Pid, err)
	}
	if mem == nil {
		mem = &process.MemoryInfoStat{}
	}
	pagefault, err := p.PageFaults()
	if err != nil {
		log.Debugf("unable to retrieve page fault information of %v: %v", p.Pid, err)
	}
	if pagefault == nil {
		pagefault = &process.PageFaultsStat{}
	}
	pCPU, err := p.CPUPercent()
	if err != nil {
		log.Debugf("unable to retrieve cpu percentage information of %v: %v", p.Pid, err)
	}
	pMem, err := p.MemoryPercent()
	if err != nil {
		log.Debugf("unable to retrieve memory percentage information of %v: %v", p.Pid, err)
	}
	cmd, err := p.Cmdline()
	if err != nil {
		log.Debugf("unable to retrieve command information of %v: %v", p.Pid, err)
	}
	return &proc{
		username: user,
		status:   status,
		rss:      mem.RSS,
		vms:      mem.VMS,
		minflt:   pagefault.MinorFaults,
		majflt:   pagefault.MajorFaults,
		pCPU:     pCPU,
		pMem:     pMem,
		cmd:      cmd,
	}
}

func networkInterfaces(r binary.Runner) error {
	f, err := os.Create(filepath.Join(r.DebugStore(), "node/network_interface.json"))
	if err != nil {
		return fmt.Errorf("unable to create file to write network interface output to: %v", err)
	}
	defer f.Close() //nolint

	is, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("unable to fetch network Interfaces: %v", err)
	}
	out, err := json.Marshal(is)
	if err != nil {
		return fmt.Errorf("unable to convert to json representation: %v", err)
	}
	fmt.Fprintln(f, string(out))

	return nil
}

// writeIOStat write iostat of devices in the form of a dictionary to json file
func writeIOStats(r binary.Runner) error {
	f, err := os.Create(filepath.Join(r.DebugStore(), "node/iostats.json"))
	defer f.Close() //nolint
	if err != nil {
		return fmt.Errorf("error in creating iostat.json: %v", err)
	}

	physicalPartitions, _ := disk.Partitions(false)
	deviceNames := make([]string, 0, len(physicalPartitions))
	//nolint:gocritic
	for _, p := range physicalPartitions {
		deviceNames = append(deviceNames, p.Device)
	}
	IOCounterStatsMap, _ := disk.IOCounters(deviceNames...)

	// serialize map to json and write to file
	jsonBytes, err := json.Marshal(IOCounterStatsMap)
	if err != nil {
		return fmt.Errorf("error in serializing IOCounterStatsMap: %v", err)
	}
	fmt.Fprintln(f, string(jsonBytes))

	return nil
}
