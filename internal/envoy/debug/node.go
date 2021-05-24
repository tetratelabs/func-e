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
	"io"
	"os"
	"path/filepath"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"

	envoy2 "github.com/tetratelabs/getenvoy/internal/envoy"
)

// enableNodeCollection is a preset option that registers collection of node level information for debugging
func enableNodeCollection(r *envoy2.Runtime) error {
	nodeDir := filepath.Join(r.GetWorkingDir(), "node")
	if err := os.MkdirAll(nodeDir, 0750); err != nil {
		return fmt.Errorf("unable to create directory %q, so node data will not be captured: %w", nodeDir, err)
	}
	n := nodeCollection{nodeDir}
	r.RegisterPreTermination(n.ps)
	r.RegisterPreTermination(n.networkInterfaces)
	r.RegisterPreTermination(n.activeConnections)
	return nil
}

type nodeCollection struct {
	nodeDir string
}

// Don't wait forever. This has hung on macOS before
const processTimeout = 3 * time.Second

func (n *nodeCollection) ps() error {
	f, err := os.Create(filepath.Join(n.nodeDir, "ps.txt"))
	if err != nil {
		return fmt.Errorf("unable to create file to write ps output to: %w", err)
	}
	defer f.Close() //nolint

	ctx, cancel := context.WithTimeout(context.Background(), processTimeout)
	defer cancel()

	processes, err := process.ProcessesWithContext(ctx)
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		if err.Error() == "not implemented yet" { // internal error used by gopsutil
			return nil // It will never work, so there's no reason to bother users
		}
		return fmt.Errorf("unable to get list of running processes: %w", err)
	}

	parsed, err := parseProcessTable(ctx, processes)
	if len(parsed) == 0 {
		if err != nil {
			return fmt.Errorf("unable to parse any processes: %w", err)
		}
		return nil
	}

	return printProcessTable(f, parsed)
}

type proc struct {
	username, cmd            string
	status                   []string
	rss, vms, minflt, majflt uint64
	pCPU                     float64
	pid                      int32
	pMem                     float32
}

// parseProcessTable returns processes that could be parsed and the first error
func parseProcessTable(ctx context.Context, processes []*process.Process) ([]*proc, error) {
	procs := make([]*proc, 0, len(processes))
	var err error
	for _, p := range processes {
		parsed, e := parseProc(ctx, p)
		if e != nil { // Continue as the process could have died between the process listing and now.
			if err == nil { // Capture only one error
				err = e
			}
			continue
		}
		procs = append(procs, parsed)
	}
	return procs, err
}

// parseProc returns a proc if there were no errors parsing its data
func parseProc(ctx context.Context, p *process.Process) (*proc, error) {
	user, err := p.UsernameWithContext(ctx)
	if w := wrapError(ctx, err, "username", p.Pid); w != nil {
		return nil, w
	}

	status, err := p.StatusWithContext(ctx)
	if w := wrapError(ctx, err, "status", p.Pid); w != nil {
		return nil, w
	}

	mem, err := p.MemoryInfoWithContext(ctx)
	if w := wrapError(ctx, err, "memory information", p.Pid); w != nil {
		return nil, w
	}
	if mem == nil {
		mem = &process.MemoryInfoStat{}
	}

	pagefault, err := p.PageFaultsWithContext(ctx)
	if w := wrapError(ctx, err, "page faults", p.Pid); w != nil {
		return nil, w
	}
	if pagefault == nil {
		pagefault = &process.PageFaultsStat{}
	}

	pCPU, err := p.CPUPercentWithContext(ctx)
	if w := wrapError(ctx, err, "CPU percent", p.Pid); w != nil {
		return nil, w
	}

	pMem, err := p.MemoryPercentWithContext(ctx)
	if w := wrapError(ctx, err, "memory percent", p.Pid); w != nil {
		return nil, w
	}

	cmd, err := p.CmdlineWithContext(ctx)
	if w := wrapError(ctx, err, "command-line", p.Pid); w != nil {
		return nil, w
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
	}, nil
}

func printProcessTable(out io.Writer, parsed []*proc) error {
	// Now, start writing the process table
	w := tabwriter.NewWriter(out, 0, 8, 5, ' ', 0)
	if _, err := fmt.Fprintln(w, "PID\tUSERNAME\tSTATUS\tRSS\tVSZ\tMINFLT\tMAJFLT\tPCPU\tPMEM\tARGS"); err != nil {
		return err
	}

	for _, p := range parsed {
		status := ""
		if len(p.status) > 0 {
			status = p.status[0]
		}
		if _, err := fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%.2f\t%.2f\t%v\n",
			p.pid, p.username, status, p.rss, p.vms, p.minflt, p.majflt, p.pCPU, p.pMem, p.cmd); err != nil {
			return err
		}
	}
	return w.Flush()
}

func (n *nodeCollection) networkInterfaces() error {
	result, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("unable to list network interfaces: %w", err)
	}

	if len(result) == 0 {
		return nil // don't write a file on an unsupported platform
	}
	return writeJSON(result, filepath.Join(n.nodeDir, "network_interface.json"))
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

func (n *nodeCollection) activeConnections() error {
	cs, err := net.Connections("all")
	if err != nil {
		return fmt.Errorf("unable to list network connections: %w", err)
	}

	result := make([]connStat, 0, len(cs))
	for i := range cs {
		st := addLabelToConnection(&cs[i])
		result = append(result, st)
	}

	if len(result) == 0 {
		return nil // don't write a file on an unsupported platform
	}
	return writeJSON(result, filepath.Join(n.nodeDir, "connections.json"))
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
