package interf

import (
	entity "silverfish/silverfish/entity"

	"github.com/PuerkitoBio/goquery"
)

// INovelFetcher export
type INovelFetcher interface {
	Match(url *string) bool
	FetchDoc(url *string) (*goquery.Document, error)

	IsSplit(doc *goquery.Document) bool
	Filter(raw *string) *string

	GetChapterURL(novel *entity.Novel, index int) *string
	FetchNovelInfo(url *string) (*entity.Novel, error)
	UpdateNovelInfo(novel *entity.Novel) (*entity.Novel, error)
	FetchNovelChapter(novel *entity.Novel, index int) (*string, error)
}

// IComicFetcher export
type IComicFetcher interface {
	Match(url *string) bool
	FetchDoc(url *string) (*goquery.Document, error)

	GetChapterURL(comic *entity.Comic, url string) *string
	FetchComicInfo(url *string) (*entity.Comic, error)
	UpdateComicInfo(comic *entity.Comic) (*entity.Comic, error)
	FetchComicChapter(comic *entity.Comic, index int) ([]string, error)
}
