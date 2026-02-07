# BigBrother

Simple Telegram bot in Go that accepts `.xlsx` files and processes them.

## Local run (long polling)

1. Create a bot with BotFather and copy the token.
2. Export your token:
   - `export TELEGRAM_BOT_TOKEN="YOUR_TOKEN"`
3. Run the bot:
   - `make run` (builds and runs the binary)
   - or `go run ./cmd/bigbrother`
4. Open your bot in Telegram, send `/start`, and upload an `.xlsx` file.

Uploaded files are stored under `./data/incoming/` by default.

## Environment variables

- `TELEGRAM_BOT_TOKEN` (required)
- `DATA_DIR` (default: `./data`)
- `BOT_MODE` (default: `polling`)
- `BOT_PUBLIC_URL` (required for `webhook` mode; not implemented yet)
- `MAX_FILE_BYTES` (default: `26214400` = 25 MiB)
- `MAX_DOCS_PER_MINUTE_CHAT` (default: `6`)

Env files are loaded automatically: `.env.local` takes priority, otherwise `.env` is used.
Existing environment variables are not overridden.

## Webhook mode (for later Hetzner deployment)

Webhook mode will require a **public HTTPS URL** that you control (Telegram does not provide this).
When you move to Hetzner, the typical setup is:

- Use a domain name that points to your server
- Terminate HTTPS with Nginx/Caddy/Traefik
- Expose port 443 to the internet
- Set `BOT_MODE=webhook` and `BOT_PUBLIC_URL=https://your-domain`

For now, the code runs in polling mode only; webhook mode will be added later.
