package entity

import "github.com/PuerkitoBio/goquery"

// NovelFetcher export
type NovelFetcher interface {
	Match(url *string) bool
	FetchDoc(url *string) *goquery.Document

	IsSplit(doc *goquery.Document) bool
	Filter(raw *string) *string
	FetchNovelInfo(url *string) *Novel
	FetchChapter(novelID *Novel, index int) *string
}
