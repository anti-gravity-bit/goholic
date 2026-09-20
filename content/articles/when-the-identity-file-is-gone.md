---
title: When the identity file is gone
slug: when-the-identity-file-is-gone
summary: Situation: laptop wiped, ciphertext still in the bucket. envlock cannot help you if you never copied the age identity.
date: 2026-09-19
project: envlock
kind: incident
---

# When the identity file is gone

Situation: disk dies. The bucket still has `prod.env.age`. `envlock pull` fails because `~/.envlock/identity.txt` is not there. There is no “email us for recovery.”

This is not a bug. It is the product.

Action:

1. Identity is a **backup**, same as a database dump. Print it, or keep it in a second offline place, or it does not exist.
2. Ciphertext in object storage is not a backup of the key.
3. `doctor` should fail before `pull` with a sentence a human can read.
4. Break-glass is `age -d -i /path/to/identity` on the `.age` object you downloaded with `s3cmd` or the console.

If a teammate asks for a hosted reset flow, the answer is: that would make envlock a vault. Vault is a different trust story, a different on-call, and a different invoice.

SDE II is knowing which problems you are **refusing** to solve.
