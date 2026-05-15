# Slide MCP example questions

Each entry pairs a plain-English user question with the canonical first
tool call. Use these as suggestions to the user, or as your own routing
table when the user is vague.

## Health and inventory

- "Are all my Slide boxes healthy?" -> `slide_overview operation=health`
- "What clients/devices/agents do we have?" -> `slide_overview operation=inventory`
- "Show me everything for ACME." -> `slide_overview operation=for_client name_hint=ACME`
- "Tell me about DC-01." -> `slide_overview operation=for_device name_hint=DC-01`

## Backups

- "Did backups run last night for ACME?" -> `slide_backups operation=status_for_client name_hint=ACME hours=24`
- "Did the file server back up?" -> `slide_backups operation=recent_for_agent name_hint=fileserver`
- "Why did the backup fail on Bob's laptop?" -> `slide_backups operation=recent_for_agent name_hint=bob` (then inspect last error_message)
- "Start a backup of the SQL box now." -> `slide_backups operation=start name_hint=sql`
- "Pause backups on this machine for 4 hours." -> `slide_agents operation=pause_backups name_hint=...` (with paused_until=RFC3339)

## Files and restores

- "Find Q4-budget.xlsx on Bob's laptop." -> `slide_files operation=search name_hint=bob search_term=Q4-budget`
- "Show me every snapshot that has C:\\Users\\bob\\Documents\\Q4-budget.xlsx." -> `slide_files operation=versions name_hint=bob path=C:\\Users\\bob\\Documents\\Q4-budget.xlsx`
- "Restore Tuesday's version." -> `slide_files operation=create_restore snapshot_id=... device_id=...`
- "Push it back to C:\\SlideRestore on the laptop." -> `slide_files operation=create_push ...`

## Recovery (BCDR / DR)

- "Boot a recovery VM from yesterday's snapshot of DC-01." -> first `slide_snapshots operation=recent_for_agent name_hint=DC-01`, then `slide_recovery operation=boot_vm snapshot_id=... device_id=...`
- "Give me an RDP file for the booted VM." -> `slide_recovery operation=get_rdp_bookmark virt_id=...`
- "Export this snapshot as a VHDX." -> `slide_recovery operation=export_image snapshot_id=... device_id=... image_type=vhdx`
- "Set up a DR network so the VM can reach the internet." -> `slide_recovery operation=create_network type=standard internet=true ...`

## Alerts and triage

- "What unresolved alerts do I have? Sort worst first." -> `slide_alerts operation=triage`
- "Is my Slide storage almost full anywhere?" -> `slide_alerts operation=triage` (look for `device_storage_space_*`)
- "Mark this alert resolved." -> `slide_alerts operation=update alert_id=... resolved=true`

## Audit and compliance

- "What changed in the last 24 hours?" -> `slide_audit operation=recent hours=24`
- "Who deleted that snapshot?" -> `slide_audit operation=list audit_action_name=snapshot_deleted`
- "Show me every action on agent X." -> `slide_audit operation=list audit_resource_type_name=agent` (then filter by resource_id)

## Admin

- "List my Slide users." -> `slide_admin operation=list_users`
- "Change the account-level alert email addresses." -> `slide_admin operation=update_account alert_emails=[...]`

## When in doubt

- "I don't know where to start." -> `slide_help operation=getting_started`
- "What can this thing do?" -> `slide_help operation=what_can_you_do`
- "Help me restore a file." -> use the `/slide.restore-file` prompt.
