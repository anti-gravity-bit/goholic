---
title: SQL you can file in git
slug: sql-you-can-file-in-git
summary: Sambodhi is an ORM that prints CREATE TABLE instead of mutating production behind your back.
date: 2026-09-20
project: sambodhi
kind: experience
---

# SQL you can file in git

I have used ORMs that “sync” the schema. I have also spent nights restoring a column that sync thought was unused.

sambodhi is the opposite temperament. A struct with tags becomes a `Table`. `CreateTableSQL` returns a string. You look at the string. You put it in a revision file. The engine is just `database/sql`.

The dialect object only knows how to quote names and placeholders. SQLite uses `?`. Postgres will use `$1` when that week of the roadmap lands. Identifier quoting doubles interior quotes so a hostile column name cannot walk out of the ident.

Auto-migrate is a demo feature. Production is a file with a version number and a recorded apply. I will not link a GitHub repo that is not public yet. The code lives in the sambodhi tree on this machine until it is.

The package is named after my son. The technical rule is the same: do not hide what you are about to do.
