# MCP — timelinize

**Design spec.** No MCP server exists yet.

- **Proposed server:** `timelinize`
- **Transport:** stdio
- **Backs onto:** the existing local HTTP API at `http://127.0.0.1:12002/api/`
  (`tlzapp/endpoints.go`, `defaultAdminAddr`)

## Fork context — read this first

This is **a fork of `timelinize/timelinize`**, not an original project.
`upstream` is set to `DO-NOT-PUSH-TO-UPSTREAM`. Two consequences for anything
built from this spec:

1. An MCP server added here is **fork-local** unless deliberately proposed
   upstream. Keep it in one new package so a rebase onto upstream stays cheap —
   do not thread MCP concerns through `tlzapp/`.
2. `endpoints.go` is upstream's file. The tool list below is a *selection from*
   it, not a request to change it.

## Why this repo wants one

Timelinize already exposes a complete, documented local API — every endpoint
carries a `Help` string and a typed `Payload`. The MCP surface is close to a
mechanical wrapper, which makes the interesting work not "what tools" but
**"which endpoints must never become tools"** (below).

The query worth having is the one the UI is bad at: *"what was I doing in
March 2019"*, *"find every conversation with this person"* — cross-source
recall over a timeline that is too large to scroll.

## Tools

All map to `POST`/`GET http://127.0.0.1:12002/api/<name>`.

| Tool | Endpoint | Returns |
|---|---|---|
| `search_items` | `search-items` (POST, `ItemSearchParams`) | items matching a filter — the core recall tool |
| `search_entities` | `search-entities` (POST, `EntitySearchParams`) | people/accounts matching a filter |
| `get_entity` | `get-entity` | one entity |
| `conversation` | `conversation` | one conversation's messages |
| `recent_conversations` | `recent-conversations` | recent threads |
| `list_repositories` | `open-repositories` | which timelines are open |
| `list_data_sources` | `data-sources` | importers this build supports |
| `item_classifications` | `item-classifications` | the classification vocabulary — needed to write a sane `search_items` filter |
| `charts` | `charts` (GET) | timeline statistics |
| `jobs` | `jobs` | job status, read-only |
| `build_info` | `build-info` | version |

`item_classifications` earns its place: without the vocabulary a model guesses
filter values and gets empty results it will read as "no such memory".

## Resources

| URI | Contents |
|---|---|
| `tlz://repositories` | open repositories |
| `tlz://data-sources` | available importers |

## What must NOT be a tool

This list is the point of the document.

**Destructive:**
- `delete-items` — permanent removal from an archive whose entire purpose is
  that nothing is lost. Never.
- `merge-entities` — merging two people is **irreversible in practice** and a
  model has exactly the wrong instinct here: two accounts with similar names
  look mergeable and often aren't. This repo already treats merge as cautionary
  in the UI (commit `99fa720`); an unattended merge tool contradicts that
  decision outright.

**State-changing:**
- `import`, `plan-import`, `start-job`, `pause-job`, `unpause-job`,
  `cancel-jobs` — long-running ingests over personal data. `jobs` is exposed
  read-only so an agent can *watch*, not steer.
- `change-settings`, `open-repository`, `close-repository` — closing a
  repository out from under the running UI is a support call waiting to happen.
- `add-entity`, `submit-graph`, `next-graph` — writes to the entity graph.

**Filesystem:**
- `file-stat`, `file-listing`, `file-selector-roots` — these are a general
  filesystem browser reachable over the API. Exposing them through MCP turns a
  timeline server into an arbitrary-directory reader, which is not what anyone
  installing it thinks they are getting.

**Logs:**
- `logs` — leaks paths and imported filenames.

## The boundary that matters here

A timelinize repository is **the user's entire personal history** — photos,
messages, location traces, contacts, for them and everyone they have ever
talked to. Not their data alone.

- **stdio and localhost only.** The API already binds `127.0.0.1`. Never add a
  remote transport, never proxy it.
- **Default to counts and structure.** `charts` and result counts answer far
  more questions than item bodies do, and cost nothing in exposure.
- **`search_items` with a wide filter dumps a life into a context window.** Cap
  the page size in the server, not by asking the model to be careful.
- **Demo mode exists** (the README's obfuscation mode). If it can be driven
  per-request, the MCP path is the obvious place to default it on.

## Implementation note

Wrap the HTTP API rather than importing `timeline/` directly. It keeps the
server in one file, it keeps the fork rebasable, and it means the MCP path sees
exactly what the frontend sees — one API, one set of semantics, no second
implementation to drift.

The app must already be running (`http://127.0.0.1:12002`). Fail with that
sentence, not a connection-refused trace.
