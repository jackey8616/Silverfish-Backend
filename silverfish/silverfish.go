package silverfish

import (
	"silverfish/silverfish/entity"
	interf "silverfish/silverfish/interface"
	usecase "silverfish/silverfish/usecase"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
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
	session *dynamodb.Client,
) *Silverfish {
	sf := new(Silverfish)
	novelFetchers := map[string]interf.INovelFetcher{
		"www.77xsw.la":      usecase.NewFetcher77xsw("www.77xsw.la"),
		"tw.hjwzw.com":      usecase.NewFetcherHjwzw("tw.hjwzw.com"),
		"www.biquge.com.cn": usecase.NewFetcherBiquge("www.biquge.com.cn"),
		"tw.aixdzs.com":     usecase.NewFetcherAixdzs("tw.aixdzs.com"),
		"www.bookbl.com":    usecase.NewFetcherBookbl("www.bookbl.com"),
	}
	comicFetchers := map[string]interf.IComicFetcher{
		//"www.99comic.co":     usecase.NewFetcher99Comic("www.99comic.co"),
		"www.nokiacn.net":    usecase.NewFetcherNokiacn("www.nokiacn.net"),
		"www.cartoonmad.com": usecase.NewFetcherCartoonmad("www.cartoonmad.com"),
		"comicbus.com":       usecase.NewFetcherComicbus("comicbus.com"),
		"www.manhuaniu.com":  usecase.NewFetcherManhuaniu("www.manhuaniu.com"),
		"www.mangabz.com":    usecase.NewFetcherMangabz("www.mangabz.com"),
		"m.happymh.com":      usecase.NewFetcherHappymh("m.happymh.com"),
		"www.baozimh.com":    usecase.NewFetcherBaozimh("www.baozimh.com"),
		"www.mfhmh.com":      usecase.NewFetcherMfhmh("www.mfhmh.com"),     // Oops...
		"www.ikanwzd.top":    usecase.NewFetcherIkanwzd("www.ikanwzd.top"), // Oops...
	}

	userInf := entity.NewDynamoInf(session, "Silverfish_Users")
	novelInf := entity.NewDynamoInf(session, "Silverfish_Novels")
	comicInf := entity.NewDynamoInf(session, "Silverfish_Comics")
	sf.Auth = NewAuth(hashSalt, userInf)
	sf.Novel = NewNovel(sf.Auth, novelInf, novelFetchers, crawlDuration)
	sf.Comic = NewComic(sf.Auth, comicInf, comicFetchers, crawlDuration)
	sf.Admin = NewAdmin(userInf)
	sf.User = NewUser(userInf)
	return sf
}
