<h1 align="center">🎵 Ajzak 2.0 — Discord Music Bot</h1>

<p align="center">
  <b>A modern, fast, and reliable Discord music bot built with a Go backend and a Node.js sidecar service.</b>
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

Ajzak 2.0 utilizes the <b>Sidecar Microservice Pattern</b>:

- <b>Go Backend (Core):</b> The core engine of the bot. It handles user commands, database operations, metadata extraction (`yt-dlp -j`), queue state management (`sync.Mutex` Queue system), and rich Discord Embed rendering.
- <b>Node.js Voice Sidecar:</b> Listens on `http://localhost:3000` as a network adapter for voice channels (`@discordjs/voice`). Handles E2EE (DAVE) encryption protocols and audio stream playback.
- <b>Go Webhook Event Listener:</b> Listens on `http://localhost:8080` to receive webhook notifications from the Node.js service (e.g., track completion) and automatically triggers playback for the next queued song.

---

<h2>✨ Features</h2>

- 🎶 <b>High-Quality Audio Playback:</b> Native audio streaming powered by `yt-dlp` and `ffmpeg`.
- 🖼️ <b>Compact Discord Embeds:</b> Clean displays featuring track title, uploader, duration, thumbnail, and requester tags.
- 📜 <b>Per-Guild Queue System:</b> Robust queue management supporting track additions, skipping, and queue previews.
- 🧹 <b>Smart Auto-Cleanup:</b>
  - Automatic voice channel disconnect when all human users leave.
  - Startup connection cleanup to clear stale sessions.
  - Safe connection teardown on process exit (`Graceful Shutdown`).

---

<h2>🛠️ Tech Stack & Prerequisites</h2>

The following dependencies must be installed and available in your system's `PATH`:

| Technology | Role |
| :--- | :--- |
| <img src="https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white" /> | Main bot logic, Queue system, and Command handlers |
| <img src="https://img.shields.io/badge/Node.js-339933?style=flat-square&logo=nodedotjs&logoColor=white" /> | Voice sidecar service (`@discordjs/voice`) |
| <img src="https://img.shields.io/badge/FFmpeg-007808?style=flat-square&logo=ffmpeg&logoColor=white" /> | Audio stream encoding and decoding |
| <img src="https://img.shields.io/badge/yt--dlp-FF0000?style=flat-square&logo=youtube&logoColor=white" /> | Extracting audio streams and video metadata |

---

<h2>🚀 Installation & Setup</h2>

### 1. Clone the Repository
```bash
git clone [https://github.com/markovic-dev/ajzak.git](https://github.com/markovic-dev/ajzak.git)
cd ajzak