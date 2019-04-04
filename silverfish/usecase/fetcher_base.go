package usecase

import (
	"crypto/md5"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"

	"github.com/kenshaw/baseconv"
)

// Fetcher export
type Fetcher struct {
	tls    bool
	dns    *string
	dnsReg *regexp.Regexp
}

// NewFetcher export
func (f *Fetcher) NewFetcher(tls bool, dns *string) {
	f.dns = dns
	f.tls = tls
	f.dnsReg = regexp.MustCompile("https?://.*?/")
}

// Match export
func (f *Fetcher) Match(url *string) bool {
	getDNS := f.dnsReg.FindAllString(*url, 1)[0]
	if strings.Contains(getDNS, "https://") {
		getDNS = strings.Replace(getDNS, "https://", "", 1)
	} else {
		getDNS = strings.Replace(getDNS, "http://", "", 1)
	}
	return getDNS[:len(getDNS)-1] == *f.dns
}

// FetchDoc export
func (f *Fetcher) FetchDoc(url *string) *goquery.Document {
	doc, err := goquery.NewDocument(*url)
	if err != nil {
		log.Fatal(errors.Wrap(err, "When FetchDoc"))
		return nil
	}
	return doc
}

// GenerateID export
func (f *Fetcher) GenerateID(url *string) *string {
	hash := md5.Sum([]byte(*url))
	md5str2 := fmt.Sprintf("%x", hash)
	id, _ := baseconv.Convert(md5str2, baseconv.DigitsHex, baseconv.Digits62)
	id = id[:7]
	return &id
}
