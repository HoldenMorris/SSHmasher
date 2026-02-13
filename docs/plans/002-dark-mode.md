# Plan: Dark Mode Toggle

**Status:** Proposed
**Priority:** Short Term

## Goal

Add a light/dark mode toggle that persists across sessions.

## Approach

Pico CSS already supports dark mode via `data-theme="dark"` on the `<html>` tag.

1. Add a toggle button/icon in the nav bar
2. JavaScript: toggle `data-theme` attribute on `<html>`
3. Persist preference in `localStorage`
4. On page load, read `localStorage` and apply before render (avoid flash)

## Changes

- `layout.templ` — add toggle button in nav
- `static/js/app.js` — theme toggle logic + localStorage
- No backend changes needed
