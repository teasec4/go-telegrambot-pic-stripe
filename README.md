# GoBotCat

Telegram бот для продажи фото через Stripe платежи.

## Установка

```bash
go mod tidy
```

## Настройка

1. Скопируй `.env.example` в `.env`:
```bash
cp .env.example .env
```

2. Заполни переменные:
```
STRIPE_PUBLISHABLE_KEY=pk_test_...
STRIPE_SECRET_KEY=sk_test_...
TELEGRAM_BOT_TOKEN=123456:ABC...
WEBHOOK_URL=https://yourdomain.com
PORT=8080
```

## Запуск

```bash
go run main.go
```

## Локальное тестирование Stripe webhook

В отдельном терминале:
```bash
stripe listen --forward-to localhost:8080/webhook/stripe
```

## Использование

- `/start` — меню
- `/pay` — ссылка на оплату
- Отправить фото — сохранится в `photos.json`

## Структура

```
config/     — конфигурация
handlers/   — обработчики запросов (bot, webhook)
services/   — бизнес-логика (Stripe, Telegram)
storage/    — хранилище (JSON для фото, SQLite для платежей)
```

## БД

- `photos.json` — список file_id фото по пользователям
- `payments.db` — история платежей

## Развёртывание

Deploy на Railway, Render или любой хостинг:
1. Установи Go 1.25+
2. Установи переменные окружения
3. `go run main.go`
