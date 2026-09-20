---
title: Packages should tell the story
slug: packages-should-tell-the-story
summary: If a stranger can read the folder names and guess the architecture, the design is doing its job.
date: 2026-09-17
project: goholic
kind: experience
---

# Packages should tell the story

A blog is a tiny company.

- The **article** room decides what a story is.
- The **shelf** room only promises to keep stories.
- The **librarian** room follows the house rules.
- The **webpage** room talks to browsers.
- The **passport** is the key to the writer desk, not a second database.

If those names are clear, you already understand [goholic](https://github.com/anti-gravity-bit/goholic).

I ship production Go services for a living. The ones that last are the ones a tired teammate can open at midnight and still explain out loud. This site is that idea with the volume turned down.

The public site on Vercel is a **static export** of the same markdown. sqlite and the writer cookie stay off the CDN. That split is an SDE II sentence: **the origin that mutates is not the origin that serves strangers.**
