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
	"io"
	"os"
	"path/filepath"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/tetratelabs/log"

	"github.com/tetratelabs/getenvoy/pkg/binary"
	"github.com/tetratelabs/getenvoy/pkg/binary/envoy"
)

// EnableNodeCollection is a preset option that registers collection of node level information for debugging
func EnableNodeCollection(r *envoy.Runtime) {
	if err := os.MkdirAll(filepath.Join(r.DebugStore(), "node"), os.ModePerm); err != nil {
		log.Errorf("unable to create directory to write node data to: %v", err)
		return
	}
	r.RegisterPreTermination(ps)
	r.RegisterPreTermination(networkInterfaces)
	r.RegisterPreTermination(activeConnections)
}

// Don't wait forever. This has hung on macOS before
const processTimeout = 3 * time.Second

// Errors from platform-specific libraries log to debug instead of raising an error or logging in an error category.
// This avoids filling logs for unresolvable reasons.
func ps(r binary.Runner) error {
	f, err := os.Create(filepath.Join(r.DebugStore(), "node/ps.txt"))
	if err != nil {
		return fmt.Errorf("unable to create file to write ps output to: %w", err)
	}
	defer f.Close() //nolint

	ctx, cancel := context.WithTimeout(context.Background(), processTimeout)
	defer cancel()

	processes, err := process.ProcessesWithContext(ctx)
	if ctx.Err() == context.DeadlineExceeded {
		log.Debugf("timeout getting a list of running processes")
		return nil
	}
	if err != nil {
		log.Debugf("unable to get list of running processes: %v", err)
		return nil
	}

	return printProcessTable(ctx, f, processes)
}

func printProcessTable(ctx context.Context, out io.Writer, processes []*process.Process) error {
	w := tabwriter.NewWriter(out, 0, 8, 5, ' ', 0)
	fmt.Fprintln(w, "PID\tUSERNAME\tSTATUS\tRSS\tVSZ\tMINFLT\tMAJFLT\tPCPU\tPMEM\tARGS")
	for _, p := range processes {
		proc := safeProc(ctx, p)
		if proc == empty { // Ignore, but continue. The process could have died between the process listing and now.
			continue
		}
		status := ""
		if proc.status != nil && len(proc.status) > 0 {
			status = proc.status[0]
		}
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%.2f\t%.2f\t%v\n", proc.pid, proc.username, status, proc.rss, proc.vms, proc.minflt,
			proc.majflt, proc.pCPU, proc.pMem, proc.cmd)
	}
	return w.Flush()
}

type proc struct {
	username, cmd            string
	status                   []string
	rss, vms, minflt, majflt uint64
	pCPU                     float64
	pid                      int32
	pMem                     float32
}

var empty = &proc{}

func safeProc(ctx context.Context, p *process.Process) *proc {
	user, err := p.UsernameWithContext(ctx)
	if logDebugOnError(ctx, err, "username", p.Pid) {
		return empty
	}

	status, err := p.StatusWithContext(ctx)
	if logDebugOnError(ctx, err, "status", p.Pid) {
		return empty
	}

	mem, err := p.MemoryInfoWithContext(ctx)
	if logDebugOnError(ctx, err, "memory information", p.Pid) {
		return empty
	}
	if mem == nil {
		mem = &process.MemoryInfoStat{}
	}

	pagefault, err := p.PageFaultsWithContext(ctx)
	if logDebugOnError(ctx, err, "page faults", p.Pid) {
		return empty
	}
	if pagefault == nil {
		pagefault = &process.PageFaultsStat{}
	}

	pCPU, err := p.CPUPercentWithContext(ctx)
	if logDebugOnError(ctx, err, "CPU percent", p.Pid) {
		return empty
	}

	pMem, err := p.MemoryPercentWithContext(ctx)
	if logDebugOnError(ctx, err, "memory percent", p.Pid) {
		return empty
	}

	cmd, err := p.CmdlineWithContext(ctx)
	if logDebugOnError(ctx, err, "command-line", p.Pid) {
		return empty
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

func logDebugOnError(ctx context.Context, err error, field string, pid int32) bool {
	if ctx.Err() == context.DeadlineExceeded {
		log.Debugf("timeout getting %s of pid %d", field, pid)
		return true
	} else if err != nil {
		log.Debugf("unable to retrieve %s of pid %d: %v", field, pid, err)
		return true
	}
	return false
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

type connStat struct {
	Fd     uint32   `json:"fd"`
	Pid    int32    `json:"pid"`
	Uids   []int32  `json:"uids"`
	Family string   `json:"family"`
	Type   string   `json:"type"`
	Status string   `json:"status"`
	Laddr  net.Addr `json:"localaddr"`
	Raddr  net.Addr `json:"remoteaddr"`
}

var familyMap = map[uint32]string{
	syscall.AF_INET:  "AF_INET",
	syscall.AF_INET6: "AF_INET6",
	syscall.AF_UNIX:  "AF_UNIX",
}

var typeMap = map[uint32]string{
	syscall.SOCK_STREAM: "SOCK_STREAM",
	syscall.SOCK_DGRAM:  "SOCK_DGRAM",
}

func activeConnections(r binary.Runner) error {
	f, err := os.Create(filepath.Join(r.DebugStore(), "node/connections.json"))
	if err != nil {
		return fmt.Errorf("unable to create file to write network interface output to: %v", err)
	}
	defer f.Close() //nolint

	cs, err := net.Connections("all")
	if err != nil {
		return fmt.Errorf("unable to fetch network Interfaces: %v", err)
	}

	ret := make([]connStat, 0, len(cs))
	for i := range cs {
		st := addLabelToConnection(&cs[i])
		ret = append(ret, st)
	}
	out, err := json.Marshal(ret)
	if err != nil {
		return fmt.Errorf("unable to convert to json representation: %v", err)
	}
	fmt.Fprintln(f, string(out))

	return nil
}

// Replace uint32 label to human readable string label.
func addLabelToConnection(orig *net.ConnectionStat) connStat {
	family, ok := familyMap[orig.Family]
	if !ok {
		family = fmt.Sprintf("unknown(%v)", orig.Family)
	}
	t, ok := typeMap[orig.Type]
	if !ok {
		t = fmt.Sprintf("unknown(%v)", orig.Type)
	}
	return connStat{
		Fd:     orig.Fd,
		Family: family,
		Type:   t,
		Laddr:  orig.Laddr,
		Raddr:  orig.Raddr,
		Status: orig.Status,
		Uids:   orig.Uids,
		Pid:    orig.Pid,
	}
}
