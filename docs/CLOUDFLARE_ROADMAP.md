# Step-by-step roadmap: put Goholic on Cloudflare

goholic.in should stay a normal Go process. Cloudflare is the front door,
the shield, and the cache. The Go binary still needs a place to sleep
and a disk for sqlite.

Pick **one** live path. Path A is the one that matches a Linode / Kubernetes
resume. Path B is Cloudflare-native if Containers are enabled on the account.

---

## Day 0 — freeze the secrets

1. Copy `.env.example` to a password manager note, not into git.
2. Choose a long writer password and a 32+ character session secret.
3. Decide the sqlite file lives on a volume named `goholic-data`.
4. Never commit `data/goholic.sqlite`.

---

## Day 1 — put the domain on Cloudflare

1. Create a Cloudflare account with the same mailbox as the domain owner.
2. Open **Websites → Add a site** and type `goholic.in`.
3. Choose the Free plan first. You can raise it later.
4. Cloudflare shows two nameservers, for example:
   - `ada.ns.cloudflare.com`
   - `bob.ns.cloudflare.com`
5. At the domain registrar, replace the old nameservers with those two.
6. Wait until the dashboard says the site is **Active**.
   This can take minutes or a few hours.

Do not flip orange-cloud proxy on yet if you still have an old host
pointing at the same name. Finish Day 2 first.

---

## Day 2 — build a container that only runs Goholic

From the project root:

```bash
docker build -t goholic-web-server:latest .
docker run --rm -p 8080:8080 \
  -e GOHOLIC_WRITER_USERNAME=abir \
  -e GOHOLIC_WRITER_PASSWORD='your-real-password' \
  -e GOHOLIC_SESSION_SECRET='your-real-secret-at-least-16' \
  -v goholic-data:/app/data \
  goholic-web-server:latest
```

Open `http://127.0.0.1:8080`. You should see the three seed stories.
Open `/writer/login` and sign in.

If that works, the same image is what Cloudflare will stand in front of.

---

## Path A (recommended): Cloudflare Tunnel in front of your own box

This path uses a small always-on machine you already know: Linode,
a Kubernetes pod, or a cheap VM. Cloudflare never sees the origin IP.

### A1. Run the container on the box

```bash
mkdir -p /var/lib/goholic
docker run -d --name goholic --restart unless-stopped \
  -p 127.0.0.1:8080:8080 \
  -e GOHOLIC_WRITER_USERNAME=abir \
  -e GOHOLIC_WRITER_PASSWORD='your-real-password' \
  -e GOHOLIC_SESSION_SECRET='your-real-secret-at-least-16' \
  -v /var/lib/goholic:/app/data \
  goholic-web-server:latest
```

Bind only to localhost. The public internet should not hit port 8080.

### A2. Install cloudflared on the same box

Follow Cloudflare Zero Trust → Networks → Tunnels → **Create a tunnel**.
Name it `goholic-origin`.

Cloudflare gives one install command. It looks like:

```bash
cloudflared service install <token>
```

### A3. Map the hostname

In the tunnel Public Hostname list:

| Field | Value |
|---|---|
| Subdomain | `@` (or leave blank) |
| Domain | `goholic.in` |
| Path | `*` |
| Service type | `HTTP` |
| URL | `http://127.0.0.1:8080` |

Add a second hostname `www.goholic.in` that points at the same URL.

### A4. DNS records Cloudflare creates for you

You should now see:

- `goholic.in` → CNAME → `<tunnel-id>.cfargotunnel.com` (proxied, orange cloud)
- `www` → CNAME → the same tunnel

### A5. SSL / TLS mode

In Cloudflare → SSL/TLS → Overview, set mode to **Full**.
The browser talks HTTPS to Cloudflare. Cloudflare talks HTTP to
localhost through the tunnel. That is safe because the tunnel is already
encrypted.

### A6. Cache only the public reading pages

Cloudflare → Caching → Cache Rules → Create rule:

- If hostname equals `goholic.in` and path starts with `/article/`
  or path equals `/`
  and request method is GET
  → Eligible for cache, edge TTL 2 hours, browser TTL 10 minutes.

Bypass cache when:

- path starts with `/writer/`
- request method is not GET
- cookie `goholic_writer_session` is present

HTMX search on `/` uses GET with `?q=`. Either bypass `/` when a query
string exists, or keep the TTL short. Short is simpler.

### A7. Lock the writer desk with Cloudflare Access

Zero Trust → Access → Applications → Add application → Self-hosted:

- Application name: `Goholic writer desk`
- Public hostname: `goholic.in`
- Path: `/writer`

Policy: email equals `webdev.abir@gmail.com`.

Now even a guessed password is not enough. Cloudflare asks for your
mailbox first. The Go login is the second lock.

### A8. Turn on the cheap shields

- Security → Settings: Bot Fight Mode on.
- Security → WAF: enable the free Cloudflare managed ruleset.
- Rules → Redirect Rules: `www.goholic.in` → `https://goholic.in` (301).
- Rules → Redirect Rules: `http` → `https` is already default.

---

## Path B: Cloudflare Containers

Use this when the account has **Containers** enabled and you want
no Linode box.

1. Install Wrangler: `npm install -g wrangler`
2. `wrangler login`
3. Create a container application from the dashboard or wrangler
   and point it at this `Dockerfile`.
4. Attach a persistent volume for `/app/data`.
   Without a volume every deploy forgets the blog.
5. Set the same three environment variables as secrets,
   not as plain text:
   - `GOHOLIC_WRITER_USERNAME`
   - `GOHOLIC_WRITER_PASSWORD`
   - `GOHOLIC_SESSION_SECRET`
6. Bind a route: `goholic.in/*` → the container.
7. Repeat Day 1 SSL, cache, Access, and WAF steps from Path A.

Containers are still a newer door than Tunnel. If a deploy eats the
volume or cold-starts slowly, fall back to Path A. The Go code does
not change.

---

## Path C: what not to do

- Do not compile this blog to a Cloudflare Worker by hand.
  Workers are great at the edge. This program wants a process and a file.
- Do not put sqlite on Cloudflare R2 as the live database.
  R2 is an object store, not a filesystem sqlite can lock.
- Do not use Pages Functions as the admin dashboard host.
  Pages is for static files. This dashboard writes.

R2 is still useful as a **nightly backup** of `goholic.sqlite`.
A cron on the origin box can run:

```bash
aws s3 cp /var/lib/goholic/goholic.sqlite s3://goholic-backups/ \
  --endpoint-url https://<accountid>.r2.cloudflarestorage.com
```

---

## Day 3 — prove the site in public

1. `https://goholic.in` shows the three seed stories.
2. `https://goholic.in/article/from-kolkata-with-go` opens.
3. Search box swaps the list without a full paint (HTMX).
4. `https://goholic.in/writer/dashboard` hits Access, then the Go login.
5. Publish a new draft. Wait two minutes. Open a private window.
   The story is on the landing page.
6. `curl -I https://goholic.in/article/from-kolkata-with-go`
   should show `cf-cache-status: HIT` after the second request.

---

## Day 4 — keep it alive

- Snapshot `/var/lib/goholic` every night (or the R2 copy).
- After each `git` change: test, build the image, restart the container,
  watch `docker logs goholic`.
- If you later outgrow sqlite, add a `postgresarticleshelf` package
  that keeps the same `ArticleShelf` promise. The librarian and the
  HTMX pages stay still. That is the point of the cabinet interface.

---

## One-page checklist

- [ ] Domain nameservers point at Cloudflare
- [ ] Image runs locally on :8080
- [ ] Origin only listens on 127.0.0.1:8080
- [ ] Tunnel or Container route for apex + www
- [ ] SSL mode Full
- [ ] Cache public GET pages, bypass `/writer`
- [ ] Access policy on `/writer`
- [ ] Nightly sqlite copy
- [ ] Writer password and session secret live only in the host env
