package main

import "log"

func main() {

	config, err := loadConfig("config.json")
	if err != nil {
		log.Printf("Error loading config: %s", err)
		return
	}

	// new Monitor
	monitor := NewMonitor()
	// add entries
	for i := 0; i < len(config.Entries); i++ {
		log.Printf("Adding entry: %s", config.Entries[i].NameRegex)
		monitor.AddEntry(&config.Entries[i])
	}

	monitor.Start()

	done := make(chan struct{})
	// wait for ctrl+c
	<-done
}
