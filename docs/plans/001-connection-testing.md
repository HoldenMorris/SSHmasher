# Plan: SSH Connection Testing

**Status:** Proposed
**Priority:** Medium

## Goal

Allow users to test SSH connections from the app — verify that a host entry in `~/.ssh/config` is reachable and authentication works.

## Approach

1. Add `TestConnection(host string, timeout time.Duration)` to `internal/ssh/`
2. Use `golang.org/x/crypto/ssh` to dial TCP and perform SSH handshake
3. Return structured result: reachable, auth method used, server banner, latency, errors
4. New API endpoint: `POST /api/config/hosts/{alias}/test`
5. UI: "Test" button on each host row, inline result display

## Considerations

- Timeout should be configurable (default 10s)
- Need to handle: host unreachable, auth failure, host key mismatch
- Should NOT store passwords — only test key-based auth or agent
- May need to read IdentityFile from config entry
