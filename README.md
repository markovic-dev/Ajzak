<h1 align="center">🎵 Ajzak 2.0 — Discord Music & Utility Bot</h1>

<p align="center">
  <b>A modern, fast, and reliable Discord bot built with a Go backend, SQLite storage, and a Node.js voice sidecar service.</b>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Node.js-339933?style=for-the-badge&logo=nodedotjs&logoColor=white" alt="Node.js" />
  <img src="https://img.shields.io/badge/Discord.js-5865F2?style=for-the-badge&logo=discord&logoColor=white" alt="Discord.js" />
  <img src="https://img.shields.io/badge/Express.js-000000?style=for-the-badge&logo=express&logoColor=white" alt="Express" />
  <img src="https://img.shields.io/badge/SQLite-003B57?style=for-the-badge&logo=sqlite&logoColor=white" alt="SQLite" />
  <img src="https://img.shields.io/badge/FFmpeg-007808?style=for-the-badge&logo=ffmpeg&logoColor=white" alt="FFmpeg" />
</p>

---

<h2>🏗️ Architecture Overview</h2>

Ajzak 2.0 utilizes the <b>Sidecar Microservice Pattern</b> to pair high-performance concurrent Go backend logic with Discord.js voice API capabilities:

- <b>Go Backend (Core Engine):</b> Handles user commands, Discord component interactions, metadata extraction (`yt-dlp`), queue state management (`sync.Mutex`), SQLite database operations, and rich Discord Embed rendering.
- <b>Node.js Voice Sidecar:</b> Listens on `http://localhost:3000` as a dedicated network adapter for voice channels (`@discordjs/voice`). Handles E2EE (DAVE) encryption protocols and audio stream playback.
- <b>Go Webhook Event Listener:</b> Listens on `http://localhost:8080` to receive webhook notifications from the Node.js service (e.g., track completion) and automatically triggers playback for the next queued song.
- <b>SQLite Database:</b> Stores quotes and tracks usage counts via generated `sqlc` queries.

---

<h2>✨ Features</h2>

- 🎶 <b>High-Quality Audio Playback:</b> Native audio streaming powered by `yt-dlp` and `ffmpeg`.
- 🔍 <b>Interactive YouTube Search:</b> Search songs by text query and choose from top results using a native <b>Discord Select Menu (Dropdown)</b> component.
- 📜 <b>Per-Guild Queue System:</b> Robust queue management supporting track additions, skipping, embed visualizers with durations, thumbnails, and requester tags.
- 🔂 <b>Track Replay / Loop Mode:</b> Repeat the current track indefinitely until toggled off, skipped, or stopped.
- ⏱️ <b>Smart Inactivity Idle Timer:</b> Automatically starts a 1-minute countdown when the queue empties or playback stops, disconnecting from the voice channel if no new songs are added.
- 💬 <b>Quotes & Statistics System ("Provale"):</b> Persistent SQLite storage to save, fetch, and track statistics for community quotes and jokes.
- 🧹 <b>Smart Auto-Cleanup:</b>
  - Automatic voice channel disconnect on idle timeout or when no human users remain.
  - Startup connection cleanup to clear stale voice sessions.
  - Safe connection teardown on process exit (`Graceful Shutdown`).

---

<h2>📜 Commands & Usage</h2>

### 🎵 Music Commands
| Command | Description |
| :--- | :--- |
| `.pusti <URL / query>` | Plays audio directly from a URL or opens an interactive dropdown menu with YouTube search results |
| `.skip` | Skips the current track and starts playing the next song in queue |
| `.replay` | Toggles repeat mode for the current track (automatically disabled on `.skip`, `.stop`, or `.leave`) |
| `.queue` | Displays current playing track and up to 10 upcoming tracks in queue |
| `.stop` | Stops playback, clears the queue, and starts the 1-minute idle timer |
| `.join` | Summons the bot to your current voice channel |
| `.leave` | Forces the bot to disconnect from the voice channel |

### 💬 Utility & Quotes Commands
| Command | Description |
| :--- | :--- |
| `.provala` | Fetches and displays a random/least used quote from the database |
| `.dodaj <text>` | Adds a new quote/provala to the SQLite database |
| `.stats` | Displays a formatted table with usage statistics for all quotes |

---

<h2>🛠️ Tech Stack & Prerequisites</h2>

The following dependencies must be installed and available in your system's `PATH`:

| Technology | Role |
| :--- | :--- |
| <img src="https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white" /> | Main core logic, Queue system, Database queries, Interactions |
| <img src="https://img.shields.io/badge/Node.js-339933?style=flat-square&logo=nodedotjs&logoColor=white" /> | Voice sidecar service (`@discordjs/voice`) |
| <img src="https://img.shields.io/badge/SQLite-003B57?style=flat-square&logo=sqlite&logoColor=white" /> | Database storage for quotes and usage statistics |
| <img src="https://img.shields.io/badge/FFmpeg-007808?style=flat-square&logo=ffmpeg&logoColor=white" /> | Audio stream encoding and decoding |
| <img src="https://img.shields.io/badge/yt--dlp-FF0000?style=flat-square&logo=youtube&logoColor=white" /> | Extracting audio streams and video metadata |

---

<h2>🚀 Installation & Setup</h2>

### 1. Clone the Repository
```bash
git clone [https://github.com/markovic-dev/ajzak.git](https://github.com/markovic-dev/ajzak.git)
cd ajzak
```

### 2. Configure Environment Variables
Create a `.env` file in the root directory:
```env
DISCORD_TOKEN=your_bot_token_here
PORT=8080
VOICE_SERVICE_URL=http://localhost:3000
```

### 3. Start Node.js Voice Sidecar
```bash
cd voice-service
npm install
node index.js
```

### 4. Run Go Backend
```bash
# Open a new terminal in the project root
go run main.go
```

---

<p align="center"><br/>
 This Discord bot is a tribute to — <b>Ajs Nigrutin (Ajzak)</b>.<br/>
</p>