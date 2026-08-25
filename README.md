# loba.
**Languages found, never lost.**

An open corpus for the endangered and underrepresented languages of East Africa — built by the speakers, for the speakers and the tools that will serve them.

## What is Loba?

Loba is a community-built, openly licensed platform that preserves and structures Dholuo — and eventually other East African languages — for the digital age.

Dholuo language data exists in fragments: academic wordlists behind paywalls, Bible translations repurposed for NLP experiments, scattered repositories with inconsistent licensing. None of it is structured consistently enough to build reliable tools on. Loba is the harmonising layer.

Every entry includes:
- The word or phrase in the source language
- English translation
- **Meaning explained in Dholuo** (Tiend wach e Dholuo) — unique to Loba
- Example sentences in both languages
- Cultural context and English equivalent
- Dialect region and part of speech

## Live platform

| | |
|---|---|
| **Dictionary** | https://loba-six.vercel.app/dictionary.html |
| **Contribute** | https://loba-six.vercel.app/index.html |
| **Leaderboard** | https://loba-six.vercel.app/leaderboard.html |
| **About** | https://loba-six.vercel.app/about.html |
| **API** | https://loba-production.up.railway.app |

## Corpus status

| Language | Entries | Status |
|----------|---------|--------|
| Dholuo | 18 | 🟢 Active |
| Kikuyu | — | 🔜 Planned |
| Kamba | — | 🔜 Planned |
| Kalenjin | — | 🔜 Planned |

## API

The REST API is publicly accessible. No API key required for read endpoints.

GET /api/v1/entries — list approved entries (paginated)
GET /api/v1/entries?category=X — filter by category
GET /api/v1/entries?lang=luo — filter by language
GET /api/v1/stats — corpus totals
GET /api/v1/leaderboard — contributors ranked by approved entries
GET /api/v1/languages — all languages with entry counts
GET /api/v1/export/jsonl — full corpus as JSONL (for ML training)
GET /api/v1/export/csv — full corpus as CSV (for research)

Export example:
```bash
curl https://loba-production.up.railway.app/api/v1/export/jsonl > dholuo.jsonl
```

## Tech stack

| Layer | Technology |
|-------|-----------|
| API | Go (standard library) |
| Database | PostgreSQL on Neon.tech |
| Frontend | HTML, CSS, JavaScript |
| Auth | GitHub OAuth + JWT |
| API hosting | Railway |
| Frontend hosting | Vercel |

## Licences

- **Code:** MIT — see [LICENSE](./LICENSE)
- **Data:** [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- **Cite as:** Loba Corpus, github.com/Omollos/loba

## Contributing

Read [CONTRIBUTING.md](./CONTRIBUTING.md) to get started.

No linguistics degree required — if you speak Dholuo, you qualify.

Entries are reviewed within 24 hours before joining the public corpus.

## Community

Questions, ideas, and discussions: [GitHub Discussions](../../discussions)

Collaboration requests from researchers and institutions welcome.

## Built in Kisumu

Started in Kisumu, Kenya in 2025 as an open source project rooted in the East African tech and language community. Part of the broader African NLP ecosystem alongside [Masakhane](https://www.masakhane.io).
