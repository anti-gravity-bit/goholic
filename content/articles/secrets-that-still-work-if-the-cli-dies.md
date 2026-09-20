---
title: Secrets that still work if the CLI dies
slug: secrets-that-still-work-if-the-cli-dies
summary: envlock encrypts .env with age and puts ciphertext in a bucket I already pay for. The break-glass path is stock age -d.
date: 2026-09-19
project: envlock
kind: experience
---

# Secrets that still work if the CLI dies

I do not want a SaaS to hold production env files. I also do not want a clever tool that only I can decrypt.

[envlock](https://github.com/anti-gravity-bit/envlock) encrypts with [age](https://age-encryption.org). The object in Spaces, Linode, S3, or R2 is ciphertext. Identity lives in `~/.envlock/identity.txt` mode 0600. If I delete the CLI tomorrow, `age -d` still opens the file.

That is the SDE II design constraint: **tools must fail open for the owner and fail closed for everyone else.**

`run --no-disk` is the other half. Decrypt into the child environment. Do not leave a plaintext `.env` on a build agent.

The stores are boring on purpose. One S3-compatible driver. Endpoints documented in `docs/stores.md`. Live tests gated on env so CI never needs my keys.

If you are a hiring manager and you only click one file, click the doctor command. A secrets tool that cannot tell you what is missing is theatre.
