> [!WARNING]
> Potok is in active development. Sync (push/pull/watch) is not implemented yet — only local vault registration and remote listing/deletion work today.

# Potok

**Potok** is a self-hosted, CLI-based tool for backing up and syncing [Obsidian](https://obsidian.md/) vaults with end-to-end encryption. Your notes stay yours — the server never sees your passwords or unencrypted data.

For more detail, check out the [Potok Docs](https://potok-docs.vercel.app/)

## Features

- **End-to-End Encryption** — Planned: vaults encrypted locally before upload. Crypto package is still a stub.
- **Self-Hosted** — Run your own Potok server (`potokd`). No third-party cloud, no vendor lock-in.
- **Multiple Vaults** — Register and manage multiple vaults locally.
- **Automatic Sync** — Not implemented yet.
- **Cross-Platform** — Client builds for Windows, macOS, and Linux.
- **Secure Key Storage** — API keys and vault passphrases are stored in your OS keyring (Windows Credential Manager, macOS Keychain, Linux Secret Service).
- **Free & Open Source** — No file size limits, no file count limits, no paywalls.

## Commands

| Command | Description | Status |
|---|---|---|
| `potok init` | Set server URL and API key | Available |
| `potok vault-add <name>` | Register a local folder as a vault | Available |
| `potok vaults-list` | List vaults registered locally | Available |
| `potok vault-remove <name>` | Remove a vault from local config | Available |
| `potok remote-list` | List vaults available on the server | Available |
| `potok remote-delete <name>` | Delete a vault from the server | Available |
| `potok doctor` | Run diagnostics on your setup | Available |
| `potok push` | Encrypt and upload a vault | Not yet |
| `potok pull` | Download and decrypt a vault | Not yet |
| `potok sync` | Watch and auto-sync a vault | Not yet |

## Getting Started

### Prerequisites

- A running Potok server
- An API key from your server admin
- Go 1.25+ (if building from source)

### Install

```bash
go install github.com/mtiluk/potok/cmd/potok@latest
```

Server binary (optional):

```bash
go install github.com/mtiluk/potok/cmd/potokd@latest
```

### Initialise

```bash
potok init
```

You'll be prompted for your server URL and API key. The URL is stored in `~/.potok/config.json`; the API key goes in your OS keyring.

## Usage

### Register a vault

```bash
potok vault-add notes
```

Prompts for a vault path and encryption passphrase. This only registers the vault locally — nothing is uploaded yet.

### List local vaults

```bash
potok vaults-list
```

Shows all vaults registered on this device with their path and last sync time.

### Remove a local vault

```bash
potok vault-remove notes
```

Removes the vault from local config and deletes its stored passphrase from the keyring.

### List remote vaults

```bash
potok remote-list
```

Lists vault names currently on the server.

### Delete a remote vault

```bash
potok remote-delete notes
```

Deletes a vault on the server. Does not change local registration.

### Check your setup

```bash
potok doctor
```

Loads config, checks the API key in the keyring, and hits `/health` and `/me`.

## Configuration

### Config file

| OS | Path |
|---|---|
| Linux / macOS | `~/.potok/config.json` (or `$XDG_CONFIG_HOME/potok/config.json`) |
| Windows | `%USERPROFILE%\.potok\config.json` |

Override with `POTOK_CONFIG_DIR`.

Fresh vault (just registered locally — `remote_id` and `last_synced_at` are omitted until the first sync):

```json
{
  "server_url": "http://localhost:8080",
  "vaults": [
    {
      "name": "journal",
      "path": "/home/user/Documents/Obsidian/Journal"
    }
  ]
}
```

Existing synced vault:

```json
{
  "server_url": "http://localhost:8080",
  "vaults": [
    {
      "name": "notes",
      "path": "/home/user/Documents/Obsidian/Notes",
      "remote_id": "vault_abc123",
      "last_synced_at": "2026-09-01T12:00:00Z"
    }
  ]
}
```

### Sensitive data

Passwords and API keys are stored in your OS keyring under the `potok` service — never in config files.

| OS | Keyring backend |
|---|---|
| Linux | Secret Service (GNOME Keyring / KDE Wallet) |
| macOS | Keychain |
| Windows | Credential Manager |

| Keyring entry | Value |
|---|---|
| `potok / api-key` | Your server API key |
| `potok / vault:{name}` | Encryption passphrase for that vault |

## Security

- Encryption and decryption are intended to happen locally; the crypto package is not wired into CLI commands yet.
- The server stores encrypted blobs once push/pull land — it should never see your passwords or plaintext.
- Passwords and API keys are stored in your OS keyring, not in config files.

## Roadmap

- [x] CLI skeleton and local vault management (`init`, `vault-add`, `vaults-list`, `vault-remove`, `doctor`)
- [x] OS keyring integration for passwords and API keys
- [x] Remote vault list and delete (`remote-list`, `remote-delete`)
- [ ] Push — encrypt and upload vaults
- [ ] Pull — download and decrypt vaults
- [ ] Automatic file watching and sync
- [ ] File-level sync
- [ ] Conflict detection and handling
- [ ] Version history
- [ ] Web dashboard for server admin
- [ ] Cross-platform installers
