<h1 align="center">🎤 Ajzak Discord Bot (v1.0)</h1>

<p align="center">
  A lightweight Discord bot written in Go that serves random quotes from a SQLite database using <code>discordgo</code>, <code>sqlc</code>, and <code>goose</code>.
</p>

---

## 🛠️ Tech Stack

* **Language:** [Go (Golang)](https://go.dev/)
* **Discord API:** [`discordgo`](https://github.com/bwmarrin/discordgo)
* **Database:** [SQLite](https://sqlite.org/) via [`modernc.org/sqlite`](https://gitlab.com/cznic/sqlite) (Pure-Go driver, no CGO required)
* **SQL Code Generation:** [`sqlc`](https://sqlc.dev/) (Type-safe SQL queries)
* **Database Migrations:** [`goose`](https://github.com/pressly/goose)
* **Environment Management:** [`godotenv`](https://github.com/joho/godotenv)

---

## 🧠 How the Rotation Algorithm Works

Rather than selecting a completely random quote from the entire database every time, the bot ensures an even distribution of responses:

1. Uses `sqlc` to fetch the **Top 5 least used quotes** from the SQLite database (`ORDER BY usage_count ASC LIMIT 5`).
2. Randomly picks one quote out of those 5 using `math/rand/v2`.
3. Sends the selected quote to the Discord channel.
4. Atomically increments the `usage_count` for that quote in the database.

---

## 📂 Project Structure

```text
.
├── bot/
│   └── bot.go          # Core Discord bot setup, handlers, and initialization
├── db/
│   ├── migrations/     # Goose SQL migrations (.sql files)
│   └── queries/        # SQL queries read by sqlc (.sql)
├── internal/
│   └── database/       # Auto-generated Go code produced by sqlc
├── .env                # Environment configuration (tokens)
├── .gitignore
├── ajzak.db            # SQLite database file (auto-generated)
├── go.mod
├── go.sum
├── main.go             # Entry point (calls bot.Start())
└── sqlc.yaml           # Configuration file for sqlc generator
```

---

## 🚀 Getting Started Locally

### 1. Prerequisites
* [Go](https://go.dev/dl/) installed (version 1.22 or newer)
* Discord Bot Token (configured in the [Discord Developer Portal](https://discord.com/developers/applications) with **Message Content Intent** enabled)

### 2. Clone the Repository & Install Dependencies
```bash
git clone [https://github.com/markovic-dev/ajzak.git](https://github.com/markovic-dev/ajzak.git)
cd ajzak
go mod download
```

### 3. Configure Environment Variables
Create a `.env` file in the root directory:
```env
DISCORD_TOKEN=your_discord_bot_token_here
```

### 4. Generate SQL Code (`sqlc`)
If you modify SQL queries in `db/queries/`, run:
```bash
sqlc generate
```

### 5. Run Database Migrations (`goose`)
To initialize the SQLite database schema and seed initial data, run:
```bash
goose -dir db/migrations sqlite3 ajzak.db up
```

### 6. Start the Bot
```bash
go run main.go
```

---

## 🎮 Commands

| Command | Description |
| :--- | :--- |
| `.provala` | Randomly picks one of the least used quotes from the database, posts it, and increments its usage count. |
| `.stats` | Displays a formatted table of all quotes, their IDs, and usage counts. |
| `.dodaj <text>` | Adds a new quote directly to the SQLite database. |

---

## 🛡️ Graceful Shutdown

The bot intercepts system signals (`SIGINT`, `SIGTERM`). When shutting down (e.g., pressing `CTRL+C`), it cleanly closes the WebSocket session with Discord and disconnects from the SQLite database to prevent data corruption.

---

## 🎤 Tribute

This project is a fun, fan-made tribute to **Ajs Nigrutin** and his iconic punchlines, humor, and hip-hop legacy.