# Plan: Copy Public Key to Clipboard

**Status:** Proposed
**Priority:** Short Term

## Goal

One-click button to copy a public key to clipboard for pasting into GitHub, servers, etc.

## Approach

1. Add a "Copy" button next to each key in the table (or in the key detail view)
2. Use the Clipboard API: `navigator.clipboard.writeText(pubkey)`
3. Show brief "Copied!" feedback on the button

## Changes

- `keys.templ` — add Copy button with `onclick` handler
- `static/js/app.js` — `copyToClipboard(text)` helper
- API: `GET /api/keys/{name}` already returns `publicKey` field
- For table view: embed public key as `data-pubkey` attribute or fetch on click

## Notes

- Clipboard API requires secure context (HTTPS or localhost) — both our modes qualify
- Wails webview supports Clipboard API
