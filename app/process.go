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

	// foundPIDMp saves the PIDs of the processes that match the given MonitorEntry
	foundPIDMp := map[int]struct{}{}
	parentPID := 0
	parentExecutableName := ""

	// find the process
	for i := 0; i < len(processList); i++ {

		//if strings.TrimSpace(processList[i].Executable()) == me.NameRegex {
		if me.getRegex().MatchString(strings.TrimSpace(processList[i].Executable())) {
			// register the PID
			foundPIDMp[processList[i].Pid()] = struct{}{}

			processInfo, err := pidusage.GetStat(processList[i].Pid())
			if err != nil {
				log.Println(err)
				continue
			}
			// get parent PID / executable name
			if processList[i].PPid() == 1 {
				parentPID = processList[i].Pid()
			} else {
				parentPID = processList[i].PPid()
			}
			parentExecutableName = processList[i].Executable()

			if processInfo.CPU >= me.CPUMaxLimit {
				// increment total attempts
				me.incrementAttemptsForPID(processList[i].Pid())

				log.Printf("Process %s [%v]:[%v] is using %f%% of CPU, attempt: %v", processList[i].Executable(), processList[i].Pid(), processList[i].PPid(), processInfo.CPU, me.getAttemptsForPID(processList[i].Pid()))

				if me.getAttemptsForPID(processList[i].Pid()) > me.TotalAttemptsBeforeKill {
					if me.KillIfCPUMaxLimit {
						// kill the parent process
						killProcess(parentPID, parentExecutableName)

						// reset total attempts after killing
						me.resetAttemptsForPID(processList[i].Pid())
					}
				}
			} else {
				if me.getAttemptsForPID(processList[i].Pid()) > 0 {
					log.Printf("Reset %s [%v]:[%v], CPU: %f%%, attempt: %v", processList[i].Executable(), processList[i].Pid(), processList[i].PPid(), processInfo.CPU, me.getAttempts())
				}

				// CPU usage is below the limit, reset attempts
				me.resetAttemptsForPID(processList[i].Pid())
			}
		}

	}

	// remove PID entries that are no longer running
	me.removeAttemptsForAllPIDsNotInTheMap(foundPIDMp)

	// kill the parent process if the sum of children attempts exceeds the total allowed attempts
	if me.KillIfChildrenCPUMaxLimit && me.getAllAttempts() > me.TotalAttemptsBeforeKill {
		log.Printf("Kill process %v, PID: %v due to sum of children attempts exceeding the total allowed attempts", parentExecutableName, parentPID)
		killProcess(parentPID, parentExecutableName)
	}

	return nil
}

func killProcess(pid int, executableName string) {
	if process, err := os.FindProcess(pid); err == nil {
		if err := process.Kill(); err != nil {
			log.Printf("Failed to kill process %s, error: %v", executableName, err.Error())
		} else {
			log.Printf("Killed process %s, PID: %v", executableName, pid)
		}
	} else {
		log.Printf("Failed to find process %s with PID: %v, error: %v", executableName, pid, err.Error())
	}
}
