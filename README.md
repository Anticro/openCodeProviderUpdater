# OpenCode Provider Updater

A CLI tool with an optional web UI that updates the `provider` section of an [opencode](https://opencode.ai) configuration file (`~/.config/opencode/opencode.json`).

It discovers installed models by querying a provider's `/api/tags` endpoint (standard for Ollama-compatible APIs) and automatically rebuilds the `models` map in the config.

## Motivation

When you pull new models via Ollama (or compatible backends), your opencode config's `models` section becomes stale. Manually editing the JSON is tedious and error-prone. This tool automates the sync.

## Features

- **Web UI** — run with `-ui` to launch a browser interface at `http://localhost:8099`. Cards show each provider's name, address, and installed models. One-click update, modal dialog to add new providers, and graceful auto-shutdown when the browser tab closes.
- **Update existing provider** — picks an existing provider from the config, fetches its current model list, and replaces only the `models` section.
- **Add a new provider** — interactively or via the web UI, configures a new Ollama-compatible provider, including model discovery.
- **List providers** — option 3 in the CLI prints all configured providers and their models without making any changes.
- **Automatic backups** — the original config file is renamed to `opencode.json_<unix_timestamp>` before any modification (atomic rename, not copy).
- **Preserves unrelated config keys** — only the `provider` key is touched; everything else remains intact.

## Prerequisites

- Go 1.21+ (to build from source)
- An opencode configuration file at `~/.config/opencode/opencode.json`
- One or more running Ollama (or Ollama-compatible) API servers

## Installation

```sh
git clone <repository-url>
cd openCodeProviderUpdater
go build -o opencode-provider-updater .
```

This produces a standalone binary in the current directory.

## Usage

### Web UI

Run with `-ui` to start the interactive web interface:

```sh
./opencode-provider-updater -ui
```

The tool:

1. Starts an HTTP server on port **8099**.
2. Opens your default browser to `http://localhost:8099`.
3. Displays provider cards with name, address, and installed models.

From the UI you can:

- **Update** a provider — click the green "Update" button on any card to refetch its model list from the API.
- **Add** a provider — click "+ Add New Provider", fill in the base URL and display name in the modal, and the tool discovers and registers the provider.
- **Graceful shutdown** — closing the browser tab automatically sends a shutdown signal and stops the server.

### CLI

Run the binary without arguments:

```sh
./opencode-provider-updater
```

You are presented with a menu:

```
1. Update an existing provider
2. Add a new provider
3. List all providers and models
Choose an option (1-3):
```

#### Option 1: Update an existing provider

Use this when you've pulled new models and want to refresh the config.

```
Available providers:
1. ollama
Choose a provider (1-based, 0 to abort): 1
```

The tool:

1. Reads the `baseURL` from the selected provider's `options`.
2. Strips any path (e.g. `/v1`) and appends `/api/tags`.
3. Sends a GET request to discover installed models.
4. Replaces only the `models` key of the selected provider.
5. Backs up the original config to `opencode.json_<timestamp>`.
6. Writes the updated config and prints the new file contents.

#### Option 2: Add a new provider

Use this to register a new Ollama-compatible provider.

```
Enter base URL (e.g. http://localhost:11434): http://192.168.1.50:11434
Enter provider name (e.g. Ollama local): Ollama (workstation)
```

The tool:

1. Appends `/api/tags` to the base URL and fetches the model list.
2. Derives a JSON-safe key from the URL host (non-alphanumeric characters replaced with `_`).
   - `http://192.168.1.50:11434` → key `192_168_1_50_11434`
   - `http://localhost:11434` → key `localhost_11434`
3. Creates the full provider block with:
   - `models` — discovered from the API response
   - `name` — the name you entered
   - `npm` — set to `@ai-sdk/openai-compatible`
   - `options.baseURL` — your base URL with `/v1` appended
4. If no `provider` object exists yet in the config, it is created.
5. Backs up the original config, writes the update, and prints the new file contents.

#### Option 3: List all providers and models

Read-only mode that prints every configured provider with its display name and installed model list. Does not modify the config or contact any API.

#### Aborting

At the provider selection prompt (update mode), entering `0` exits without making any changes.

## Config file format

The program operates on `~/.config/opencode/opencode.json`, which has the following structure:

```json
{
   "other_keys": {...},
   "provider": {
      "ollama": {
         "models": {
            "gemma4:26b": {
               "name": "gemma4:26b"
            }
         },
         "name": "Ollama (local)",
         "npm": "@ai-sdk/openai-compatible",
         "options": {
            "baseURL": "http://localhost:11434/v1"
         }
      }
   }
}
```

- Keys at the top level that are not `provider` are preserved as-is.
- Each provider key (e.g. `ollama`) is derived from the URL host when adding.
- The `models` object maps each model name to `{"name": "<model-name>"}`.
- The `options.baseURL` must point to an OpenAI-compatible endpoint (the tool adds `/v1` for new providers and strips it to reach `/api/tags`).

## API compatibility

The tool expects the Ollama [Tags API](https://github.com/ollama/ollama/blob/main/docs/api.md#list-local-models) at `{baseURL}/api/tags`, which returns:

```json
{
   "models": [
      {"name": "gemma4:26b", "model": "gemma4:26b", ...}
   ]
}
```

Any server implementing this endpoint (Ollama, LM Studio, etc.) is supported.

## REST API (web UI mode)

When running with `-ui`, the following endpoints are available:

| Method | Endpoint                  | Description                           |
|--------|----------------------------|----------------------------------------|
| `GET`  | `/api/providers`  | Returns JSON array of all providers    |
| `POST` | `/api/providers/update` | Body: `{"key": "ollama"}`              |
| `POST` | `/api/providers/add`    | Body: `{"baseURL": "..","name": "..."}` |
| `POST` | `/api/shutdown`         | Gracefully stops the server            |

## Backups

Before every write, the existing config is atomically renamed to:

```
~/.config/opencode/opencode.json_<unix-timestamp>
```

If something goes wrong, restore from the backup:

```sh
cp ~/.config/opencode/opencode.json_<timestamp> ~/.config/opencode/opencode.json
```

## Build

```sh
go build -o opencode-provider-updater .
```

Cross-compile for other platforms:

```sh
GOOS=linux GOARCH=amd64 go build -o opencode-provider-updater-linux .
GOOS=windows GOARCH=amd64 go build -o opencode-provider-updater.exe .
```