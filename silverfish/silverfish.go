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
		"www.77xsw.la":  usecase.NewFetcher77xsw("www.77xsw.la"),
		"tw.hjwzw.com":  usecase.NewFetcherHjwzw("tw.hjwzw.com"),
		"tw.aixdzs.com": usecase.NewFetcherAixdzs("tw.aixdzs.com"),
	}
	comicFetchers := map[string]interf.IComicFetcher{
		"www.nokiacn.net":    usecase.NewFetcherNokiacn("www.nokiacn.net"),
		"www.cartoonmad.com": usecase.NewFetcherCartoonmad("www.cartoonmad.com"),
		"www.mangabz.com":    usecase.NewFetcherMangabz("www.mangabz.com"),
		"www.baozimh.com":    usecase.NewFetcherBaozimh("www.baozimh.com"),
		"jmd8.com":           usecase.NewFetcherJmd8("jmd8.com"), // Oops...
	}

	sf.Auth = NewAuth(hashSalt, userInf, sessionInf)
	sf.Novel = NewNovel(sf.Auth, novelInf, novelFetchers, crawlDuration)
	sf.Comic = NewComic(sf.Auth, comicInf, comicFetchers, crawlDuration)
	sf.Admin = NewAdmin(userInf)
	sf.User = NewUser(userInf)
	return sf
}
