package entity

import "github.com/PuerkitoBio/goquery"

// NovelFetcher export
type NovelFetcher interface {
	Match(url *string) bool
	FetchDoc(url *string) *goquery.Document

	GetChapterURL(novel *Novel, index int) *string
	IsSplit(doc *goquery.Document) bool
	Filter(raw *string) *string
	FetchNovelInfo(url *string) *Novel
	FetchNovelChapter(novel *Novel, index int) *string
}

// ComicFetcher export
type ComicFetcher interface {
	Match(url *string) bool
	FetchDoc(url *string) *goquery.Document

	GetChapterURL(comic *Comic, url string) *string
	FetchComicInfo(url *string) *Comic
	FetchComicChapter(comic *Comic, index int) []string
}
