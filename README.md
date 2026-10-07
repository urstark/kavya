# 🌸 Kavya (Golang)

A high-performance, ultra-lightweight Telegram bot engineered in **Go (Golang)** with `telebot.v3`, the **Groq API** (powered by Qwen 27B), **ElevenLabs** for ultra-realistic TTS, and **MongoDB Atlas** persistent storage.

Kavya is configured with an authentic, dynamic **21-year-old Indian college girl** persona - sharp, witty, warm, texting in casual Hinglish/English without canned scripts, and speaking with a natural expressive voice.

---

## ⚡ Key Highlights & Architecture

* **Standalone Go Binary:** Single compiled binary (~14MB), runtime memory **<25MB RAM** (compared to 1-2GB in traditional Python setups). Starts up in milliseconds.
* **Intelligent AI Engine (Qwen 27B on Groq):** Blazing fast responses using state-of-the-art open models that strictly follow complex system instructions and JSON schema tool calling.
* **Expressive Voice Notes (ElevenLabs):** Generates ultra-realistic, highly expressive voice notes with laughs, sighs, and natural pacing when appropriate. Falls back to Gemini TTS if needed.
* **MongoDB Atlas Persistence:** Direct integration with your MongoDB Atlas cluster (`kavya.messages`). Automatically creates compound indexes (`{chat_id: 1, _id: -1}`) for sub-millisecond history retrieval across reboots.
* **Organic Persona (Kavya):**
  * 21-year-old college girl living in India.
  * Effortless blend of English and polite Hinglish ("tum", "aap").
  * Independent dynamic reasoning: NO repetitive canned scripts. Dynamic, witty comebacks when faced with creeps or boundary-crossers.
  * Multilingual mirroring when spoken to in pure Hindi, Urdu, French, Spanish, etc.
* **Human Pacing & Burst Aggregator:**
  * Emulates natural typing and recording delays via Telegram Chat Actions.
  * Coalesces rapid sequential user texts sent within 2.2 seconds into a single prompt before generating a reply.
* **Cloud Vision Pipeline:** Offloads user photo analysis to Qwen Vision on Groq (or Gemini Vision fallback) - zero local overhead required.
* **Stickers & Reactions:** Contextual Telegram reactions (`setMessageReaction`) and categorized stickers.

---

## Project Structure

```text
kavya/
├── ai/
│   ├── groq.go         # Groq client (structured JSON schema calling & vision)
│   ├── prompt.go       # Organic Kavya personality & system instructions
│   ├── sanitizer.go    # Casual texting formatting
│   ├── sanitizer_test.go
│   └── tts.go          # ElevenLabs and Gemini TTS generation
├── bot/
│   ├── bot.go          # Telebot routing, chat actions, photo, voice & sticker handlers
│   └── burst.go        # Thread-safe message burst aggregator
├── config/
│   └── config.go       # Configuration loader with Go defaults
├── db/
│   ├── store.go        # Storage interface (SaveMessage, GetHistory, ClearHistory)
│   ├── mongo.go        # MongoDB Atlas storage implementation
│   └── sqlite.go       # Local SQLite storage fallback
├── stickers/
│   └── stickers.go     # Sticker catalog & category lookup
├── .env.example        # Configuration template
├── Dockerfile          # Multi-stage production container build
├── go.mod / go.sum     # Go dependency definitions
└── main.go             # Application entrypoint with graceful shutdown
```

---

## Quick Start

### 1. Requirements
* Go 1.22+
* Telegram Bot Token (from @BotFather)
* Groq API Key (from console.groq.com)
* ElevenLabs API Key (from elevenlabs.io)
* Gemini API Key (Optional fallback)
* MongoDB Connection String (Atlas or self-hosted)

### 2. Configuration
Copy `.env.example` to `.env` and fill in your keys:
```bash
cp .env.example .env
```

```env
BOT_TOKEN=your_telegram_bot_token
GROQ_API_KEY=your_groq_api_key
MONGO_URI=mongodb+srv://<username>:<password>@cluster0.mongodb.net/
ELEVENLABS_API_KEY=your_elevenlabs_api_key
ELEVENLABS_VOICE_ID=your_custom_voice_id
GEMINI_API_KEY=your_gemini_api_key
```

### 3. Run Locally
```bash
go run .
```

### 4. Build Standalone Binary
```bash
go build -o kavya_bot .
./kavya_bot
```

---

## Developer Credits
* **Developer:** urstarkz
* **Website:** urstark.is-a.dev
* **Telegram:** t.me/urstarkz

---

## License
This project is licensed under the GNU General Public License v3.0 (LICENSE).
