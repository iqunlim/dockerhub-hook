package main

import (
	"log/slog"
	"os"
)


func main() {

	config := GetConfig()
	loglevel, err := ConvertLogLevel(config["LOG_LEVEL"].First())
	if err != nil {
		panic(err)
	} 
	logger = slog.New(slog.NewTextHandler(os.Stdout, loglevel))
	logger.Info("Docker Hub Hooks v0.2")
	StartWebHandler(config)
}
