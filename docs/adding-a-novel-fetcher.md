# Adding a new novel-source fetcher

This is the playbook for wiring a new Chinese novel site into the
crawler. The same recipe applies to comic sites — swap `INovelFetcher`
for `IComicFetcher` and the corresponding entity types. The fetcher
contracts live in `silverfish/interface/fetcher.go`.

The example URL used throughout is `https://uukanshu.cc/book/25264/`,
which became `silverfish/usecase/fetcher_uukanshu.go`. That fetcher was
subsequently dropped (the Alpine `chromium` shipped in the prod image
fails Cloudflare's JS challenge, even though full Google Chrome locally
passes it — see `project-dead-fetchers` memory) so the file is no longer
in the tree; consult git history for the original. The process below
still applies unchanged.

## 1. Probe the target

The first goal is to see the real HTML the server returns so you can
pick stable selectors. Two paths exist; **always start with plain HTTP**
and only escalate to Rod if you must.

### 1a. Plain HTTP

```bash
curl -sL -A "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) \
  AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36" \
  -H "Accept-Language: zh-CN,zh;q=0.9" \
  "<info-page-url>" -o /tmp/probe.html
```

Look for:

- `<title>Just a moment...</title>` → Cloudflare interstitial, jump to 1b.
- Empty or near-empty body → JS-rendered, jump to 1b.
- Mojibake → non-UTF-8 charset; note it (`gbk`, `gb18030`, `big5`) and
  use `FetchDocWithEncoding` instead of `FetchDoc`.
- `og:novel:*` meta tags → you can lift Title/Author/Description/Cover
  from meta directly (see `fetcher_hjwzw.go`).

### 1b. Rod (headless Chromium)

When plain HTTP fails, drive the page through Chromium. The repo's
helper is `Fetcher.FetchDocViaRod` (`silverfish/usecase/fetcher_base.go`).
For one-off probing outside the repo, a minimal standalone script:

```go
// /tmp/probe/main.go
package main
import (
  "os"; "time"
  "github.com/go-rod/rod"
  "github.com/go-rod/rod/lib/launcher"
)
func main() {
  l := launcher.New().
    Bin("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome").
    Headless(true).
    Set("disable-gpu").Set("no-sandbox").
    Set("ignore-certificate-errors")
  b := rod.New().ControlURL(l.MustLaunch()).MustConnect()
  defer b.MustClose()
  p := b.MustPage(os.Args[1]).MustWaitLoad()
  time.Sleep(8 * time.Second) // let Cloudflare clear
  p.MustWaitLoad()
  html := p.MustElement("html").MustHTML()
  os.WriteFile(os.Args[2], []byte(html), 0644)
}
```

```bash
cd /tmp/probe && go mod init probe && go mod tidy
go run . "<url>" /tmp/probe.html
```

Note Rod's API surface changed between versions — the in-repo code
pins `go-rod/rod v0.88.2`. `launcher.NewBrowser().LookPath()` only
exists on newer Rod, so the standalone script hard-codes the Chrome
binary path. The repo's `chromiumBin()` handles this correctly for the
production code path; you only need the workaround for ad-hoc probes.

### 1c. What to capture from the probe

Both pages must work:

1. **Info page** (`<url>` — the book landing page). Identify:
   - Title, author, description, cover image (prefer `og:` meta).
   - Chapter list container + per-chapter link shape.

2. **One chapter page** (any chapter URL from the list). Identify:
   - The element that holds the body text.
   - Any ad/script/footer noise to strip in `Filter()`.
   - Whether the chapter is split across multiple pages — if so,
     `IsSplit` returns true and you'll need pagination handling
     (see `fetcher_hjwzw.go` is a non-split reference; no current
     fetcher uses split — handle inline in `FetchNovelChapter`).

Useful one-liners against the captured HTML:

```bash
# verify Cloudflare cleared
grep -c "Just a moment" /tmp/probe.html

# list og: meta
python3 -c "import re; [print(m[:200]) for m in
  re.findall(r'<meta[^>]+>', open('/tmp/probe.html').read())
  if 'og:' in m]"

# locate chapter list container near a known chapter href
python3 -c "
h = open('/tmp/probe.html').read()
i = h.find('<known-chapter-href>')
print(h[max(0,i-400):i+200])"
```

## 2. Pick a template fetcher

Three reference fetchers cover most cases:

| Reference                | Use when                                                                  |
| ------------------------ | ------------------------------------------------------------------------- |
| `fetcher_hjwzw.go`       | Plain HTTP works, `og:novel:*` meta tags present, explicit chapter table. |
| `fetcher_aixdzs.go`      | Rod required (JS-rendered or anti-bot), chapter list inferred from page.  |

For a non-UTF-8 upstream (GBK / GB18030 / Big5), use
`Fetcher.FetchDocWithEncoding(url, "gbk")` from `fetcher_base.go` instead
of plain `FetchDoc`. No remaining fetcher currently exercises that path
— grep the git history for a deleted example if needed.

Copy the closest match into a new `fetcher_<site>.go`. The fetcher
package is `silverfish/usecase`. Embed `Fetcher` from `fetcher_base.go`
to inherit `Match`, `FetchDoc`, `FetchDocWithEncoding`, `FetchDocViaRod`,
`GenerateRodBrowser`, `GenerateID`.

## 3. Implement `INovelFetcher`

The contract (`silverfish/interface/fetcher.go`):

```go
Match(url) bool
FetchDoc(url) (*goquery.Document, error)
IsSplit(doc) bool
Filter(raw *string) *string
GetChapterURL(novel, index) *string
CrawlNovel(url) (*entity.Novel, error)
FetchNovelInfo(novelID, doc) (*entity.NovelInfo, error)
FetchChapterInfo(doc, title, url) []entity.NovelChapter
UpdateNovelInfo(novel) (*entity.Novel, error)
FetchNovelChapter(novel, index) (*string, error)
```

Constructor convention: `NewFetcher<Site>(dns string) *Fetcher<Site>`.
The `dns` argument is the host the fetcher claims; it's matched against
the URL in `Match`. Pass `true` to `NewFetcher` for HTTPS sites.

### Required behaviors

- **`CrawlNovel`** — fetch the info page, derive ID via `GenerateID(url)`,
  call `FetchNovelInfo` and `FetchChapterInfo`, assemble `entity.Novel`,
  `SetNovelInfo`. Return error on missing required fields, not partial
  data.
- **`FetchNovelInfo`** — return `nil, error` if **any** of title /
  author / description / cover URL is missing. The crawler treats this
  as a hard failure. `LastCrawlTime: time.Now()`.
- **`FetchChapterInfo`** — return the full chapter list in **read order**
  (chapter 1 first). The stored chapter URL is what `GetChapterURL`
  will turn into a fetchable URL — usually a relative path that
  `GetChapterURL` prefixes with `https://<dns>`.
- **`FetchNovelChapter`** — fetch one chapter page, extract body HTML,
  pass through `Filter`. Return the inner HTML (with `<br/>`/`<p>` —
  the frontend renders HTML directly).
- **`UpdateNovelInfo`** — same as `CrawlNovel` but mutates the existing
  `*entity.Novel` so `NovelID` and other persisted IDs survive. This is
  called by `Novel.GetNovelByID` when `LastCrawlTime` is older than
  `CrawlDuration` minutes.
- **`Filter`** — strip ad scripts, injected promos, repeated boilerplate.
  Compile regexes at package level, not per-call.
- **`IsSplit`** — `return false` unless the site paginates a single
  chapter across multiple pages.

### Style notes lifted from existing fetchers

- Compile regexes once as package vars; comment **why** each one
  exists. See `aixdzsChapterRe` for the shape.
- Trim whitespace on extracted text (titles especially have
  leading/trailing spaces).
- If you log instead of erroring, use `logrus.Print` / `logrus.Printf`
  — match the existing tone.
- No comments explaining what the code does — names already do that.
  Comments earn their keep by explaining hidden constraints (e.g.
  "info page also carries a 'latest chapter' promo link with the same
  href shape" in `fetcher_uukanshu.go`).

## 4. Register the fetcher

`silverfish/silverfish.go` — add to `novelFetchers`:

```go
novelFetchers := map[string]interf.INovelFetcher{
  ...
  "<host>": usecase.NewFetcher<Site>("<host>"),
}
```

The map key must match the host portion of incoming URLs exactly — it's
how `Match` picks the fetcher.

## 5. Add a live-test case

`silverfish/usecase/fetchers_live_test.go` — append to `novelCases`:

```go
{"<site>", "SILVERFISH_TEST_URL_<SITE>",
  usecase.NewFetcher<Site>("<host>"),
  "<known-good-info-url>"},
```

The default URL is what CI hits when `SILVERFISH_TEST_URL_<SITE>` isn't
set. Pick a URL that's known to be stable in production — typically
one already in `silverfish.novel` in the prod MongoDB. Set the env var
to `SKIP` to bypass the case when upstream is down.

## 6. Verify

```bash
# compile
go build ./...
go vet ./...

# live test for just the new fetcher
go test -tags=live -run 'TestNovelFetchersLive/<site>' \
  -v -timeout 300s ./silverfish/usecase/
```

A green liveness run only exercises `CrawlNovel`. **Also verify chapter
body extraction** — the suite doesn't cover it. Drop a temporary
`*_test.go` next to the fetcher that calls `FetchNovelChapter(novel, 0)`
and checks:

- non-empty body,
- no residual `<script` tags (ad noise filtered),
- a known string from the first chapter is present.

Delete the temp test once it passes; the standing live suite is the
durable coverage.

If Rod is involved, the local dev box needs Chrome/Chromium installed.
`chromiumBin()` (`fetcher_base.go`) auto-discovers: `ROD_BIN` env →
Rod's per-OS LookPath → `/usr/bin/chromium` (the Alpine path the prod
Dockerfile installs).

## 7. Commit

Follow the existing commit prefix style — see `git log --oneline`:

```
feat(<site>): add Rod-based novel fetcher for <host>
fix(<site>): switch to Rod, rewrite selectors, ...
chore: drop dead fetchers ...
```

Body: one short paragraph on what the site is doing (Cloudflare,
JS-rendered, encoded) and which selectors carry the load. That's the
context future-you needs when selectors rot.

## Common gotchas

- **Cloudflare "Just a moment"** — Rod-only. Plain HTTP returns the
  interstitial; the body has no useful data even after the challenge.
- **Two hrefs per chapter** — info pages often promote the "latest
  chapter" at the top using the same href shape as the list. Scope
  the selector to the actual list container (`dd > a`, `tbody > tr > a`,
  etc.) or filter with a regex on the URL pattern.
- **Chapter numbering gaps** — the dd count and the visible "chapter N"
  may not match (sites delete or merge chapters but keep IDs).
  Enumerate what's on the page, don't derive from the highest number.
- **Encoding** — `FetchDocWithEncoding` exists for `gbk`/`gb18030`/`big5`.
  Symptom is mojibake on `doc.Find(...).Text()` even though the raw
  HTML looks fine in a browser.
- **`go-rod` version mismatch** — repo pins `v0.88.2`. Don't copy
  examples from upstream docs without checking the API still exists at
  that version.
- **Dead sites** — see `memory/project_dead_fetchers.md`. Some sites
  are unfixable at the source (CAPTCHAs, account walls, dead domains).
  Don't burn time patching selectors for those — open an issue and
  move on, or follow the `chore: drop N dead fetchers` precedent.

## File checklist for the PR

- [ ] `silverfish/usecase/fetcher_<site>.go` — new
- [ ] `silverfish/silverfish.go` — entry added to `novelFetchers` (or
      `comicFetchers`)
- [ ] `silverfish/usecase/fetchers_live_test.go` — case added with
      stable default URL
- [ ] `go build ./...` clean
- [ ] `go vet ./...` clean
- [ ] `go test -tags=live -run 'TestNovelFetchersLive/<site>' …` passes
- [ ] Chapter body extraction manually verified
