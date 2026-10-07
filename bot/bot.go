package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	tele "gopkg.in/telebot.v3"

	"kavya/ai"
	"kavya/config"
	"kavya/db"
	"kavya/stickers"
)

type KavyaBot struct {
	teleBot  *tele.Bot
	cfg      *config.Config
	aiClient *ai.GroqClient
	store    db.Store
	burst    *BurstManager
}

func NewKavyaBot(cfg *config.Config) (*KavyaBot, error) {
	pref := tele.Settings{
		Token:  cfg.BotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("telebot init error: %w", err)
	}

	aiClient := ai.NewGroqClient(cfg.GroqAPIKey, cfg.GroqModel, cfg.GroqVisionModel)

	var store db.Store
	if cfg.MongoURI != "" {
		log.Println("Connecting to MongoDB Atlas...")
		mongoStore, err := db.NewMongoStore(cfg.MongoURI, "kavya")
		if err != nil {
			return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
		}
		log.Println("Successfully connected to MongoDB (database: kavya)")
		store = mongoStore
	} else {
		log.Printf("No MONGO_URI provided, falling back to SQLite (%s)...", cfg.SQLitePath)
		sqliteStore, err := db.NewSQLiteStore(cfg.SQLitePath)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize sqlite store: %w", err)
		}
		store = sqliteStore
	}

	kb := &KavyaBot{
		teleBot:  b,
		cfg:      cfg,
		aiClient: aiClient,
		store:    store,
	}

	kb.burst = NewBurstManager(cfg.BurstDelaySec, kb.handleAggregatedMessage)
	kb.registerHandlers()
	return kb, nil
}

func (kb *KavyaBot) registerHandlers() {
	kb.teleBot.Handle("/start", func(c tele.Context) error {
		_ = c.Notify(tele.Typing)
		time.Sleep(1 * time.Second)
		userName := c.Sender().FirstName
		if userName == "" {
			userName = "friend"
		}
		greeting := fmt.Sprintf("hey %s what's up", ai.SanitizeText(userName))
		_ = kb.store.SaveMessage(c.Chat().ID, "assistant", greeting)
		return c.Send(greeting)
	})

	kb.teleBot.Handle("/reset", func(c tele.Context) error {
		_ = kb.store.ClearHistory(c.Chat().ID)
		return c.Send("history wiped let's start fresh")
	})

	kb.teleBot.Handle("/ping", func(c tele.Context) error {
		return c.Send(fmt.Sprintf("%s is online and running smooth", kb.cfg.BotName))
	})

	kb.teleBot.Handle(tele.OnText, func(c tele.Context) error {
		kb.burst.AddMessage(c, c.Text())
		return nil
	})

	kb.teleBot.Handle(tele.OnSticker, func(c tele.Context) error {
		st := c.Message().Sticker
		alt := "sticker"
		if st != nil && st.Emoji != "" {
			alt = st.Emoji
		}
		desc := fmt.Sprintf("[user sent a sticker with emoji: %s]", alt)
		kb.burst.AddMessage(c, desc)
		return nil
	})

	kb.teleBot.Handle(tele.OnPhoto, func(c tele.Context) error {
		photo := c.Message().Photo
		if photo == nil {
			return nil
		}

		_ = c.Notify(tele.Typing)
		fileRC, err := kb.teleBot.File(&photo.File)
		if err != nil {
			log.Printf("Error downloading photo: %v", err)
			return c.Send("couldn't open that photo for some reason")
		}
		defer fileRC.Close()

		photoBytes, err := io.ReadAll(fileRC)
		if err != nil {
			log.Printf("Error reading photo bytes: %v", err)
			return c.Send("couldn't read that photo")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		visionDesc, err := kb.aiClient.DescribeImage(ctx, photoBytes)
		if err != nil {
			log.Printf("Vision description error: %v", err)
			visionDesc = "an image"
		}

		caption := c.Message().Caption
		var userPrompt string
		if caption != "" {
			userPrompt = fmt.Sprintf("[user sent a photo showing: %s with caption: '%s']", visionDesc, caption)
		} else {
			userPrompt = fmt.Sprintf("[user sent a photo showing: %s]", visionDesc)
		}

		kb.burst.AddMessage(c, userPrompt)
		return nil
	})

	kb.teleBot.Handle(tele.OnAnimation, func(c tele.Context) error {
		kb.burst.AddMessage(c, "[user sent a GIF/animation]")
		return nil
	})

	kb.teleBot.Handle(tele.OnVoice, func(c tele.Context) error {
		kb.burst.AddMessage(c, "[user sent a voice note]")
		return nil
	})

	kb.teleBot.Handle(tele.OnVideo, func(c tele.Context) error {
		kb.burst.AddMessage(c, "[user sent a video]")
		return nil
	})

	kb.teleBot.Handle(tele.OnDocument, func(c tele.Context) error {
		doc := c.Message().Document
		fileName := "a document"
		if doc != nil && doc.FileName != "" {
			fileName = doc.FileName
		}
		kb.burst.AddMessage(c, fmt.Sprintf("[user sent a document: %s]", fileName))
		return nil
	})
}

func (kb *KavyaBot) handleAggregatedMessage(c tele.Context, aggregatedText string) {
	chatID := c.Chat().ID
	msgID := c.Message().ID
	userName := c.Sender().FirstName
	if userName == "" {
		userName = c.Sender().Username
	}

	_ = c.Notify(tele.Typing)

	typingDuration := time.Duration(kb.cfg.TypingDelaySec * float64(time.Second))
	if typingDuration > 0 {
		time.Sleep(typingDuration)
	}

	_ = kb.store.SaveMessage(chatID, "user", aggregatedText)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	systemPrompt := ai.BuildSystemPrompt(kb.cfg.BotName, userName)
	chatHist, err := kb.store.GetHistory(chatID, kb.cfg.HistoryLimit)
	if err != nil {
		log.Printf("Error retrieving chat history: %v", err)
	}

	aiResp, err := kb.aiClient.GenerateResponse(ctx, systemPrompt, chatHist)
	if err != nil {
		log.Printf("AI error for chat %d: %v", chatID, err)
		_ = c.Send("mera mood thoda off hai abhi")
		return
	}

	if aiResp.ReactionEmoji != "" && aiResp.ReactionEmoji != "no_reaction" {
		go kb.sendReaction(chatID, msgID, aiResp.ReactionEmoji)
	}

	replyText := aiResp.ReplyText
	hasText := replyText != "" && replyText != "no_output"

	if hasText {
		_ = kb.store.SaveMessage(chatID, "assistant", replyText)
		_ = c.Notify(tele.Typing)
		if err := c.Send(replyText); err != nil {
			log.Printf("Error sending text reply: %v", err)
		}
	}

	if aiResp.StickerCategory != "" && aiResp.StickerCategory != "no_sticker" {
		stickerID := stickers.GetStickerForCategory(aiResp.StickerCategory)
		if stickerID != "" {
			_ = c.Notify(tele.ChoosingSticker)
			time.Sleep(1 * time.Second)
			st := &tele.Sticker{File: tele.File{FileID: stickerID}}
			if err := c.Send(st); err != nil {
				log.Printf("Error sending sticker reply: %v", err)
			}
		}
	}

	if aiResp.SendVoiceNote != "" && aiResp.SendVoiceNote != "no_voice" {
		_ = kb.store.SaveMessage(chatID, "assistant", "[Voice Note]: "+aiResp.SendVoiceNote)

		ttsSuccess := false
		if kb.cfg.ElevenLabsAPIKey != "" {
			_ = c.Notify(tele.RecordingAudio)
			time.Sleep(2 * time.Second)
			audioBytes, err := ai.GenerateElevenLabsTTS(ctx, kb.cfg.ElevenLabsAPIKey, kb.cfg.ElevenLabsVoiceID, aiResp.SendVoiceNote)
			if err == nil {
				voice := &tele.Voice{File: tele.FromReader(bytes.NewReader(audioBytes))}
				if err := c.Send(voice); err == nil {
					ttsSuccess = true
				} else {
					log.Printf("Error sending voice note: %v", err)
				}
			} else {
				log.Printf("ElevenLabs TTS error: %v", err)
			}
		} else if kb.cfg.GeminiAPIKey != "" {
			_ = c.Notify(tele.RecordingAudio)
			time.Sleep(2 * time.Second)
			audioBytes, err := ai.GenerateTTS(ctx, kb.cfg.GeminiAPIKey, aiResp.SendVoiceNote)
			if err == nil {
				voice := &tele.Voice{File: tele.FromReader(bytes.NewReader(audioBytes))}
				if err := c.Send(voice); err == nil {
					ttsSuccess = true
				} else {
					log.Printf("Error sending voice note: %v", err)
				}
			} else {
				log.Printf("Gemini TTS error: %v", err)
			}
		} else {
			log.Printf("No API key configured for TTS")
		}

		if !ttsSuccess {
			// Fallback: send as text if audio generation or sending failed
			_ = c.Send("*(Voice Note Failed)*\n" + aiResp.SendVoiceNote)
		}
	}

	if aiResp.SendImagePrompt != "" && aiResp.SendImagePrompt != "no_image" {
		_ = kb.store.SaveMessage(chatID, "assistant", "[Image Sent]: "+aiResp.SendImagePrompt)
		_ = c.Notify(tele.UploadingPhoto)
		time.Sleep(2 * time.Second)
		encodedPrompt := url.QueryEscape(aiResp.SendImagePrompt)
		imageURL := fmt.Sprintf("https://image.pollinations.ai/prompt/%s?width=1024&height=1024&nologo=true", encodedPrompt)
		photo := &tele.Photo{File: tele.FromURL(imageURL)}
		if err := c.Send(photo); err != nil {
			log.Printf("Error sending generated image: %v", err)
		}
	}
}

func (kb *KavyaBot) sendReaction(chatID int64, messageID int, emoji string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setMessageReaction", kb.cfg.BotToken)
	payload := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
		"reaction": []map[string]string{
			{"type": "emoji", "emoji": emoji},
		},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}

func (kb *KavyaBot) Close() error {
	return kb.store.Close()
}

func (kb *KavyaBot) Start() {
	log.Printf("Starting %s Telegram bot (@%s)...", kb.cfg.BotName, kb.teleBot.Me.Username)
	kb.teleBot.Start()
}
