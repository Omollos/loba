# Contributing to Loba

Thank you for helping preserve a language. Every word you add matters.

## Local development setup

### Prerequisites
- Go 1.18 or higher (`go version` to check)
- Git
- A terminal

### Steps

1. **Fork and clone the repo**
```bash
git clone https://github.com/YOUR-USERNAME/loba.git
cd loba/api
```

2. **Create your `.env` file**
```bash
cp .env.example .env
```
Then edit `.env` with your own values. You will need:
- `DATABASE_URL` — contact the maintainer for a development database branch
- `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` — create your own OAuth App at github.com/settings/developers (Homepage URL: `http://localhost:8080`, Callback URL: `http://localhost:8080/auth/callback`)
- `SESSION_SECRET` — any long random string you choose
- `FRONTEND_URL` — `http://localhost:5500`

3. **Download dependencies**
```bash
go mod download
```
If this times out on a slow connection, run it again — Go resumes from what was already cached.

4. **Run the API**
```bash
go run cmd/server/main.go
```
You should see:

Connected to Neon database successfully
Loba API starting on port 8080

5. **Serve the frontend**
```bash
cd ../web
python3 -m http.server 5500
```
Then open `http://localhost:5500/dictionary.html` in your browser.

### Important
- Never commit your `.env` file — it is in `.gitignore`
- Run all Go commands from inside the `api/` folder
- Run all Git commands from the repo root (`loba/`)
- Database migrations live in `scripts/` — do not run them against production without discussing with the maintainer first

## Who can contribute?

Anyone. You do not need to be a linguist or a developer. If you speak Dholuo — or any of the languages Loba will support — you are qualified.

## What makes a good entry?

A strong entry has four things:

1. **The word or phrase** in the source language, spelled correctly
2. **An English translation** that captures the real meaning, not just a literal word-for-word version
3. **An example sentence** showing the word used in natural context
4. **A note** on pronunciation, tone, register, or regional variation if relevant

### Good example
| Field | Value |
|-------|-------|
| Dholuo | Chuny |
| English | Heart, soul, spirit — used in both literal and emotional contexts |
| Example (Dholuo) | Chuny maber en geno mar oganda |
| Example (English) | A good heart is the hope of the community |
| Notes | The 'ch' is palatal. Used freely in both everyday and poetic speech. |

### What to avoid
- Single words with no context or example
- Direct dictionary definitions copied from elsewhere
- Entries you are not confident about — flag uncertainty in the notes field instead

## What to contribute

These categories need the most work right now:

- 🟡 Proverbs — most underrepresented, highest cultural value
- 🟡 Emotions — feeling words are hard to find in existing resources  
- 🟡 Body and health — important for future health tool applications
- 🟢 Greetings — good starting point for new contributors

## How entries are reviewed

Every submission enters a review queue. It will not appear in the public corpus until at least two other contributors verify it. If something is flagged as inaccurate, it returns to pending and the submitter is notified.

This is not a judgment — it is how we keep the corpus trustworthy enough to train AI on.

## How to submit

1. Go to the Loba submission form (link coming — platform in development)
2. Fill in all fields you can — partial entries are fine, empty example fields are not
3. Note your dialect region — Kisumu, Siaya, Homa Bay, Migori, and Tanzania Luo all differ in ways worth capturing

## Data licence

Everything you submit is released under **CC BY 4.0**. This means it is free for anyone to use, forever, as long as they credit the Loba corpus. You will be credited as a contributor by name.

## Questions and ideas

Open a thread in [GitHub Discussions](../../discussions). All decisions about the project are made openly there.
