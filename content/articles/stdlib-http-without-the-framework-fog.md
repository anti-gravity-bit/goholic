---
title: stdlib HTTP without the framework fog
slug: stdlib-http-without-the-framework-fog
summary: Why I keep a small Go HTTP toolkit in public, and what SDE II interviews actually test when they say “just use Gin.”
date: 2026-09-18
project: bodhiApi
kind: experience
---

# stdlib HTTP without the framework fog

I maintain [bodhiApi](https://github.com/anti-gravity-bit/bodhiApi) because I got tired of not being able to point at the line that turns a `error` into a JSON 500.

SDE II interviews in India still start with a coding round, but the backend round is a walk through **lifecycle**. Where does the request ID come from. What happens if a handler panics after it already wrote a header. How does Ctrl-C drain in-flight work.

Gin and Chi are fine. They are also a lot of source to read at midnight. bodhiApi is deliberately a linear-scan router on `net/http`. With twenty routes, a trie does not save the p99. Readable matching rules do.

The production habit this encodes: **unknown errors become generic 500s**. `err.Error()` is for logs, not for clients. I have seen Postgres connection strings leak in a hurry. The framework should make the safe path the default.

If you open the repo, start at `api.go`, then `internal/core/app`, then middleware. That is the tour I would give a teammate. The rest of the README is for humans who will never work here and still need to trust the code.
