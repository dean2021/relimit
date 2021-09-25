// Copyright 2021 Dean.
// Authors: Dean <dean@csoio.com>
// Date: 2021/9/23 4:01 下午

package relimit

import (
	"fmt"
	"github.com/dean2021/relimit/reexec"
	"github.com/shirou/gopsutil/v3/process"
	"log"
	"os"
	"time"
)

type Op struct {
	// Child process name
	Name string
	// Memory limit in bytes
	MemoryUsageBytes uint64
	// CPU usage limit
	CpuUsage float64
	// Subprocess entry function
	Main func()
}

type ReLimit struct {
	Name             string
	MemoryUsageBytes uint64
	CpuUsage         float64
	Process          *process.Process
}

func (rl *ReLimit) cpuLimit() {
	log.Println("Start monitoring CPU")
	var isSuspend bool
	for {
		time.Sleep(time.Millisecond * 500)
		if rl.Process != nil {
			percent, err := rl.Process.CPUPercent()
			if err != nil {
				if err.Error() == "exit status 1" {
					rl.Process = nil
				}
				log.Printf("Unable to limit cpu:%v\n", err)
				return
			}
			if percent > rl.CpuUsage {
				if isSuspend {
					continue
				}
				err := rl.Suspend()
				if err != nil {
					log.Println(err)
					return
				}
				isSuspend = true
			} else {
				if !isSuspend {
					continue
				}
				err := rl.Resume()
				if err != nil {
					log.Println(err)
					return
				}
				isSuspend = false
			}
		} else {
			break
		}
	}
	log.Println("Stop monitoring CPU")
}

func (rl *ReLimit) memoryLimit() {

	log.Println("Start monitoring memory")
	for {
		time.Sleep(time.Millisecond * 500)
		if rl.Process != nil {
			info, err := rl.Process.MemoryInfo()
			if err != nil {
				if err.Error() == "exit status 1" {
					rl.Process = nil
				}
				log.Printf("Unable to limit memory:%v\n", err)
				return
			}
			if info.RSS > rl.MemoryUsageBytes {
				err := rl.Stop()
				if err != nil {
					log.Println(err)
					return
				}
				log.Printf("Out of memory : Kill process %d", rl.Process.Pid)
			}
		} else {
			break
		}
	}
	log.Println("Stop monitoring memory")
}

func (rl *ReLimit) Start() {
	go func() {
		if reexec.Init() {
			os.Exit(0)
		}
		cmd := reexec.Command(rl.Name)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			log.Printf("failed to run command: %s\n", err)
			return
		}

		newProcess, err := process.NewProcess(int32(cmd.Process.Pid))
		if err != nil {
			log.Printf("failed to start process: %s\n", err)
			return
		}
		rl.Process = newProcess

		go rl.cpuLimit()
		go rl.memoryLimit()

		if err := cmd.Wait(); err != nil {
			log.Printf("failed to wait command: %s\n", err)
			return
		}
	}()
}

func (rl *ReLimit) Stop() error {
	if rl.Process == nil {
		return nil
	}
	err := rl.Process.Kill()
	if err != nil {
		return fmt.Errorf("Unable to stop process: %v\n", err)
	}
	return nil
}

func (rl *ReLimit) Suspend() error {
	if rl.Process == nil {
		return nil
	}
	err := rl.Process.Suspend()
	if err != nil {
		return fmt.Errorf("Unable to suspend process: %v\n", err)
	}
	return nil
}

func (rl *ReLimit) Resume() error {
	if rl.Process == nil {
		return nil
	}
	err := rl.Process.Resume()
	if err != nil {
		return fmt.Errorf("Unable to resume process: %v\n", err)
	}
	return nil
}

func (rl *ReLimit) IsRunning() bool {
	if rl.Process == nil {
		return false
	} else {
		return true
	}
}

func New(op Op) *ReLimit {
	reexec.Register(op.Name, op.Main)
	return &ReLimit{
		Name:             op.Name,
		CpuUsage:         op.CpuUsage,
		MemoryUsageBytes: op.MemoryUsageBytes,
	}
}
