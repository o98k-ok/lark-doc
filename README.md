<div align="center">

# Lark Doc

**Search Feishu documents from Alfred. Sign in with a QR code.**

Go CLI · Alfred workflow · Local credentials

[Get started](#get-started) · [Usage](#usage) · [Privacy](#privacy-and-local-data) · [Troubleshooting](#troubleshooting) · [Development](#development)

</div>

Lark Doc sends your keywords to Feishu and returns document titles, previews,
authors, and links in Alfred. Results are sorted by last visit time. Select a
document and press Return to open it in your browser.

Login happens on your machine: scan a terminal QR code with the Feishu app, then
confirm on your phone. You do not need to copy browser cookies or register a
developer application.

> [!IMPORTANT]
> This version is under development. QR login and keyword search have been tested
> against Feishu from Linux. The newer macOS Keychain implementation and the
> Alfred workflow have **not yet completed native end-to-end verification**.
> Build from source for evaluation; do not treat this as a verified release.

## What it does

| Feature | Behavior |
| --- | --- |
| Keyword search | Queries Feishu directly, rather than filtering a local recent-document list |
| QR login | Saves the confirmed session locally; prompts for login again when it expires |
| Alfred actions | Search, open document links, start login, and clear local login data |
| Optional cache | Stores document metadata for empty queries; populated with the CLI `trigger` command |
| Credential storage | macOS Keychain; owner-only JSON file on Linux |

The current implementation targets Feishu at `www.feishu.cn`. International Lark
domains are not configurable in this version. It uses private Feishu web APIs,
not the official OpenAPI, so upstream changes can break login or search.

## Get started

### 1. Prepare your Mac

| Requirement | Needed for |
| --- | --- |
| Alfred with Powerpack | Running the workflow |
| Go 1.26 or newer | Building the executable; the module selects toolchain Go 1.27.3 |
| Xcode Command Line Tools and CGO enabled | Building the native Keychain integration |
| Python 3 | Packaging the workflow |
| Feishu mobile app and an account | Scanning and confirming login |

Alfred is not required for CLI-only use. Linux supports the CLI, but does not use
macOS Keychain.

### 2. Build and import

Run these commands on the Mac where you intend to use the workflow:

```sh
git clone https://github.com/o98k-ok/lark-doc.git
cd lark-doc
CGO_ENABLED=1 go build -o lark-doc .
python3 workflow/package.py ./lark-doc ./Lark-Doc.alfredworkflow
open ./Lark-Doc.alfredworkflow
```

Confirm the import in Alfred. The package includes the executable, workflow
manifest, and two action scripts. It does not bundle credentials or document caches.

Build separately for Apple Silicon and Intel Macs, or supply a universal macOS
executable to the packaging script. Do not package a Linux executable for Alfred.

### 3. Scan and search

1. Open Alfred, type `lark-login`, and press Return.
2. Terminal opens with a QR code. Scan it with Feishu and confirm on your phone.
3. Wait for the login success message, then return to Alfred.
4. Type `lark your keywords`. Select a result and press Return.

If you search before logging in, choose the **登录飞书** result to start login.
macOS may request permission to control Terminal or access Keychain.

QR codes regenerate when they expire. Press `Ctrl+C` to cancel; a login attempt
times out after five minutes. Sessions do not renew automatically. Repeat login
when the saved session expires.

## Usage

### Alfred

| Input | Action |
| --- | --- |
| `lark <keywords>` | Search documents accessible to the signed-in account |
| `lark` | Show cached results if available, otherwise query Feishu with no keywords |
| `lark-login` | Open Terminal for QR login |
| `lark-logout` | Delete the local credential and current document cache |

### Command line

```sh
# Sign in on this machine
./lark-doc login

# Search and return Alfred-compatible JSON
./lark-doc query "project notes"

# Return up to 15 results
count=15 ./lark-doc query "project notes"

# Fetch a document batch and replace the local cache
./lark-doc trigger

# Read the cache, or query Feishu if there is no cache
./lark-doc query

# Remove the local credential and current cache
./lark-doc logout
```

The cache is not a full-text index or a document backup. It contains metadata,
including titles, previews, authors, links, and timestamps. Keyword queries still
go to Feishu. The program requires a locally available credential even when
showing cached results.

### Configuration

| Variable | Default | Scope |
| --- | --- | --- |
| `count` | `9` | Maximum displayed results; non-positive or invalid values fall back to 9 |
| `session` | Saved credential | Optional manual override for CLI use only |

The bundled Alfred workflow ignores `session` and uses the local credential
store. Prefer QR login. Do not put session values in shell history, workflow
variables, screenshots, or shared files. `cache_path` is no longer supported.

## Privacy and local data

### Where data goes

Login and search requests go directly from the machine running Lark Doc to Feishu
over HTTPS. There is no application backend between the plugin and Feishu.
Search terms are sent to Feishu; document metadata is returned to Alfred or CLI
stdout. Debug output, terminal recordings, and shared search results can therefore
contain private document information.

The program does not read browser Cookie databases or browser Keychain entries.
On macOS, it uses its own Keychain entry under service `com.o98k.lark-doc`.

> [!WARNING]
> A session cookie is a web login credential, not a read-only OAuth token scoped
> to document search. Protect it like a password. Keychain reduces plaintext
> exposure; it does not make a compromised machine safe.

### Storage locations

| Data | macOS | Linux |
| --- | --- | --- |
| Login credential | Keychain, service `com.o98k.lark-doc` | `$XDG_CONFIG_HOME/lark-doc/session.json`, normally `~/.config/lark-doc/session.json` |
| Document cache | `~/Library/Caches/lark-doc/documents.json` | `$XDG_CACHE_HOME/lark-doc/documents.json`, normally `~/.cache/lark-doc/documents.json` |

Linux credentials and document cache files use owner-only permissions (`0600`).
These files are not encrypted by the application. Newly created storage
directories use `0700`. Successful login clears the previous account's cache
before saving the new credential; a cache cleanup failure stops the account change.

### Logout and upgrades

`logout` deletes the current local credential and cache, plus the earlier
plaintext session file at the standard configuration path. **It does not revoke
the Feishu server-side session**, delete backups, or erase previous terminal
output. Use Feishu's session/device controls when server-side revocation is needed.

If upgrading from the plaintext-session version on macOS, log in again. Successful
Keychain storage removes the old `~/Library/Application Support/lark-doc/session.json`.
The old file is not automatically imported.

Earlier `lark.json` caches and files created with a custom `cache_path` are no
longer used and are not automatically deleted. Remove those copies yourself,
particularly from synced folders or workflow directories.

## Troubleshooting

| Symptom | What to check |
| --- | --- |
| Alfred does not import or run the workflow | Confirm Alfred Powerpack is enabled and the packaged executable matches your Mac |
| Build reports missing Keychain symbols | Build on macOS with CGO enabled and Xcode Command Line Tools installed |
| Terminal does not open | Check macOS Automation permissions for Alfred controlling Terminal; alternatively run `./lark-doc login` |
| Keychain is unavailable | Unlock Keychain, review the system permission prompt, and retry login |
| Search offers login again | The credential may be missing, inaccessible, or expired; log in again |
| Login fails while clearing the cache | Check access to the cache path; the new credential is not saved until cleanup succeeds |
| Search returns no matching documents | Check the keywords and the account's document permissions in Feishu |
| Login works but search fails | Private web APIs can change; organization restrictions may also apply. Compatibility across tenants is not guaranteed |

When reporting an issue, include the command, OS/architecture, and a redacted
error message. Do not attach session files, live QR codes, document caches, or
unredacted internal search results.

## Development

```sh
go test -race ./...
go vet ./...
go mod verify
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Tests cover the QR protocol using synthetic responses, credential persistence,
cache permissions and cleanup, and search with saved or manually supplied
sessions. macOS tests can create and delete synthetic Keychain entries; Linux
tests cannot validate the native Keychain implementation.

| Area | Verification status |
| --- | --- |
| Linux tests, build, and vulnerability scan | Passed during development |
| Live Feishu QR login and keyword search | Exercised with one account on Linux |
| macOS Keychain implementation | Native verification pending |
| Alfred import, login action, and result selection | In-app verification pending |

The QR login protocol follows the approach in `deal/scripts/lark`. Lark Doc keeps
its original server-side keyword search rather than switching to recent-document
filtering.

## Earlier version preview

<details>
<summary>Historical screenshot and demo</summary>

These assets show the earlier version, not the current QR login flow. Do not
follow the old manual-cookie setup shown in the demo. Historical assets may
contain account or workspace details; they are not sanitized sample data.

<img width="718" alt="Earlier version of Lark Doc in Alfred" src="https://user-images.githubusercontent.com/63107263/209918773-731dfbb4-d5bc-4abe-91d1-4fa479db7c4b.png">

![Earlier Lark Doc workflow demo](assets/output.gif)

</details>
