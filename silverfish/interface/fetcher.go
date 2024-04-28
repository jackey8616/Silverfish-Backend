package interf

import (
	"net/http"
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
	CrawlNovel(url *string) (*entity.Novel, error)
	FetchNovelInfo(novelId *string, doc *goquery.Document) (*entity.NovelInfo, error)
	FetchChapterInfo(doc *goquery.Document, title, url string) []entity.NovelChapter
	UpdateNovelInfo(novel *entity.Novel) (*entity.Novel, error)
	FetchNovelChapter(novel *entity.Novel, index int) (*string, error)
}

// IComicFetcher export
type IComicFetcher interface {
	Match(url *string) bool
	FetchDoc(url *string) (*goquery.Document, error)

	GetChapterURL(comic *entity.Comic, url string) *string
	CrawlComic(url *string) (*entity.Comic, error)
	FetchComicInfo(comicId *string, doc *goquery.Document, cookies []*http.Cookie) (*entity.ComicInfo, error)
	FetchChapterInfo(doc *goquery.Document, cookies []*http.Cookie, title, url string) []entity.ComicChapter
	UpdateComicInfo(comic *entity.Comic) (*entity.Comic, error)
	FetchComicChapter(comic *entity.Comic, index int) ([]string, error)
}
