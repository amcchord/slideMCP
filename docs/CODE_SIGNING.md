# Code Signing

This document has been superseded by [`SIGNING.md`](SIGNING.md), which
covers the full sign + notarize + verify pipeline plus all the
macOS-specific quirks we've debugged into the ground.

If you're an LLM or new developer working on signing, releases, or the
`.mcpb` build pipeline, **read [`SIGNING.md`](SIGNING.md) first**. It
will save you hours.

Quick links:

- One-time setup on a new Mac: [`scripts/setup-signing.sh`](../scripts/setup-signing.sh)
- Read-only sanity check: `make doctor-signing`
- Production build: `source scripts/release-env.sh && make pack-dxt-signed && make verify-dxt`
- Full reference: [`SIGNING.md`](SIGNING.md)
