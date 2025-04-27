package main

import (
	"time"
	"yupi/internal/service/agent"
)

func main() {

	agent := agent.NewAgent(
		"http://localhost:8080",
		2*time.Second,
		10*time.Second,
	)
	agent.Run()
}
