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

[sambodhi](https://github.com/anti-gravity-bit/sambodhi) is the opposite temperament. A struct with `sambodhi` tags becomes a `Table`. `CreateTableSQL` returns a string. You look at the string. You put it in a revision file. The engine is just `database/sql`.

The dialect object only knows how to quote names and placeholders. SQLite uses `?`. Postgres will use `$1` when that week of the roadmap lands. Identifier quoting doubles interior quotes so a hostile column name cannot escape the ident.

That is enough for an SDE II conversation about **migrations as reviewable artifacts**. Auto-migrate is a demo feature. Production is a file with a version number and a recorded apply.

The package is named after my son. The technical rule is the same: do not hide what you are about to do.
