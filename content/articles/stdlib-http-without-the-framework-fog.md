---
title: stdlib HTTP without the framework fog
slug: stdlib-http-without-the-framework-fog
summary: I got tired of not being able to point at the line that turns an error into a JSON 500.
date: 2026-09-18
project: bodhiApi
kind: experience
---

# stdlib HTTP without the framework fog

I keep [bodhiApi](https://github.com/anti-gravity-bit/bodhiApi) public because I got tired of not being able to point at the line that turns an `error` into a JSON 500.

Gin and Chi are fine. They are also a lot of source to read at midnight. bodhiApi is a linear-scan router on `net/http`. With twenty routes a trie does not save the p99. Readable matching rules do.

The habit I care about: **unknown errors become generic 500s**. `err.Error()` is for logs, not for clients. I have seen a Postgres connection string leak in a hurry. The framework should make the safe path the default.

If you open the repo, start at `api.go`, then `internal/core/app`, then middleware. That is the tour I would give a tired teammate. The rest of the README is for people who will never work here and still need to trust the code.
