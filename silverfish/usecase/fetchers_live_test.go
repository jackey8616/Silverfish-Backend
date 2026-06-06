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

// Default sample URLs were taken from the production MongoDB
// (silverfish.novel/silverfish.comic) — one historically-valid URL per
// fetcher — so they reflect the real path shape the fetcher was designed
// for, not a guessed `/<id>/` template.
var novelCases = []novelCase{
	{"hjwzw", "SILVERFISH_TEST_URL_HJWZW",
		usecase.NewFetcherHjwzw("tw.hjwzw.com"),
		"https://tw.hjwzw.com/Book/1644/"},
	{"aixdzs", "SILVERFISH_TEST_URL_AIXDZS",
		usecase.NewFetcherAixdzs("tw.aixdzs.com"),
		"https://tw.aixdzs.com/d/8/8282/"},
	{"ttkan", "SILVERFISH_TEST_URL_TTKAN",
		usecase.NewFetcherTtkan("www.ttkan.co"),
		"https://www.ttkan.co/novel/chapters/quanzhifashi-luan"},
}

var comicCases = []comicCase{
	{"mangabz", "SILVERFISH_TEST_URL_MANGABZ",
		usecase.NewFetcherMangabz("www.mangabz.com"),
		"http://www.mangabz.com/15261bz/"},
	{"baozimh", "SILVERFISH_TEST_URL_BAOZIMH",
		usecase.NewFetcherBaozimh("www.baozimh.com"),
		"https://www.baozimh.com/comic/woduzishengji-duburedicestudio_gi486f"},
	{"jmd8", "SILVERFISH_TEST_URL_JMD8",
		usecase.NewFetcherJmd8("jmd8.com"),
		"https://jmd8.com/manga/52366"},
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
