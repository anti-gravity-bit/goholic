---
title: A panic after the headers
slug: a-panic-after-the-headers
summary: Situation: the handler wrote 200 and then panics. Recovery middleware cannot unwrite bytes. Here is how I talk about it.
date: 2026-09-18
project: bodhiApi
kind: incident
---

# A panic after the headers

Situation: a helper logs and then writes JSON. A later line panics on a nil map. The process has a recover middleware. The client still sees a truncated 200, not a 500.

This is the SDE II version of “we have recover, so we are safe.” You are not.

Once `WriteHeader` has been called, the status is committed. Recover can log the panic and maybe metric it. It cannot honestly change the status. If you wrap `ResponseWriter` to detect a write, you can at least refuse a second write and log `response_already_committed`.

Action I want in a review:

- Recover middleware runs **first** (outermost).
- Handlers return `error`. They do not `panic` for business faults.
- JSON helper records that it wrote.
- Tests: panic before write → 500 JSON; panic after write → log, no double header.

[bodhiApi](https://github.com/anti-gravity-bit/bodhiApi) now refuses a second `JSON` or `Status` with `bodhiApi: response already written`. Recover still cannot unwrite bytes. The honest interview answer is the same: **I know the hole, I would not pretend recover is a time machine.**

That answer is more senior than a clean demo that never panics.
