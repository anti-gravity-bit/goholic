---
title: Shutdown is a feature
slug: shutdown-is-a-feature
summary: Situation: SIGTERM during an upload. If you os.Exit, the client retries and you double-write. Drain, then close the pool.
date: 2026-09-16
project: disposable-camera-api
kind: incident
---

# Shutdown is a feature

Situation: the orchestrator sends SIGTERM in the middle of a POST. The handler is in a transaction. If the process dies immediately, the client times out and retries. You now have two events or a half-written row.

SDE II shutdown:

1. Stop accepting new connections.
2. Wait up to `SHUTDOWN_TIMEOUT` for in-flight requests.
3. Close the pgx pool.
4. Exit 0 if the drain finished, non-zero if you had to cut.

This is the same story as bodhiApi’s graceful shutdown example, applied to a service with a database. Timeouts are part of the API contract. Ten seconds is a number you pick because the load balancer’s idle timeout is larger, not because it felt nice.

If a request cannot finish in the drain window, make the handler **idempotent** (event id, unique constraint). Shutdown without idempotency is how you get duplicate photos in a “disposable” camera, which is a bad joke.

I want the interview to stay on that sequence, not on Firebase product trivia.
