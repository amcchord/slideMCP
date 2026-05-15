# Troubleshooting the Slide MCP extension

## "Your Slide API token is invalid or expired"

The token currently configured in Claude Desktop's Extensions settings
isn't accepted by api.slide.tech. To fix:

1. Open <https://console.slide.tech> and sign in.
2. Navigate to **My Settings -> API Tokens**.
3. Generate a fresh token (or revoke + reissue the old one).
4. In Claude Desktop, open **Settings -> Extensions -> Slide Backup**
   and paste the new token.
5. Restart the extension (toggle it off and on).

The token gives full access to your Slide account; it is stored by
Claude Desktop and only forwarded to the slide-mcp-server process at
launch. It is never written into any project file.

## "Could not reach api.slide.tech"

slide-mcp-server reached out to <https://api.slide.tech> and got a
network-level error before any HTTPS handshake completed. Common
causes:

- No internet on the machine running Claude Desktop.
- A corporate firewall or zero-trust agent is blocking outbound HTTPS
  to api.slide.tech. Ask your network admin to allow the host.
- A captive portal is intercepting requests (hotel / airport wifi).
- DNS isn't resolving api.slide.tech.

You can verify connectivity from the same machine with:

```
curl -sSI https://api.slide.tech/v1/account -H "Authorization: Bearer YOUR_TOKEN"
```

A `200`, `401`, or `403` confirms reachability. A timeout or
"connection refused" confirms a network-level block.

## "Operation 'X' not available in 'read-only' mode"

The extension is currently in `read-only` mode (or `safe` mode, which
blocks deletes / device poweroff / reboot). Two ways to fix:

- Recommended: keep the safer mode and reword your request so it
  doesn't need the blocked operation.
- Otherwise: open **Settings -> Extensions -> Slide Backup** and set
  **Tool permissions** to `safe` or `full`.

Permission tiers:

- `read-only` - only list/get/search/browse/triage. No mutation at all.
- `safe` (default) - everything in read-only PLUS restores, backup
  launch, agent/device/network settings, alert resolution. Deletes and
  device poweroff/reboot are still blocked.
- `full` - everything, including deletes and remote power-cycle.

## "No match for name_hint='X'"

You passed a `name_hint` that didn't fuzzy-match any agent / device /
client visible to the current API token. Options:

- Call `slide_overview operation=inventory` to see the full list of
  clients, devices, and agents the token can see.
- Try a different substring of the hostname or display name.
- Confirm the entity is still in this Slide account (it may have been
  decommissioned).

## "Ambiguous name_hint - multiple matches"

The hint matched more than one entity. The response includes a
`candidates` array with each match's id, name, and kind. Pick one and
re-call the tool with the explicit `*_id` instead of `name_hint`.

## "Slide API server-side error"

A `5xx` response from api.slide.tech. The MCP server retries once
transient errors automatically. If you keep hitting these, check
<https://status.slide.tech> for an active incident.

## "I don't know where to start"

Call `slide_help operation=getting_started`. Or open the slash-command
menu in Claude Desktop and run `/slide.welcome`.

## Verifying your install end-to-end

Outside Claude Desktop, run:

```
slide-mcp-server --doctor
```

This checks: token validity, network reachability, sample reads
against each major endpoint family, and the version of the binary.
Exits 0 if everything passes, non-zero on any failure.
