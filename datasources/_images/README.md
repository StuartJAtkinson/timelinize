# Datasource icons

One file per datasource, named in that datasource's `Icon:` field.

**SVG for flat logos, PNG for detailed artwork.** A flat logo — solid shapes, a
wordmark, a single glyph — is drawn as SVG so it stays crisp at any size.
Detailed multi-colour artwork (`apple_contacts.png`, `email.png`,
`flighty.png`, `media.png`, `sms_backup_restore.png`) stays PNG: hand-vectorising
it would change how it looks, for no gain.

**PNG is only acceptable with an alpha channel.** The icons render on a dark
theme, so an opaque raster shows as a white box behind the logo. That is what
retired the last two JPGs. Check before adding one: the PNG must be colour type
4/6, or type 3 with a `tRNS` chunk.

# Assets license

## firefox.svg

From https://github.com/alrra/browser-logos.

Copyright (c) Cătălin Mariș

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
