package main

import (
	"Replicated-Blockchain/client/wuu_bernstein"
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"

	log "github.com/Sirupsen/logrus"
)

func configureLogger(level string) {
	log.SetOutput(os.Stderr)
	switch strings.ToLower(level) {
	case "panic":
		log.SetLevel(log.PanicLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warning", "warn":
		log.SetLevel(log.WarnLevel)
	}
}

func main() {
	var (
		logLevel string
		clientId int
	)

	flag.StringVar(&logLevel, "level", "DEBUG", "Set wuu_bernstein level.")
	flag.IntVar(&clientId, "client_id", 1, "client ID(1, 2, 3)")
	flag.Parse()

	configureLogger(logLevel)

	ctx, cancel := context.WithCancel(context.Background())
	wuu_bernstein.Client = wuu_bernstein.NewClient(ctx, clientId)
	wuu_bernstein.Client.Start(ctx)

	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			log.Info("Received an interrupt, stopping all connections...")
			cancel()
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}
