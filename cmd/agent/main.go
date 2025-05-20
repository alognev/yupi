package main

import (
	"log"
	"yupi/internal/config"
	"yupi/internal/service/agent"
)

func main() {
	cfg, err := config.SetConfig()

	if err != nil {
		log.Fatal(err)
	}

	myAgent := agent.NewAgent(
		cfg.ServerAddr,
		cfg.PollInterval,
		cfg.ReportInterval,
	)
	myAgent.Run()
}
