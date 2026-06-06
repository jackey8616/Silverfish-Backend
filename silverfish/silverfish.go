package silverfish

import (
	entity "silverfish/silverfish/entity"
	interf "silverfish/silverfish/interface"
	usecase "silverfish/silverfish/usecase"
)

// Silverfish export
type Silverfish struct {
	Auth  *Auth
	Admin *Admin
	User  *User
	Novel *Novel
	Comic *Comic
}

// New export
func New(
	hashSalt *string,
	crawlDuration int,
	userInf, novelInf, comicInf, sessionInf *entity.MongoInf,
) *Silverfish {
	sf := new(Silverfish)
	novelFetchers := map[string]interf.INovelFetcher{
		"tw.hjwzw.com":  usecase.NewFetcherHjwzw("tw.hjwzw.com"),
		"tw.aixdzs.com": usecase.NewFetcherAixdzs("tw.aixdzs.com"),
		"www.ttkan.co":  usecase.NewFetcherTtkan("www.ttkan.co"),
	}
	comicFetchers := map[string]interf.IComicFetcher{
		"www.mangabz.com": usecase.NewFetcherMangabz("www.mangabz.com"),
		"www.baozimh.com": usecase.NewFetcherBaozimh("www.baozimh.com"),
		// jmd8.com 301-redirects to 91jmd.com (the new canonical host).
		// Keep the old entry so existing comic records (DNS=jmd8.com)
		// still resolve their fetcher; they pay one redirect per chapter
		// request until re-crawled. New URLs should use 91jmd.com directly.
		"jmd8.com":  usecase.NewFetcherJmd8("jmd8.com"),
		"91jmd.com": usecase.NewFetcherJmd8("91jmd.com"),
	}

	sf.Auth = NewAuth(hashSalt, userInf, sessionInf)
	sf.Novel = NewNovel(sf.Auth, novelInf, novelFetchers, crawlDuration)
	sf.Comic = NewComic(sf.Auth, comicInf, comicFetchers, crawlDuration)
	sf.Admin = NewAdmin(userInf)
	sf.User = NewUser(userInf)
	return sf
}
