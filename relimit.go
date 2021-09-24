// Copyright 2021 Dean.
// Authors: Dean <dean@csoio.com>
// Date: 2021/9/23 4:01 下午

package relimit

import (
	"fmt"
	"github.com/dean2021/relimit/reexec"
	"github.com/shirou/gopsutil/v3/process"
	"os"
)

type Op struct {
	// 子进程名
	Name string
	// 内存限制
	MemoryUsageBytes uint64
	// cpu使用率限制
	CpuUsage float64
	// 子进程入口函数
	Main func()
}

type ReLimit struct {
	Name             string
	MemoryUsageBytes uint64
	CpuUsage         float64
	Process          *process.Process
}

func (rl *ReLimit) cpuLimit() {
	var isSuspend bool
	for {
		percent, err := rl.Process.CPUPercent()
		if err != nil {
			panic(err)
			return
		}
		if percent > rl.CpuUsage {
			if isSuspend {
				continue
			}
			err := rl.Suspend()
			if err != nil {
				panic(err)
				return
			}
			isSuspend = true
		} else {
			if !isSuspend {
				continue
			}
			err := rl.Resume()
			if err != nil {
				panic(err)
				return
			}
			isSuspend = false
		}
	}
}

func (rl *ReLimit) memoryLimit() {
	for {
		info, err := rl.Process.MemoryInfo()
		if err != nil {
			panic(err)
			return
		}
		if info.RSS > rl.MemoryUsageBytes {
			err := rl.Stop()
			if err != nil {
				panic(err)
				return
			}
		}
	}
}

func (rl *ReLimit) Run() error {
	if reexec.Init() {
		os.Exit(0)
	}
	cmd := reexec.Command(rl.Name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to run command: %s", err)
	}

	newProcess, err := process.NewProcess(int32(cmd.Process.Pid))
	if err != nil {
		return fmt.Errorf("failed stop: %s", err)
	}
	rl.Process = newProcess

	go rl.cpuLimit()
	go rl.memoryLimit()

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait command: %s", err)
	}
	return nil
}

func (rl *ReLimit) Stop() error {
	err := rl.Process.Kill()
	if err != nil {
		return fmt.Errorf("failed stop: %s", err)
	}
	return nil
}

func (rl *ReLimit) Suspend() error {
	err := rl.Process.Suspend()
	if err != nil {
		return fmt.Errorf("failed suspend: %s", err)
	}
	return nil
}

func (rl *ReLimit) Resume() error {
	err := rl.Process.Resume()
	if err != nil {
		return fmt.Errorf("failed suspend: %s", err)
	}
	return nil
}

func New(op Op) *ReLimit {
	reexec.Register(op.Name, op.Main)
	return &ReLimit{
		Name:             op.Name,
		CpuUsage:         op.CpuUsage,
		MemoryUsageBytes: op.MemoryUsageBytes,
	}
}
