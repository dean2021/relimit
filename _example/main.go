package main

import (
	"fmt"
	"github.com/dean2021/relimit"
	"time"
)

// WorkerMain Subprocess entry function
func WorkerMain() {
	var memory []string
	for {
		// Dead loop, simulating CPU usage increase
		memory = append(memory, "AAAAAAA")
	}
}

func main() {

	control := relimit.New(relimit.Op{
		Name:             "worker-demo",
		MemoryUsageBytes: 1024 * 1024 * 1024,
		CpuUsage:         10,
		Main:             WorkerMain,
	})

	go func(control *relimit.ReLimit) {
		// After 60s, manually stop the process and test it
		time.Sleep(time.Second * 60)
		err := control.Stop()
		if err != nil {
			panic(err)
		}
	}(control)

	err := control.Run()
	if err != nil {
		_ = fmt.Errorf(err.Error())
	}

	fmt.Println("Daemon stopped running")
}
