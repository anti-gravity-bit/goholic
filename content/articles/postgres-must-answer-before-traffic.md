---
title: Postgres must answer before traffic
slug: postgres-must-answer-before-traffic
summary: Halide treats readiness as a ping, not a process that happens to be up.
date: 2026-09-16
project: halide
kind: experience
---

# Postgres must answer before traffic

[halide](https://github.com/anti-gravity-bit/halide) is a chi + pgx service for a disposable-camera product. The interesting part is not the camera metaphor. It is the startup clipboard.

Live (`/health`) means the process can talk HTTP. Ready means the pool pinged Postgres inside the connect timeout. Kubernetes should not send users to a replica that only completed `ListenAndServe`.

I have watched a deploy “succeed” with the database still running migrations. The process was alive. Every request 500’d. A readiness probe tied to ping would have kept the old replica in rotation.

The JSON envelope in `pkg/apiresponse` is the other small discipline: every handler returns the same shape. Clients and tests do not special-case a snowflake error.

Firebase sits at the edge. Credentials stay in a gitignored file. If you clone the repo and find a service account, I failed the SDE II hygiene check.
