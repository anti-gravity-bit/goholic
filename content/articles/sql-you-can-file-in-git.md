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

The dialect object only knows how to quote names and placeholders. SQLite uses `?`. Postgres uses `$1`. Identifier quoting doubles interior quotes so a hostile column name cannot walk out of the ident.

Auto-migrate is a demo feature. Production is a file with a version number and a recorded apply. [sambodhi](https://github.com/anti-gravity-bit/sambodhi) is public. `cmd/bodhi up` is the writer. The HTTP process still does not call `CreateTableSQL` for you.

The package is named after my son. The technical rule is the same: do not hide what you are about to do.
