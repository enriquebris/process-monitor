package main

import (
	"github.com/robfig/cron/v3"
	"log"
	"regexp"
	"strings"
	"sync"
)

// *****************************************************************************
// MonitorEntry
// *****************************************************************************

type MonitorEntry struct {
	// process(es) to monitor
	NameRegex string `json:"name_regex"`
	// how often to check
	Cron string `json:"cron"`
	// CPU max limit
	CPUMaxLimit float64 `json:"cpu_max_limit"`
	// whether to kill process(es) if CPU limit is reached (after TotalAttemptsBeforeKill attempts)
	KillIfCPUMaxLimit bool `json:"kill_if_cpu_max_limit"`
	// whether to kill the parent process if the sum of children attempts exceeds the TotalAttemptsBeforeKill limit
	KillIfChildrenCPUMaxLimit bool `json:"kill_if_children_cpu_max_limit"`
	// number of attempts to check before killing process(es) (only if KillIfCPUMaxLimit is true)
	TotalAttemptsBeforeKill int `json:"total_attempts_before_kill"`

	// processes keep track of processes matching NameRegex: [pid]total consecutive times exceeding CPUMaxLimit
	processes               sync.Map
	regex                   *regexp.Regexp
	totalAttemptsBeforeKill uint
}

func (st *MonitorEntry) Sanitize() {
	st.NameRegex = strings.TrimSpace(st.NameRegex)
	st.Cron = strings.TrimSpace(st.Cron)
}

// GetCronFunction returns a func() to be executed by cron
func (st *MonitorEntry) GetCronFunction() func() {
	return func() {
		if err := processChecker(st); err != nil {
			log.Printf("%s", err)
		}
	}
}

// resetAttemptsForPID resets the number of consecutive times a process has exceeded CPUMaxLimit (for a given PID)
func (st *MonitorEntry) resetAttemptsForPID(pid int) {
	st.processes.Store(pid, 0)
}

// removeAttemptsForAllPIDsNotInTheMap removes all processes having PIDs not in the given map
func (st *MonitorEntry) removeAttemptsForAllPIDsNotInTheMap(mp map[int]struct{}) {
	st.processes.Range(func(key, value interface{}) bool {
		if _, ok := mp[key.(int)]; !ok {
			st.processes.Delete(key)
		}
		return true
	})
}

// deprecated
// resetAttempts resets the attempts counter
func (st *MonitorEntry) resetAttempts() {
	st.totalAttemptsBeforeKill = 0
}

// incrementAttemptsForPID increments the total attempts counter for a given PID
func (st *MonitorEntry) incrementAttemptsForPID(pid int) {
	var actual int

	if actualRaw, ok := st.processes.Load(pid); ok {
		actual = actualRaw.(int)
	}

	st.processes.Store(pid, actual+1)
}

// deprecated
func (st *MonitorEntry) incrementAttempts() {
	st.totalAttemptsBeforeKill++
}

// getAttemptsForPID returns the number of consecutive times a process has exceeded CPUMaxLimit (for a given PID)
func (st *MonitorEntry) getAttemptsForPID(pid int) int {
	actualRaw, ok := st.processes.Load(pid)
	if !ok {
		return 0
	}
	return actualRaw.(int)
}

// getAllAttempts returns the number of consecutive times all processes have exceeded CPUMaxLimit
func (st *MonitorEntry) getAllAttempts() int {
	total := 0

	st.processes.Range(func(key, value interface{}) bool {
		total += value.(int)

		return true
	})

	return total
}

// deprecated
func (st *MonitorEntry) getAttempts() uint {
	return st.totalAttemptsBeforeKill
}

// getRegex returns the regex compiled from NameRegex
func (st *MonitorEntry) getRegex() *regexp.Regexp {
	if st.regex == nil {
		st.regex = regexp.MustCompile(st.NameRegex)
	}

	return st.regex
}

// *****************************************************************************
// Monitor
// *****************************************************************************

type Monitor struct {
	cron     *cron.Cron
	entryMap map[string][]*MonitorEntry
}

func NewMonitor() *Monitor {
	ret := &Monitor{}
	ret.initialize()

	return ret
}

func (st *Monitor) initialize() {
	st.cron = cron.New()
	st.entryMap = make(map[string][]*MonitorEntry)
}

func (st *Monitor) Start() {
	// TODO ::: get the context to wait until the running jobs (if any) are done
	st.cron.Stop()

	// add the entries
	for _, entryList := range st.entryMap {
		for i := 0; i < len(entryList); i++ {
			if _, err := st.cron.AddFunc(entryList[i].Cron, entryList[i].GetCronFunction()); err != nil {
				log.Printf("error adding cron job for %v: %v", entryList[i].NameRegex, err.Error())
			}
		}
	}

	st.cron.Start()
}

func (st *Monitor) AddEntry(entry *MonitorEntry) {
	if _, ok := st.entryMap[entry.NameRegex]; !ok {
		st.entryMap[entry.NameRegex] = []*MonitorEntry{entry}
	} else {
		st.entryMap[entry.NameRegex] = append(st.entryMap[entry.NameRegex], entry)
	}
}
