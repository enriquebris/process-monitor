package main

import (
	"github.com/robfig/cron/v3"
	"log"
	"regexp"
)

// *****************************************************************************
// MonitorEntry
// *****************************************************************************

type MonitorEntry struct {
	NameRegex               string `json:"name_regex"`
	regex                   *regexp.Regexp
	Cron                    string  `json:"cron"`
	CPUMaxLimit             float64 `json:"cpu_max_limit"`
	KillIfCPUMaxLimit       bool    `json:"kill_if_cpu_max_limit"`
	TotalAttemptsBeforeKill uint    `json:"total_attempts_before_kill"`
	totalAttemptsBeforeKill uint
}

// GetCronFunction returns a func() to be executed by cron
func (st *MonitorEntry) GetCronFunction() func() {
	return func() {
		if err := processChecker(st); err != nil {
			log.Printf("%s", err)
		}
	}
}

// resetAttempts resets the attempts counter
func (st *MonitorEntry) resetAttempts() {
	st.totalAttemptsBeforeKill = 0
}

func (st *MonitorEntry) incrementAttempts() {
	st.totalAttemptsBeforeKill++
}

func (st *MonitorEntry) getAttempts() uint {
	return st.totalAttemptsBeforeKill
}

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
			st.cron.AddFunc(entryList[i].Cron, entryList[i].GetCronFunction())
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
