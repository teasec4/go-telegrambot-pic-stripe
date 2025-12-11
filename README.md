# GoBotCat

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)
[![Telegram Bot](https://img.shields.io/badge/Telegram-Bot-0088cc)](https://telegram.org)

Telegram bot for selling photos via Stripe and Tron blockchain payments.

## Quick Start

### 1. Setup Environment Variables

Create `.env` file with required variables:

```env
# Telegram
TELEGRAM_BOT_TOKEN=your_bot_token_here

# Stripe (optional, for production)
STRIPE_PUBLISHABLE_KEY=pk_test_...
STRIPE_SECRET_KEY=sk_test_...

# Tron Network
TRON_API_KEY=your_trongrid_api_key

# Server
WEBHOOK_URL=https://yourdomain.com
PORT=8080
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Run

```bash
go run main.go
```

### 4. Test Webhook (Stripe)

In separate terminal:
```bash
stripe listen --forward-to localhost:8080/webhook/stripe
```

## Bot Commands

- `/start` — show menu
- `/pay` — payment link
- `/id` — get user ID
- Send photo → saves to database

## Environment Variables Reference

| Variable | Description | Example |
|----------|-------------|---------|
| `TELEGRAM_BOT_TOKEN` | Bot token from @BotFather | `123456:ABC...` |
| `STRIPE_PUBLISHABLE_KEY` | Stripe public key | `pk_test_...` |
| `STRIPE_SECRET_KEY` | Stripe secret key | `sk_test_...` |
| `TRON_API_KEY` | TronGrid API key | `api_key...` |
| `WEBHOOK_URL` | Public webhook URL | `https://yourdomain.com` |
| `PORT` | Server port | `8080` |

## Project Overview

**GoBotCat** is a Telegram bot that allows users to purchase photos via two payment methods:
- **Stripe** — traditional card payments (production ready)
- **Tron** — blockchain payments on Shasta Testnet (testing mode)

Users send photos to bot (stored in database), then buyers pay and receive a random photo in return.

### Architecture

```
handlers/      → Telegram & webhook handlers
services/      → Stripe, Tron, Telegram API integration
storage/       → Database layer (GORM, SQLite)
config/        → Configuration management
```

### Databases

- **payments.db** — Stripe payment records
- **tron_payments.db** — Tron transaction records
- **photo.db** — Photo file IDs

## Payment Methods

### Stripe
- Uses test/production API keys
- Webhook verification at `/webhook/stripe`
- Instant payment confirmation

### Tron (Testnet)
- Uses Shasta testnet (free TRX from faucet)
- Polling every 30 seconds for payment confirmation
- Balance-check based verification (testing approach)

## Deployment

Requires Go 1.22+

```bash
# Set environment variables
export TELEGRAM_BOT_TOKEN=...
export STRIPE_SECRET_KEY=...

# Run
go run main.go
```

Works on Railway, Render, or any hosting with Go support.
