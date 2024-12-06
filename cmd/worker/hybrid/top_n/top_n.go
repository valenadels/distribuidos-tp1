package main

import (
	"tp1/internal/healthcheck"
	"tp1/internal/worker/hybrid/top_n"
	"tp1/pkg/logs"
)

func main() {
	hc, err := healthcheck.NewService()
	if err != nil {
		logs.Logger.Errorf("Failed to launch health checker: %s", err.Error())
		return
	}

	go hc.Listen() //TODO ver si sale con 0 al stoppear

	topNWorker, err := top_n.New()
	if err != nil {
		logs.Logger.Errorf("Failed to create new top N worker: %s", err.Error())
		return
	}

	if err = topNWorker.Init(); err != nil {
		logs.Logger.Errorf("Failed to initialize top N worker: %s", err.Error())
		return
	}

	topNWorker.Start()
}
