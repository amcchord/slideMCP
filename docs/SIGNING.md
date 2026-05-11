# macOS Signing + Notarization Reference

This is the canonical guide for the slide-mcp-server release-signing pipeline.
Every quirk here was paid for in real debugging time; **do not reinvent
the wheel by trying alternatives until you've read this document end to
end**.

If you only have 60 seconds: run `make doctor-signing`, fix what it
reports, then `source scripts/release-env.sh && make pack-dxt-signed &&
make verify-dxt`. The four lines you want from `verify-dxt` are:

```
codesign:     signed and verified
chain:        Developer ID Application: ...
notarization: notarized by Apple (verified online)
stapled:      no (Apple can't staple raw Mach-O; online verification works)
```

(`stapled: no` is expected and fine; see "Stapling" below.)

---

## 1. The trust model and what we ship

macOS Gatekeeper accepts a binary on a fresh user's Mac if and only if
**all** of these are true:

1. The binary is signed with a valid `Developer ID Application` cert
   that chains to **Apple Root CA**.
2. The signature uses the **hardened runtime** (`--options runtime`).
3. The signature has an **Apple Trusted Timestamp**
   (`--timestamp`).
4. Apple's notary service has the binary's CDHash on file with status
   `Accepted` (it gets there via `xcrun notarytool submit`).
5. Either: (a) a notarization ticket is **stapled** to the binary
   (offline), OR (b) the user's Mac can reach Apple's CDN to verify
   notarization on first launch.

Our pipeline produces `build/slide-mcp-server.mcpb`, a Claude Desktop
Extension bundle whose `server/slide-mcp-server-darwin-universal` is
the file Gatekeeper actually evaluates when Claude Desktop spawns it.
The bundle layer is just a zip; only the inner binary needs all five
properties above.

```mermaid
flowchart LR
    A[go build per arch] --> B[lipo darwin-amd64+arm64]
    B --> C[codesign --options runtime --timestamp]
    C --> D[notarytool submit --wait]
    D --> E[stapler staple]
    E --> F[stage server/ tree]
    F --> G[zip into .mcpb]
    G --> H[upload to GitHub release]
    H --> I[user drag-drops]
    I --> J[Claude Desktop spawns inner binary]
    J --> K[Gatekeeper checks signature + notarization]
```

The whole pipeline lives in [`Makefile`](../Makefile) under the
`pack-dxt-signed` target. Do not try to reorder these steps; the order
is load-bearing. Notably, signing must happen **before** the binary is
zipped for notary submission, and the stapled binary must be the one
copied into the staging dir for the `.mcpb`, not the unstapled one
sitting elsewhere in `build/`.

---

## 2. First-time setup (new dev machine)

Run **one** command from the repo root:

```bash
./scripts/setup-signing.sh
```

The script is idempotent. It will:

1. Verify Xcode CLI tools / openssl / `security` are present.
2. If no `Developer ID Application` identity is in the login keychain,
   generate a fresh CSR, open Apple's portal, wait for you to download
   the `.cer`, and import it (with all the format conversions described
   in "Section 3 - macOS quirks").
3. Auto-install the **Developer ID G2 intermediate** certificate from
   `https://www.apple.com/certificateauthority/DeveloperIDG2CA.cer`.
4. Read `APPLE_ID` + `APP_SPECIFIC_PASSWORD` from
   `scripts/release-env.sh` if present, else prompt.
5. Auto-derive the Team ID from the imported cert.
6. Run `xcrun notarytool store-credentials slide-mcp-release ...` to
   stash credentials in the keychain.
7. Atomically rewrite `scripts/release-env.sh` (gitignored, mode 600)
   with `DEVELOPER_ID`, `KEYCHAIN_PROFILE`, `APPLE_TEAM_ID`,
   `APPLE_ID`, `APP_SPECIFIC_PASSWORD`.
8. Smoke-sign + verify a throwaway binary as proof.
9. Offer to run `make pack-dxt-signed` immediately.

If the script ever errors mid-flow, it prints Apple's actual error
message indented under the failing step. Do not silence those errors;
they are the only way to know whether the failure is (e.g.) a stale
app-specific password vs a wrong Team ID.

### Things you need from Apple before running the script

- An **Apple Developer Account** ($99/yr) belonging to your team.
- An **App-specific password** generated at
  <https://account.apple.com/account/manage> -> Sign-In and Security
  -> App-Specific Passwords. **These expire** when you change your Apple
  ID password or after long disuse; symptom is `HTTP 401 Invalid
  credentials` from notarytool. Generate a new one and update
  `scripts/release-env.sh` if that happens.

### Read-only sanity check on any machine

```bash
make doctor-signing
```

Calls `scripts/setup-signing.sh --check`. Exits non-zero if anything
is missing. Use this in CI before invoking `pack-dxt-signed`.

---

## 3. macOS quirks (the lessons paid for in time)

Every item below is something we tried "the obvious way" and got bitten
by. Future LLMs / humans: **do not change the script to use the obvious
way without reading why we don't**.

### 3.1 Apple ships `.cer` files in DER, not PEM

`openssl x509 -in cert.cer -outform PEM` succeeds; `openssl pkcs12
-export -in cert.cer ...` does not, because pkcs12 expects PEM input.

**Fix in script**: convert via `openssl x509 -inform DER -in $cer
-outform PEM -out $pem` before any further use. Fall back to
`-inform PEM` in case Apple ever changes formats.

### 3.2 OpenSSL 3.x writes private keys in PKCS#8 by default

`openssl genrsa -out key.pem 2048` produces a file starting with
`-----BEGIN PRIVATE KEY-----` (PKCS#8). macOS's `security import`
*only* accepts traditional PKCS#1 (`-----BEGIN RSA PRIVATE KEY-----`)
when importing a private key with `-f openssl`.

**Fix in script**: `openssl rsa -in key.pem -traditional -out
key-pkcs1.pem`, then import the PKCS#1 version.

### 3.3 macOS Sequoia/Tahoe `security import` rejects ALL modern .p12

We tried, in order: Homebrew OpenSSL 3.6 default; OpenSSL 3.6 with
`-legacy`; OpenSSL 3.6 with `-legacy -keypbe PBE-SHA1-3DES -certpbe
PBE-SHA1-3DES -macalg SHA1`; macOS LibreSSL 3.3.6 default. Every
single combination produced a .p12 that `security import` rejected
with `SecKeychainItemImport: Unknown format in import.`

**Fix in script**: skip the .p12 path entirely. Import the private key
and cert as **separate files** (`security import key.pem -t priv -f
openssl ...`, then `security import cert.pem -t cert ...`). macOS
auto-pairs them in the keychain by matching public keys, producing a
usable Identity.

If you read older Apple docs that say "use a .p12", they're correct
for macOS Monterey and earlier. Sequoia/Tahoe are stricter.

### 3.4 The `Developer ID Certification Authority` (G2) intermediate is NOT in the system bundle

macOS ships Apple Root CA in `/System/Library/Keychains/SystemRootCertificates.keychain`,
but the **G2 intermediate** that signs Developer ID Application certs
is NOT bundled. Without it in your keychain, `codesign` fails with
`unable to build chain to self-signed root for signer "..."`.

**Fix in script**: download the G2 intermediate from
`https://www.apple.com/certificateauthority/DeveloperIDG2CA.cer` and
add it to the login keychain via `security add-certificates`.

### 3.5 NEVER use `add-trusted-cert -r trustAsRoot` on the G2 intermediate

This was the trap that cost us the most time. `security add-trusted-cert
-r trustAsRoot DeveloperIDG2CA.cer` *appears* to install the
intermediate, but it actually writes a `kSecTrustSettings` entry that
tells macOS "treat the G2 cert AS a root CA". Subsequent chain
validation breaks because the validator stops walking at G2 and tries
to verify it's self-signed (it isn't; it's signed by Apple Root CA).
Result: `unable to build chain to self-signed root` even though the
intermediate IS present.

**Fix in script**: use `security add-certificates -k login.keychain
DeveloperIDG2CA.cer` (NO trust override). This adds the cert as a
plain intermediate, lets macOS walk the chain naturally to Apple
Root CA in the system bundle, and validation succeeds.

If you ever see this error after a fresh install, check
`security trust-settings-export` for stale `trustAsRoot` entries on
the G2 intermediate; if present, remove via `security
remove-trusted-cert <cert.pem>`.

### 3.6 `xcrun stapler` returns Error 73 on raw Mach-O binaries

Apple's `stapler staple` only works on `.app` bundles, `.pkg` files,
`.dmg` images, or compressed archives that contain those. Raw Mach-O
binaries (which is what we ship inside the `.mcpb`) cannot have a
ticket physically attached to them. This is documented Apple behavior,
not a bug in our pipeline.

**The binary is still notarized.** Apple's notary service has the
CDHash on file with `status: Accepted`. macOS Gatekeeper queries
Apple's CDN over HTTPS the first time the binary launches on a new
user's machine, gets the notarization confirmation, and caches the
result locally. Subsequent launches work offline.

**The Makefile target ignores Error 73 by design** (the leading `-` on
the `xcrun stapler staple` line). Do not change this; it would fail
every release build.

The only edge case where this matters is a fully air-gapped Mac on
first launch. Those Macs are not our target audience.

### 3.7 `spctl --type execute` always rejects raw Mach-O

`spctl --assess --type execute --verbose` is meant for `.app` bundles.
On a raw Mach-O it always prints `rejected (the code is valid but does
not seem to be an app)` regardless of whether the binary is signed and
notarized. This is **not** a Gatekeeper rejection of our binary; it's
spctl saying "I don't have a policy for this artifact type".

**Use this instead** to verify notarization status:

```bash
codesign --test-requirement="=notarized" --verify --verbose=2 \
    build/slide-mcp-server-darwin-universal
```

`explicit requirement satisfied` = notarized (macOS queries Apple
online to confirm). This is what `make verify-dxt` runs on macOS.

### 3.8 App-specific passwords expire silently

Apple invalidates app-specific passwords when:

- You generate a new one (max ~25 active at a time before old ones drop)
- Your main Apple ID password changes
- You revoke it manually at appleid.apple.com
- It hasn't been used in a long time

Symptom: `xcrun notarytool ...` returns
`Error: HTTP status code: 401. Invalid credentials.` even though the
password format looks correct.

**Fix**: generate a fresh password at
<https://account.apple.com/account/manage> -> Sign-In and Security ->
App-Specific Passwords. Update `scripts/release-env.sh`'s
`APP_SPECIFIC_PASSWORD` value. Re-run `./scripts/setup-signing.sh`.

### 3.9 Make subshells need exported env vars

When `setup-signing.sh` calls `(cd $REPO_ROOT && make pack-dxt-signed)`,
the make subshell sees only **exported** env vars from the parent. The
script's local `DEVELOPER_ID="$(find_existing_identity)"` assignment
doesn't export, so make sees `DEVELOPER_ID=""` and fails.

**Fix in script's `offer_test_build`**: source `release-env.sh` *inside*
the make subshell with `set -a; . scripts/release-env.sh; set +a` so
its `export` statements take effect for make. Same pattern applies if
you ever invoke make from another script - either source the env file
in the subshell or pass `DEVELOPER_ID=... KEYCHAIN_PROFILE=...` on the
command line.

### 3.10 Quote `$(DEVELOPER_ID)` in Makefile codesign invocations

Real Developer IDs contain spaces and parentheses
(`Developer ID Application: Austin McChord (7PTN7E8EDS)`). An unquoted
`codesign --sign $(DEVELOPER_ID)` causes the shell to split this into
multiple arguments, and the parentheses become a parse error
(`/bin/sh: syntax error near unexpected token '('`). Always quote:

```make
codesign --sign "$(DEVELOPER_ID)" \
    --options runtime --timestamp --force \
    "$(BINARY)"
```

Same for `xcrun notarytool ... --keychain-profile "$(KEYCHAIN_PROFILE)"`.

---

## 4. Where things live

- **Identity (cert + private key)**: macOS login keychain
  (`~/Library/Keychains/login.keychain-db`). Auto-paired into an
  Identity by macOS when both are present with matching public keys.
- **G2 intermediate cert**: same login keychain, added by the script
  via `add-certificates`.
- **Notarytool credentials** (Apple ID, app-specific password,
  Team ID): macOS keychain under generic-password service
  `com.apple.gke.notary.tool`, account name `slide-mcp-release`. Set
  by `xcrun notarytool store-credentials`. Never written to disk in
  plaintext.
- **The private key file** the CSR was generated from:
  `~/.config/slide-mcp-signing/signing.key` (PKCS#8 PEM, mode 600,
  directory mode 700). Kept around so a re-issued cert can be paired
  back without generating a new key. Not strictly required after
  import but useful for cert rotation.
- **Convenience env file**: [`scripts/release-env.sh`](../scripts/release-env.sh)
  (gitignored, mode 600). Exports `APPLE_ID`, `APP_SPECIFIC_PASSWORD`,
  `APPLE_TEAM_ID`, `DEVELOPER_ID`, `KEYCHAIN_PROFILE`. Must be
  `source`d before running make targets that need credentials.
- **Setup script**: [`scripts/setup-signing.sh`](../scripts/setup-signing.sh)
  (gitignored via the existing `**/setup-signing.sh` rule). Personal
  to each dev machine; may contain machine-specific tweaks.

What is **NOT** anywhere in the repo: the cert itself (only in your
keychain), the private key (only in your keychain + the local
`~/.config/slide-mcp-signing/` dir), the app-specific password (only
in `release-env.sh` which is gitignored, plus the macOS keychain).
The repo carries zero secrets.

---

## 5. Build / verify commands

```bash
# Required once per shell session before any signed-build target
source scripts/release-env.sh

# Production build: signed + notarized .mcpb (also signs the per-arch
# tarballs; takes 30s-2min for the notary round-trip with Apple).
make pack-dxt-signed

# Verify the .mcpb. Look for the four green lines under
# "Checking macOS code signing".
make verify-dxt

# Read-only health check (no Apple round-trip). Use in CI / new-machine
# preflight.
make doctor-signing

# Dev iteration: unsigned .mcpb. Don't ship this; use for local testing
# only.
make pack-dxt
```

---

## 6. Troubleshooting matrix

| Symptom | Likely cause | Fix |
|---|---|---|
| `0 valid identities found` | No Developer ID cert in keychain. | Run `./scripts/setup-signing.sh`. |
| `unable to build chain to self-signed root` | Missing G2 intermediate, OR stale `trustAsRoot` override on it. | Re-run setup script (auto-installs intermediate, no override). For manual cleanup: `security remove-trusted-cert /tmp/DeveloperIDG2CA.cer`. |
| `SecKeychainItemImport: Unknown format` | Trying to import a modern .p12. | Don't. Import key + cert as separate files. The setup script does this. |
| `Error 73` from `xcrun stapler staple` | Trying to staple a raw Mach-O. | Expected; ignore. The Makefile already does (`-` prefix). Notarization works without stapling via online verification. |
| `HTTP status code: 401. Invalid credentials.` from notarytool | App-specific password is stale. | Generate new at appleid.apple.com, update `scripts/release-env.sh`, re-run setup. |
| `Error: DEVELOPER_ID must be set` from `make pack-dxt-signed` | Forgot `source scripts/release-env.sh` in the current shell. | `source scripts/release-env.sh && make pack-dxt-signed`. |
| `spctl: rejected (the code is valid but does not seem to be an app)` | Using wrong tool to check raw Mach-O notarization. | Ignore that line. Use `codesign --test-requirement="=notarized"` instead. `make verify-dxt` already does. |
| `xcrun notarytool` hangs >5min on `--wait` | Apple's queue is occasionally slow. | Wait 10 min total; if still stuck, Ctrl-C and `xcrun notarytool log <submission-id> --keychain-profile slide-mcp-release` for diagnostics. |
| Build succeeds but `verify-dxt` shows `notarization: NOT notarized` | `pack-dxt-signed` errored mid-pipeline (probably env not exported), produced signed but un-notarized binary. | `source scripts/release-env.sh && make notarize-darwin-universal && make pack-dxt-signed`. |

---

## 7. Cert rotation (when the cert expires)

Developer ID Application certs are valid for 5 years. When yours
expires:

1. Keep `~/.config/slide-mcp-signing/signing.key` - the private key is
   reusable. Do not regenerate.
2. Run `openssl req -new -key ~/.config/slide-mcp-signing/signing.key
   -out ~/.config/slide-mcp-signing/signing.csr -subj
   "/emailAddress=$APPLE_ID/CN=Slide MCP Signing/C=US"` to create a
   fresh CSR from the same key.
3. Upload the CSR at <https://developer.apple.com/account/resources/certificates/add>,
   download the new `.cer`.
4. Re-run `./scripts/setup-signing.sh --force` to re-import (use
   `--force` to bypass the "identity already present" check).
5. Old cert can be deleted from Keychain Access by hand. Notarytool
   profile and app-specific password are unaffected by cert rotation.

---

## 8. What not to do

- Do not commit the `.cer`, the private key, the app-specific
  password, the notarytool credentials, or any decrypted form of
  `release-env.sh`. The `.gitignore` already covers everything that
  matters; if you find yourself adding new patterns, also add to it.
- Do not add `add-trusted-cert -r trustAsRoot` calls anywhere.
  See section 3.5.
- Do not switch to `.p12`-based imports because "the script is too
  long". See section 3.3 - it doesn't work on modern macOS.
- Do not ship unsigned or unnotarized `.mcpb` files. Always
  `make pack-dxt-signed`, never `make pack-dxt`, for distribution.
- Do not use `spctl --type execute` to check whether a raw Mach-O is
  notarized. See section 3.7.
- Do not silence Apple's error output in any new tooling you add. The
  setup script's notarytool failure path was cryptic until we taught
  it to print Apple's actual error indented under the failing step.

---

## 9. Reference: what a known-good signed binary looks like

After `make pack-dxt-signed` on a properly-set-up machine, the inner
universal binary will pass these checks:

```bash
$ codesign -dvv build/slide-mcp-server-darwin-universal
...
Authority=Developer ID Application: <Your Name> (<TEAMID>)
Authority=Developer ID Certification Authority
Authority=Apple Root CA
Timestamp=<RFC3339 from Apple TSA>
TeamIdentifier=<TEAMID>
Runtime Version=12.0.0

$ codesign --verify --verbose=2 build/slide-mcp-server-darwin-universal
...: valid on disk
...: satisfies its Designated Requirement

$ codesign --test-requirement="=notarized" --verify --verbose=2 \
       build/slide-mcp-server-darwin-universal
...: valid on disk
...: satisfies its Designated Requirement
...: explicit requirement satisfied

$ xcrun notarytool history --keychain-profile slide-mcp-release | head
# Should list your recent submission with status: Accepted
```

If all four pass, the `.mcpb` will install cleanly on every Mac with
internet access. Done.
