package entity

import "github.com/PuerkitoBio/goquery"

// NovelFetcher export
type NovelFetcher interface {
	Match(url *string) bool
	FetchDoc(url *string) *goquery.Document

	IsSplit(doc *goquery.Document) bool
	FetchNovelInfo(url *string) *Novel
	FetchChapter(id *string) *string
	FetcherNewChapter(novelID *Novel, index int) *string
}
