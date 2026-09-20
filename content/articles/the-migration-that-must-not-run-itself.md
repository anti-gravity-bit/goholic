---
title: The migration that must not run itself
slug: the-migration-that-must-not-run-itself
summary: Situation: two app replicas boot and both try to CREATE TABLE. Visible SQL is not enough if apply is not exclusive.
date: 2026-09-20
project: sambodhi
kind: incident
---

# The migration that must not run itself

Situation: you put `CreateTableSQL` in `main` “just for local.” Two Kubernetes replicas start. Both run the statement. Today it is `IF NOT EXISTS` and you get away with it. Next week it is `ALTER TABLE ... DROP COLUMN` and you do not.

SDE II fix:

- Migrations are a command, not a side effect of serving HTTP.
- A `sambodhi_revisions` table records which files ran.
- Apply takes a lock (`pg_advisory_lock` on Postgres, a busy-wait table on sqlite).
- The HTTP process **refuses** to auto-migrate. It can ping and refuse to become ready if a required revision is missing.

Dual-write (app + migrate in the same boot) is how you get split-brain schema. The outbox pattern is the cousin of this idea on the event path. Same instinct: **one writer for the dangerous step.**

sambodhi’s roadmap says `cmd/bodhi up` is the writer. Until that CLI is thick, I still will not call `Exec(createSQL)` from `ListenAndServe`.
