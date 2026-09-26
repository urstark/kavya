package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"kavya/bot"
	"kavya/config"
)

func main() {
	log.Println("Initializing Kavya Bot in Go...")

	cfg := config.LoadConfig()

	if cfg.BotToken == "" {
		log.Fatal("FATAL: BOT_TOKEN is required in .env or environment")
	}

	if cfg.GroqAPIKey == "" {
		log.Fatal("FATAL: GROQ_API_KEY is required in .env or environment")
	}

	b, err := bot.NewKavyaBot(cfg)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize bot: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received termination signal, shutting down bot...")
		_ = b.Close()
		os.Exit(0)
	}()

	b.Start()
}
