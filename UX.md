# UX conventions — timelinize

Verified conventions derived from a look-at-it pass over `frontend/pages/*.html` and `frontend/resources/css/*.css`. New code should match these unless there's a reason to deviate; the reason goes into ISSUES.md.

## Page chrome

Every page opens with the same page-header skeleton, except `setup.html` (wizard) and `input.html` (manually-edited card):

```html
<div class="page-header d-print-none">           <!-- add page-header-overlap + text-white for dark dashboards -->
  <div class="container-xl">
    <div class="row g-2 align-items-center">
      <div class="col">                         <!-- sometimes an icon SVG is nested before the title -->
        <div class="page-pretitle">Explore</div>  <!-- or "Overview" / "Item" / "Entity" -->
        <h2 class="page-title">…</h2>
        <div class="text-secondary mt-1">…</div>  <!-- optional subtitle -->
      </div>
      <!-- optional col-auto ms-auto d-print-none for the page-level action(s) -->
    </div>
  </div>
</div>

<div class="page-body">
  <div class="container-xl">
    …
  </div>
</div>
```

- `<title>` matches the `<h2 class="page-title">` text (or its module label) so the browser tab is useful. `entity.html` and `item.html` are the two exceptions — both carry `<!-- TODO: use JS to change the title -->` and currently hardcode "Entity" / "Item".
- Page-level primary action lives in the right-hand `col-auto ms-auto d-print-none` slot (`btn-list` wrapper). Examples: Dashboard has none, Entities has the "Merge" button (`btn-warning`), Item has Download (`btn-primary`) + Edit (`btn-outline disabled`) + Delete (`btn-outline-danger disabled`), Import has "Add files..." (`btn-primary`).
- Page subtitle ("Showing X most recent", "Showing results 1-50") lives directly under the title in `text-secondary mt-1`.

## Buttons — what colour means what

- `btn` (plain): row actions and inside-content actions (e.g. "View" on a list row, "Cancel" in some contexts).
- `btn btn-primary`: the single main page-level action, plus Save/Create-type confirms.
- `btn btn-outline-danger`: destructive confirms and row-level danger actions (Delete, Remove).
- `btn btn-warning`: "Merge" — the one action whose result bites (entity merge), so amber everywhere it appears (page, table row, modal).
- `btn btn-link link-secondary`: modal Cancel — lighter than the confirm it sits beside.
- `btn btn-ghost-info` (Tabler class): "Expand all" / "Collapse all" — informational toggles that don't change data.

## Filter UIs

Conversations, Map, Gallery, Timeline all share the same filter-column pattern:

```html
<div class="col-3 filter-column">       <!-- conversations -->
<div id="filter-column" class="col-sm-4 col-lg-3">  <!-- gallery -->
<div class="filter col-3">              <!-- timeline -->
```

The filter column opens with the same `<div class="page-pretitle">Explore</div>` + `<h2 class="page-title">` + (optional icon SVG) header as the rest of the app — never a bare `<h2>`.

## Tables

- Row of action buttons lives in `<div class="btn-list flex-nowrap">`.
- Row action "View" is `btn` (plain) when there's a higher-level page action elsewhere on the row.

## Cards / datagrids

- Definition-style rows (label : value) use Tabler's `.datagrid` / `.datagrid-item` / `.datagrid-title` / `.datagrid-content`. Subheaders in cards are `subheader` (`text-secondary`, small caps in Tabler).

## Dark-theme

- `--tblr-*` tokens drive everything in light *and* dark themes. Literal hex / rgb colours are avoided in app CSS — `frontend/resources/css/items.css:51` already shows the migration pattern (`var(--tblr-bg-surface-secondary)` replaces `#f5f5f5` so the dark theme reads correctly).
- Items carry a `clickable` class for cards that open their detail page; `.clickable:hover` is tokenised.
