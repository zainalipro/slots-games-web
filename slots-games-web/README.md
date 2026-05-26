# Slotopol Web 🎰

Modern web interface for the [slotopol/server](https://github.com/slotopol/server) slot game engine.

Browse 170+ slot games across 10 providers, play them with animated reels, and track your balance — all through a sleek casino-themed UI.

## Features

- **Game Browser** — Browse all games by provider with search and filter
- **Slot Machine** — Play any game with animated reels and win highlights
- **Authentication** — Login with demo accounts (player/dealer/admin)
- **Wallet** — Track your balance in real-time

## Quick Start

1. Make sure the [slotopol server](https://github.com/slotopol/server) is running on port 8080
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start the web app:
   ```bash
   npm start
   ```
4. Open http://localhost:3000

### Demo Accounts

| Email | Password | Role |
| :--- | :--- | :--- |
| player@example.org | iVI05M | Player |
| dealer@example.org | LtpkAr | Dealer |
| admin@example.org | 0YBoaT | Admin |

## Architecture

- **Frontend:** Vanilla HTML/CSS/JS SPA with dark casino theme
- **Backend:** Express.js server that serves static files and proxies API calls to the slotopol Go server
- **API Proxy:** All `/api/*` requests are forwarded to `http://localhost:8080/*`

## Configuration

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `3000` | Web server port |
| `API_TARGET` | `http://localhost:8080` | Slotopol server URL |
