---
title: HMAC cookies are not a session store
slug: hmac-cookies-are-not-a-session-store
summary: Situation: someone copies the writer cookie. Expiry and HMAC are the whole defense. There is no server-side session table to revoke.
date: 2026-09-17
project: goholic
kind: incident
---

# HMAC cookies are not a session store

Situation: the writer cookie leaks in a screenshot. There is no “log out all sessions” table. The cookie is `unix.hmac`. If the HMAC matches and the unix timestamp is still in the window, the desk opens.

That is a trade: one sqlite file, no Redis session. The cost is **revocation**.

SDE II response:

- Short expiry.
- Secret from the environment, not a compiled default in production.
- Cookie flags: `HttpOnly`, `Secure`, `SameSite=Strict` on the real domain.
- Changing the secret invalidates every cookie. That is the kill switch.
- Markdown is escaped so a draft cannot steal the cookie with a script tag.

If I needed revoke-one-user, I would add a `token_id` in the cookie and a tiny sqlite table of revoked ids. I have not, yet. The honest README says so.

People say “JWT is stateless” and then cannot name the logout hole. This post is that name.
