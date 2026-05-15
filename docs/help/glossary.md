# Slide glossary

The terms below are always interpreted in the Slide sense, not the
generic IT sense. When a user says "device" or "agent" without context,
assume the Slide meaning if any other Slide vocabulary is in the
conversation.

## Core entities

- **client** (`c_xxxxxxxxxxxx`) - an MSP customer or organisational
  unit. Devices, agents, and alerts are grouped by client.

- **device** (`d_xxxxxxxxxxxx`) - a physical Slide backup appliance
  (the "Slide box"). Lives in a client's office or rack. Stores
  snapshots locally and replicates them to Slide's cloud.

- **agent** (`a_xxxxxxxxxxxx`) - a backup agent installed on a
  protected server, workstation, or laptop. Reports to a single device
  and takes backups of the protected system on a schedule.

- **protected system** - the server or endpoint that a Slide agent
  protects. Not a Slide entity directly; agents are the addressable
  proxy.

- **snapshot** (`s_xxxxxxxxxxxx`) - a single point-in-time backup of
  an agent. Restore points are snapshots. Verified snapshots have
  `verify_boot_status` / `verify_service_status` set.

- **restore session** (`rest_xxxxxxxxxxxx`) - an active file-restore
  session that mounts a snapshot's filesystem so it can be browsed or
  pushed back to a protected system. Time-limited; gets cleaned up
  automatically.

- **VM** / **virtual machine** (`v_xxxxxxxxxxxx`) - a recovery VM
  booted on a Slide device from a snapshot. Has VNC + optional RDP
  access. Used for actual disaster-recovery failover.

- **image export** (`img_xxxxxxxxxxxx`) - an on-demand export of a
  snapshot as a VHD/VHDX/VMDK/QCOW2/RAW disk image for use outside
  the Slide appliance (e.g. uploading to Azure / VMware).

- **alert** (`al_xxxxxxxxxxxx`) - an unresolved condition raised by a
  device or agent. Common types: `device_storage_space_critical`,
  `agent_backup_failed`, `agent_not_checking_in`.

- **audit entry** (`aud_xxxxxxxxxxxx`) - a single row in the account
  audit log. Records who did what to which resource and when.

## Backup vocabulary

- **backup run** - one execution of an agent's scheduled or on-demand
  backup. Lives in `slide_backups`; produces a snapshot on success.
- **backup schedule** - per-agent interval + start/end hours + day mask
  controlling when backups are allowed to run.
- **retention policy** - per-agent rule for how long snapshots are kept
  locally. Presets: `lean`, `balanced`, `comprehensive`.
- **file index** - per-agent flag enabling cross-snapshot file search.
  Required for `slide_files operation=search` to work.
- **pause / resume** - per-agent control to halt backups until a
  timestamp or indefinitely.

## Recovery vocabulary

- **boot VM** - create a recovery VM from a snapshot. Lives in
  `slide_recovery operation=boot_vm`.
- **DR network** - a virtual network on the Slide device that booted
  VMs can attach to. Supports IPSec, port-forwards, and WireGuard
  peers for external access.
- **RDP bookmark** - a downloadable `.rdp` file generated for a
  running VM so the operator can connect with Windows Remote Desktop.
- **service verification** - the per-snapshot health check that boots
  the snapshot in a sandbox and confirms Windows services come up
  cleanly. Status appears as `verify_service_status` on snapshots.

## MSP / BCDR vocabulary

- **BCDR** - business continuity and disaster recovery. Slide is a
  BCDR product.
- **RTO** - recovery time objective. How fast you can restore service.
- **RPO** - recovery point objective. How much data you might lose.
  Defined by the agent's backup interval.
- **failover** - switching production workload over to a recovery VM
  during an outage.
- **DR runbook** - the documented procedure for failing over a
  protected system. `/slide.dr-runbook` generates one from real data.
