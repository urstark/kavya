package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Essential credentials (from .env or environment)
	BotToken          string
	GroqAPIKey        string
	GeminiAPIKey      string
	ElevenLabsAPIKey  string
	ElevenLabsVoiceID string
	MongoURI          string

	// Persona settings (sensibly defaulted in code)
	BotName string

	// AI models (sensibly defaulted in code)
	GroqModel       string
	GroqVisionModel string

	// Pacing & context limits (sensibly defaulted in code)
	BurstDelaySec  float64
	TypingDelaySec float64
	HistoryLimit   int

	// Database fallback (if MongoURI is empty)
	SQLitePath string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found or reading from environment")
	}

	botToken := os.Getenv("BOT_TOKEN")
	groqKey := os.Getenv("GROQ_API_KEY")
	geminiKey := os.Getenv("GEMINI_API_KEY")
	elevenKey := os.Getenv("ELEVENLABS_API_KEY")
	elevenVoice := os.Getenv("ELEVENLABS_VOICE_ID")
	if elevenVoice == "" {
		elevenVoice = "21m00Tcm4TlvDq8ikWAM" // Default: Rachel
	}
	mongoURI := os.Getenv("MONGO_URI")

	// Allow setting BOT_NAME if desired, default is Kavya
	botName := os.Getenv("BOT_NAME")
	if botName == "" {
		botName = "Kavya"
	}

	groqModel := os.Getenv("GROQ_MODEL")
	if groqModel == "" {
		groqModel = "qwen/qwen3.8-27b"
	}

	groqVisionModel := os.Getenv("GROQ_VISION_MODEL")
	if groqVisionModel == "" {
		groqVisionModel = "qwen/qwen3.8-27b"
	}

	return &Config{
		BotToken:          botToken,
		GroqAPIKey:        groqKey,
		GeminiAPIKey:      geminiKey,
		ElevenLabsAPIKey:  elevenKey,
		ElevenLabsVoiceID: elevenVoice,
		MongoURI:          strings.TrimSpace(mongoURI),
		BotName:           botName,
		GroqModel:         groqModel,
		GroqVisionModel:   groqVisionModel,
		BurstDelaySec:     2.2,
		TypingDelaySec:    1.8,
		HistoryLimit:      20,
		SQLitePath:        "data/kavya.db",
	}
}
