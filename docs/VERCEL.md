# Point goholic.in at Vercel

The domain already uses Cloudflare nameservers (`aida.ns.cloudflare.com`, `kip.ns.cloudflare.com`). There is no A/CNAME yet.

Static files are in `public/`. The live site is the book *Defer Nothing*. Do not overwrite `public/` with `go run ./cmd/goholic-static`. sqlite and the writer desk do **not** run on Vercel.

## One-time

1. Push this repo to GitHub (`anti-gravity-bit/goholic`).
2. [vercel.com/new](https://vercel.com/new) → import `goholic`.
3. **Root Directory:** `public`
4. Framework preset: Other. No build command.
5. Project → Settings → Domains → add `goholic.in` and `www.goholic.in`.
6. In Cloudflare DNS for `goholic.in`:
   - `CNAME @` → `cname.vercel-dns.com` (or the exact target Vercel shows)
   - `CNAME www` → `cname.vercel-dns.com`
   - Proxy can stay orange; if TLS loops, switch to DNS-only (grey).
7. Cloudflare SSL/TLS mode: **Full (strict)** once Vercel cert is issued.

## Update posts later

```bash
# edit content/articles/*.md
go run ./cmd/goholic-static
git add content public && git commit -m "essay" && git push
```

Vercel redeploys from GitHub.
