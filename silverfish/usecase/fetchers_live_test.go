//go:build live

// Liveness suite: drives every registered fetcher through the full
// CrawlNovel / CrawlComic pipeline against the real upstream site, so we
// catch both upstream outages and selector rot.
//
// Gated by `-tags=live` so normal `go test ./...` stays offline-safe.
// Rod-based fetchers (happymh, mfhmh, ikanwzd, jmd8) need Chromium at
// /usr/bin/chromium — see fetcher_base.GenerateRodBrowser.
//
// Each case has a default sample URL plus a per-fetcher env var override
// (SILVERFISH_TEST_URL_<NAME>). Set the env var to "SKIP" to bypass a
// single fetcher (e.g. when an upstream is down or behind a paywall).
package usecase_test

import (
	"os"
	"strings"
	"testing"

	interf "silverfish/silverfish/interface"
	usecase "silverfish/silverfish/usecase"
)

type novelCase struct {
	name    string
	envKey  string
	fetcher interf.INovelFetcher
	defURL  string
}

type comicCase struct {
	name    string
	envKey  string
	fetcher interf.IComicFetcher
	defURL  string
}

var novelCases = []novelCase{
	{"77xsw", "SILVERFISH_TEST_URL_77XSW",
		usecase.NewFetcher77xsw("www.77xsw.la"),
		"https://www.77xsw.la/book/8402/"},
	{"hjwzw", "SILVERFISH_TEST_URL_HJWZW",
		usecase.NewFetcherHjwzw("tw.hjwzw.com"),
		"https://tw.hjwzw.com/Book/1644/"},
	{"biquge", "SILVERFISH_TEST_URL_BIQUGE",
		usecase.NewFetcherBiquge("www.biquge.com.cn"),
		"https://www.biquge.com.cn/book/14422/"},
	{"aixdzs", "SILVERFISH_TEST_URL_AIXDZS",
		usecase.NewFetcherAixdzs("tw.aixdzs.com"),
		"https://tw.aixdzs.com/d/91/91851/"},
	{"bookbl", "SILVERFISH_TEST_URL_BOOKBL",
		usecase.NewFetcherBookbl("www.bookbl.com"),
		"https://www.bookbl.com/novel/VlJeXgIE.html"},
}

var comicCases = []comicCase{
	{"nokiacn", "SILVERFISH_TEST_URL_NOKIACN",
		usecase.NewFetcherNokiacn("www.nokiacn.net"),
		"http://www.nokiacn.net/comiclist/1.html"},
	{"cartoonmad", "SILVERFISH_TEST_URL_CARTOONMAD",
		usecase.NewFetcherCartoonmad("www.cartoonmad.com"),
		"https://www.cartoonmad.com/comic/1.html"},
	{"comicbus", "SILVERFISH_TEST_URL_COMICBUS",
		usecase.NewFetcherComicbus("comicbus.com"),
		"https://comicbus.com/comic/1/"},
	{"manhuaniu", "SILVERFISH_TEST_URL_MANHUANIU",
		usecase.NewFetcherManhuaniu("www.manhuaniu.com"),
		"https://www.manhuaniu.com/comic/1/"},
	{"mangabz", "SILVERFISH_TEST_URL_MANGABZ",
		usecase.NewFetcherMangabz("www.mangabz.com"),
		"http://www.mangabz.com/manhua/1/"},
	{"happymh", "SILVERFISH_TEST_URL_HAPPYMH",
		usecase.NewFetcherHappymh("m.happymh.com"),
		"https://m.happymh.com/comic/1/"},
	{"baozimh", "SILVERFISH_TEST_URL_BAOZIMH",
		usecase.NewFetcherBaozimh("www.baozimh.com"),
		"http://www.baozimh.com/comic/1/"},
	{"mfhmh", "SILVERFISH_TEST_URL_MFHMH",
		usecase.NewFetcherMfhmh("www.mfhmh.com"),
		"https://www.mfhmh.com/comic/1/"},
	{"ikanwzd", "SILVERFISH_TEST_URL_IKANWZD",
		usecase.NewFetcherIkanwzd("www.ikanwzd.top"),
		"https://www.ikanwzd.top/comics/1/"},
	{"jmd8", "SILVERFISH_TEST_URL_JMD8",
		usecase.NewFetcherJmd8("jmd8.com"),
		"https://jmd8.com/1/"},
}

func resolveURL(envKey, def string) string {
	if v := strings.TrimSpace(os.Getenv(envKey)); v != "" {
		return v
	}
	return def
}

// safeCrawl converts a panic from a fetcher into a test failure for the
// current subtest, so one fetcher blowing up doesn't abort the whole suite.
// Several fetchers index into slices without bounds checks when the upstream
// HTML structure changes — that's the kind of regression we want to see, but
// per-subtest, not as a test-binary-wide panic.
func safeCrawl(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic during crawl: %v", r)
		}
	}()
	fn()
}

func TestNovelFetchersLive(t *testing.T) {
	for _, tc := range novelCases {
		t.Run(tc.name, func(t *testing.T) {
			url := resolveURL(tc.envKey, tc.defURL)
			if strings.EqualFold(url, "SKIP") {
				t.Skipf("skipped via %s=SKIP", tc.envKey)
			}
			if !tc.fetcher.Match(&url) {
				t.Fatalf("Match() returned false for %s — host mismatch", url)
			}
			safeCrawl(t, func() {
				novel, err := tc.fetcher.CrawlNovel(&url)
				if err != nil {
					t.Fatalf("CrawlNovel(%s): %v", url, err)
				}
				if novel == nil {
					t.Fatalf("CrawlNovel(%s) returned nil", url)
				}
				if strings.TrimSpace(novel.Title) == "" {
					t.Errorf("empty Title for %s", url)
				}
				if len(novel.Chapters) == 0 {
					t.Errorf("empty Chapters for %s", url)
				}
				t.Logf("OK: %q (%d chapters) via %s", novel.Title, len(novel.Chapters), url)
			})
		})
	}
}

func TestComicFetchersLive(t *testing.T) {
	for _, tc := range comicCases {
		t.Run(tc.name, func(t *testing.T) {
			url := resolveURL(tc.envKey, tc.defURL)
			if strings.EqualFold(url, "SKIP") {
				t.Skipf("skipped via %s=SKIP", tc.envKey)
			}
			if !tc.fetcher.Match(&url) {
				t.Fatalf("Match() returned false for %s — host mismatch", url)
			}
			safeCrawl(t, func() {
				comic, err := tc.fetcher.CrawlComic(&url)
				if err != nil {
					t.Fatalf("CrawlComic(%s): %v", url, err)
				}
				if comic == nil {
					t.Fatalf("CrawlComic(%s) returned nil", url)
				}
				if strings.TrimSpace(comic.Title) == "" {
					t.Errorf("empty Title for %s", url)
				}
				if len(comic.Chapters) == 0 {
					t.Errorf("empty Chapters for %s", url)
				}
				t.Logf("OK: %q (%d chapters) via %s", comic.Title, len(comic.Chapters), url)
			})
		})
	}
}
