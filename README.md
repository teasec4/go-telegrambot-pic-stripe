# GoBotCat

Telegram bot for selling photos via Stripe payments.

## Installation

```bash
go mod tidy
```

## Configuration

1. Copy `.env.example` to `.env`:
```bash
cp .env.example .env
```

2. Fill in the variables:
```
STRIPE_PUBLISHABLE_KEY=pk_test_...
STRIPE_SECRET_KEY=sk_test_...
TELEGRAM_BOT_TOKEN=123456:ABC...
WEBHOOK_URL=https://yourdomain.com
PORT=8080
```

## Running

```bash
go run main.go
```

## Local Stripe webhook testing

In a separate terminal:
```bash
stripe listen --forward-to localhost:8080/webhook/stripe
```

## Usage

- `/start` — menu
- `/pay` — payment link
- Send photo — saves to `photos.json`

## Project Structure

```
config/     — configuration
handlers/   — request handlers (bot, webhook)
services/   — business logic (Stripe, Telegram)
storage/    — storage layer (JSON for photos, SQLite for payments)
```

## Databases

- `photos.json` — list of photo file_ids by user
- `payments.db` — payment history

## Deployment

Deploy to Railway, Render or any hosting:
1. Ensure Go 1.25+ is installed
2. Set environment variables
3. `go run main.go`
