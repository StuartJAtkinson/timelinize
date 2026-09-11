# Considerations

## Open questions for Auto Continue

One line per question, `- ` prefixed — that is the only shape
`consideration_items` (atelier-harness `meta/markdown.rs:280`) recognises.
Answer them inline when atelier interviews this project.

- The "convert to SVG" decision covered the two JPGs and `archivebox.png`, which were flat logos I could redraw faithfully; five raster icons remain (`apple_contacts.png`, `email.png`, `flighty.png`, `media.png`, `sms_backup_restore.png`, shown together in `preview-remaining-pngs.png`). They are detailed multi-colour artwork that already has transparency, so they don't show the white-box problem the JPGs did — hand-vectorising them would change how they look. Convert these five too (accepting a redrawn, simplified look), re-source them as official vendor SVGs, or keep them as PNG and write "PNG is fine for detailed artwork, SVG for flat logos" down as the rule?

