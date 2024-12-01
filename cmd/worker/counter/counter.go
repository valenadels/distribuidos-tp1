package main

import (
	"tp1/internal/worker/aggregator"
	"tp1/pkg/logs"
)

func main() {
	filter, err := aggregator.NewCounter()
	if err != nil {
		logs.Logger.Errorf("Failed to create new reviews filter: %s", err.Error())
		return
	}

	if err = filter.Init(); err != nil {
		logs.Logger.Errorf("Failed to initialize new reviews filter: %s", err.Error())
		return
	}

	filter.Start()
}
