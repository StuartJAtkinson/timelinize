# Considerations

## Open questions for Auto Continue

One line per question, `- ` prefixed — that is the only shape
`consideration_items` (atelier-harness `meta/markdown.rs:280`) recognises.
Answer them inline when atelier interviews this project.

## Open questions for Auto Continue

One line per question, `- ` prefixed — the only shape `consideration_items` recognises.

- `preview-icons.png` (580×160, 17 KB, four datasource icons on a dark background) sits at the repo root and is referenced by nothing — no HTML, no CSS, no MD, no JS. Stale dev artifact, or somewhere it should live?
- Destructive-action red uses two vocabularies: `text-red` (Tabler palette utility, on `frontend/pages/map.html:102`, `frontend/resources/js/dashboard.js:74,79,411-416`, `frontend/resources/js/jobs.js:230-469`) and `text-danger` (Bootstrap semantic, on `frontend/pages/import.html:372`, `frontend/resources/html/includes/modals.html:403,410`). Same red, two names — pick one as canonical.
- `btn-orange` on `frontend/pages/job.html:71` (Restart) and `btn-warning` on `frontend/pages/job.html:35` / `frontend/pages/entity.html:35` / `frontend/pages/entities.html:45,121` (Merge). Both destructive-ish. Should Restart match the Merge amber, or stay the more saturated Tabler orange?

