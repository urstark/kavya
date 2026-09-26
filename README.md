# 🌸 Kavya (Golang)

A high-performance, ultra-lightweight Telegram bot engineered in **Go (Golang)** with `telebot.v3`, the **Groq API** (`openai/gpt-oss-20b`), and **MongoDB Atlas** persistent storage.

Kavya is configured with an authentic, dynamic **19-year-old Indian Gen Z college girl** persona—sharp, witty, effortlessly cool, texting in casual Hinglish/English without canned scripts.

---

## ⚡ Key Highlights & Architecture

* **Standalone Go Binary:** Single compiled binary (~14MB), runtime memory **<25MB RAM** (compared to 1–2GB in traditional Python setups). Starts up in milliseconds.
* **MongoDB Atlas Persistence:** Direct integration with your MongoDB Atlas cluster (`kavya.messages`). Automatically creates compound indexes (`{chat_id: 1, _id: -1}`) for sub-millisecond history retrieval across reboots.
* **Organic Gen Z Persona (Kavya):**
  * 19-year-old college girl living in India.
  * Effortless blend of English and Hinglish slang (*"yaar"*, *"ngl"*, *"fr"*, *"scene"*, *"vibe"*, *"arre"*, *"chal na"*, *"bro"*, *"lmaoo"*).
  * Independent dynamic reasoning: NO repetitive canned scripts. Dynamic, witty comebacks when faced with creeps or boundary-crossers.
  * Multilingual mirroring when spoken to in Hindi, Urdu, French, Spanish, etc.
* **Clean Texting Sanitizer:** Formats output for that relaxed Gen Z Telegram texting aesthetic (lowercased flow, trimmed corporate periods).
* **Human Pacing & Burst Aggregator:**
  * Emulates natural typing delays (1.8s) via `tele.Typing`.
  * Coalesces rapid sequential user texts sent within 2.2 seconds into a single prompt before generating a reply.
* **Cloud Vision Pipeline:** Offloads user photo analysis to Groq vision—zero local PyTorch/BLIP overhead required.
* **Stickers & Reactions:** Contextual Telegram reactions (`setMessageReaction`) and 150+ categorized stickers.

---

## Project Structure

```
kavya/
├── ai/
│   ├── groq.go         # Groq client (structured function calling & vision)
│   ├── prompt.go       # Organic Kavya personality & system instructions
│   ├── sanitizer.go    # Casual lowercased texting formatting
│   └── sanitizer_test.go
├── bot/
│   ├── bot.go          # Telebot routing, typing latency, photo & sticker handlers
│   └── burst.go        # Thread-safe message burst aggregator
├── config/
│   └── config.go       # Minimal config loader with sensible Go defaults
├── db/
│   ├── store.go        # Storage interface (SaveMessage, GetHistory, ClearHistory)
│   ├── mongo.go        # MongoDB Atlas storage implementation
│   └── sqlite.go       # Local SQLite storage fallback
├── stickers/
│   └── stickers.go     # Sticker catalog & category lookup (150+ stickers)
├── .env.example        # Minimal 3-line configuration template
├── Dockerfile          # Multi-stage production container build
├── go.mod / go.sum     # Go dependency definitions
└── main.go             # Application entrypoint with graceful shutdown
```

---

## Quick Start

### 1. Requirements
* Go 1.22+ (tested on Go 1.26.0)
* Telegram Bot Token (from [@BotFather](https://t.me/BotFather))
* Groq API Key (from [console.groq.com](https://console.groq.com))
* MongoDB Connection String (Atlas or self-hosted)

### 2. Configuration
Copy `.env.example` to `.env` and fill in your keys:
```bash
cp .env.example .env
```

```env
BOT_TOKEN=your_telegram_bot_token_from_botfather
GROQ_API_KEY=your_groq_api_key_from_console_groq_com
MONGO_URI=mongodb+srv://<username>:<password>@cluster0.mongodb.net/
```

### 3. Run Locally
```bash
go run .
```

### 4. Build Standalone Binary
```bash
go build -ldflags="-s -w" -o kavya .
./kavya
```

### 5. Docker Deployment
```bash
docker build -t kavya .
docker run -d --name kavya --env-file .env kavya
```

---

## Developer Credits
* **Developer:** urstarkz
* **Website:** [urstark.is-a.dev](https://urstark.is-a.dev)
* **Telegram:** [t.me/urstarkz](https://t.me/urstarkz)

---

## 📜 License
This project is licensed under the [GNU General Public License v3.0](LICENSE).
