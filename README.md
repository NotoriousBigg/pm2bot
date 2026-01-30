# PM2 Telegram Bot

A lightweight Telegram bot written in Go to manage your PM2 processes remotely.

## Features

- 📊 **List Processes**: View status, memory usage, and CPU consumption of all PM2 processes.
- 🔄 **Restart/Stop**: Control your applications directly from Telegram.
- 🛠️ **Command-line integration**: Easily manage your server without SSH.
- 🔒 **Admin restricted**: Only the specified Admin ID can control the bot.

## Prerequisites

- **Go**: 1.18 or higher.
- **PM2**: Must be installed and available in the system PATH.
- **Telegram Bot Token**: Obtainable from [@BotFather](https://t.me/BotFather).

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/NotoriousBigg/pm2bot.git
   cd pm2bot
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Setup environment variables:
   ```bash
   cp sample.env .env
   ```
   Edit `.env` and fill in your `BOT_TOKEN` and `ADMIN_ID`.

## Configuration

The bot uses a `.env` file for configuration:

- `BOT_TOKEN`: Your Telegram Bot API token.
- `ADMIN_ID`: Your Telegram User ID (you can get it from [@userinfobot](https://t.me/userinfobot)).
- `DEBUG`: (Optional) Set to `true` for verbose logging.

## Running the Bot

### Normal run:
```bash
go run .
```

### Run with PM2 (recommended):
To keep the bot running in the background:
```bash
pm2 start "go run ." --name "pm2-telegram-bot"
```

## Usage

Once the bot is running, send `/start` to your bot on Telegram. If your ID matches the `ADMIN_ID` in the configuration, you will see the control panel.

### Commands
- `/start` - Open the main control panel.
- `/status` - (Future/Internal) View process status.
- `/startapp <script> <name>` - Start a new process via PM2.

## License

MIT
