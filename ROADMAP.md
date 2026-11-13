# Roadmap — timelinize (personal fork)

> **Auto Continue reads this file.** The `feature` phase takes the first unchecked
> `- [ ]` line below as its whole brief and ticks it by exact text, so milestones and
> phases are headings and every slice is one commit-sized, self-contained line.
> **Human-only** items carry no checkbox, so Auto never picks them up.

**Now:** Milestone 3 — *Interactive review*. Milestones 1 and 2 are shipped: the fork is
canonical, upstream PR work was ported, and socialMediaArchiver archives are supported.

## Milestone 2 — Own sources

Timelinize has no Bluesky, Mastodon, Reddit or YouTube data source, and
socialMediaArchiver (`H:\GitHub\socialMediaArchiver`) already normalises all of those (plus
X, RSS and Facebook Pages) into one `Item` schema (`core/models.py`: author, text, timestamp,
media with `local_path`). One importer for that archive covers all of them.

- [x] **M2a socialMediaArchiver data source** — new `datasources/socialmediaarchiver/` that recognises a socialMediaArchiver output folder and imports each normalised item as a timeline item: timestamp, text, author as entity (one per platform account), media from `local_path`, original URL kept. Registered in `datasources.go`, with an SVG icon per the datasources icon rule. Go test over a small fixture archive (one post per platform). Done when `go test ./datasources/socialmediaarchiver/...` passes.
- [x] **M2b Dedup against platform exports** — when the same post also arrives through an existing source (e.g. `twitter` from an official archive), the socialMediaArchiver item merges into it instead of duplicating, keyed on the platform's post id. Done when a test importing both shows one item.

## Milestone 3 — Interactive review

- [x] **M3a Interactive import review** — stage each incoming graph and its content, hold the import until the user saves or skips it, and restore staged readers before processing. The UI previews content, supports timestamp and classification edits, preserves pending state across reloads, and reports completion after the graph stream closes.

## Human-only

- **Point it at the real archive** — import your socialMediaArchiver output folder through the UI once M2a lands.
