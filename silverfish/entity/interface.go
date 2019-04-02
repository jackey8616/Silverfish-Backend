package entity

import "github.com/PuerkitoBio/goquery"

// NovelFetcher export
type NovelFetcher interface {
	Match(url *string) bool
	FetchNovelInfo(url *string) *Novel
	FetchChapter(id *string) *string
	FetchDoc(url *string) *goquery.Document
}
