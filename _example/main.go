package main

import (
	"github.com/dean2021/relimit"
	"log"
	"time"
)

// WorkerMain Subprocess entry function
func WorkerMain() {
	var memory []string
	for {
		// Dead loop, simulating CPU usage increase

		// Allocate memory, resulting in process OOM
		memory = append(memory, "AAAAAAA")
	}
}

func main() {

	control := relimit.New(relimit.Op{
		Name:             "worker-demo",
		MemoryUsageBytes: 1024 * 1024 * 1024 * 1,
		CpuUsage:         10,
		Main:             WorkerMain,
	})

	control.Start()

	// Guard the child process. When the child process stops running, pull it up again
	for {
		time.Sleep(time.Second * 1)
		if !control.IsRunning() {
			log.Println("Stop the subprocess and pull up the subprocess")
			control.Start()
		}
	}
}
