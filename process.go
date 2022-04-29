package main

import (
	ps "github.com/mitchellh/go-ps"
	"github.com/struCoder/pidusage"
	"log"
	"os"
	"strings"
)

func processChecker(me *MonitorEntry) error {
	// get all running processes
	processList, err := ps.Processes()
	if err != nil {
		return err
	}

	found := false

	// find the process
	for i := 0; i < len(processList); i++ {

		//if strings.TrimSpace(processList[i].Executable()) == me.NameRegex {
		if me.getRegex().MatchString(strings.TrimSpace(processList[i].Executable())) {
			// only check parent process
			if processList[i].PPid() != 1 {
				continue
			}
			found = true

			processInfo, err := pidusage.GetStat(processList[i].Pid())
			if err != nil {
				log.Println(err)
				continue
			}

			if processInfo.CPU > me.CPUMaxLimit {
				// increment total attempts
				me.incrementAttempts()

				log.Printf("Process %s [%v]:[%v] is using %f%% of CPU, attempt: %v", me.NameRegex, processList[i].Pid(), processList[i].PPid(), processInfo.CPU, me.getAttempts())

				if me.getAttempts() > me.TotalAttemptsBeforeKill {

					if me.KillIfCPUMaxLimit {
						if process, err := os.FindProcess(processList[i].Pid()); err == nil {
							if err := process.Kill(); err != nil {
								log.Printf("Failed to kill process %s, error: %v", processList[i].Executable(), err.Error())
							} else {
								log.Printf("Killed process %s, PID: %v", processList[i].Executable(), processList[i].Pid())
							}
						} else {
							log.Printf("Failed to find process %s with PID: %v, error: %v", processList[i].Executable(), processList[i].Pid(), err.Error())
						}

						// reset total attempts after killing
						me.resetAttempts()
					}
				}
			} else {
				if me.getAttempts() > 0 {
					log.Printf("Reset %s [%v]:[%v], CPU: %f%%, attempt: %v", me.NameRegex, processList[i].Pid(), processList[i].PPid(), processInfo.CPU, me.getAttempts())
				}

				// CPU usage is below the limit, reset attempts
				me.resetAttempts()
			}
		}

	}

	if !found {
		me.resetAttempts()
	}

	return nil
}
