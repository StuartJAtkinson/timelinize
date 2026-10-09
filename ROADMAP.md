# Roadmap — timelinize (personal fork)

> **Auto Continue reads this file.** The `feature` phase takes the first unchecked
> `- [ ]` line below as its whole brief and ticks it by exact text, so milestones and
> phases are headings and every slice is one commit-sized, self-contained line.
> **Human-only** items carry no checkbox, so Auto never picks them up.

**Now:** Milestone 2 — *Own sources*. Milestone 1 (fork made canonical, module path
renamed, every open upstream PR — #106, #163, #184, #186, #197, #198 — ported) is shipped;
upstream is archived, so there is nothing left to harvest.

## Milestone 2 — Own sources

Timelinize has no Bluesky, Mastodon, Reddit or YouTube data source, and
socialMediaArchiver (`H:\GitHub\socialMediaArchiver`) already normalises all of those (plus
X, RSS and Facebook Pages) into one `Item` schema (`core/models.py`: author, text, timestamp,
media with `local_path`). One importer for that archive covers all of them.

- [x] **M2a socialMediaArchiver data source** — new `datasources/socialmediaarchiver/` that recognises a socialMediaArchiver output folder and imports each normalised item as a timeline item: timestamp, text, author as entity (one per platform account), media from `local_path`, original URL kept. Registered in `datasources.go`, with an SVG icon per the datasources icon rule. Go test over a small fixture archive (one post per platform). Done when `go test ./datasources/socialmediaarchiver/...` passes.
- [ ] **M2b Dedup against platform exports** — when the same post also arrives through an existing source (e.g. `twitter` from an official archive), the socialMediaArchiver item merges into it instead of duplicating, keyed on the platform's post id. Done when a test importing both shows one item.

## Human-only

- **Point it at the real archive** — import your socialMediaArchiver output folder through the UI once M2a lands.
