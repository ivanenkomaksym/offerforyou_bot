# offerforyou_bot
Offer For You Telegram Bot

# Run locally

Telegram requires an HTTPS URL for webhooks. Your local machine, by default, serves over HTTP. So, to run your Telegram bot locally using webhooks, you need a way to expose your local HTTP server to the internet via an HTTPS URL. This is where tunneling services come in.

1. Start ngrok
```sh
ngrok http 8080 --host-header=localhost
```
2. Set `WEBHOOK_URL` in your `.env` file or environment variable
The `WEBHOOK_URL` you need to set will be the `https://` forwarding URL provided by ngrok, followed by your webhook path (`/telegram-webhook`).
```sh
WEBHOOK_URL=https://xxxxxxxxxxxx.ngrok-free.app/telegram-webhook
```
3. Run your Go bot locally
```sh
go run main.go
```
4. Use telegram bot

This setup will allow Telegram to send updates to your ngrok public URL, which then tunnels them securely to your Go bot running on `localhost:8080`.

If you want to have a custom domain that persists between launches, you can user ngrok-provided static domain

1. Run
```sh
ngrok http 8080 --host-header=localhost --domain=mybotdev.ngrok-free.app
```
*Replace `mybotdev.ngrok-free.app` with your actual reserved static domain.*

2. Update `WEBHOOK_URL` in `.env`:
```sh
WEBHOOK_URL=https://mybotdev.ngrok-free.app/telegram-webhook
```