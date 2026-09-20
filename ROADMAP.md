# Goholic — SDE II completion roadmap

**Bar:** you can defend cookies, markdown escaping, sqlite durability, and a static/CDN split in an SDE II loop.

## Done

- [x] Package-as-rooms architecture with tests
- [x] sqlite shelf + memory shelf
- [x] HMAC writer cookie
- [x] HTMX landing search
- [x] Static export of `content/articles` to `public/`
- [x] Ten SDE II essays (two per public Go project)

## Remaining

### Week 1 — Secrets and process

- Refuse to boot if password/secret are the compiled defaults
- `log/slog` JSON in production
- Graceful shutdown on SIGINT (match bodhiApi)
- `/healthz` and `/readyz` (sqlite ping)

### Week 2 — Content pipeline

- `goholic-static` fails CI if a slug collides
- RSS feed in `public/feed.xml`
- Canonical URLs from `WebsiteHomeAddress`

### Week 3 — Durability

- sqlite WAL documented
- backup script for `data/goholic.sqlite`
- optional `postgresarticleshelf` behind the same interface

### Week 4 — Interview polish

- One page sequence diagram: login → cookie → dashboard → publish → public GET
- Load test the landing search with `hey`
- Threat note: cookie theft, XSS via markdown, brute force on `/writer/login`

## Definition of done

You can stand up and answer:

1. Why the passport HMAC includes expiry
2. How markdown avoids stored XSS
3. What a static export cannot do that the Go server can
4. Why Vercel is the public door and sqlite stays off Vercel
