package usecase

import (
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"

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
func (f *Fetcher) FetchDoc(url *string) (*goquery.Document, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
	}

	res, err := client.Get(*url)
	if err != nil {
		return nil, errors.Wrap(err, "When FetchDoc client.Get")
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "When FetchDoc")
	}
	return doc, nil
}

// FetchDocWithEncoding export
func (f *Fetcher) FetchDocWithEncoding(url *string, charset string) (*goquery.Document, []*http.Cookie, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
	}

	res, err := client.Get(*url)
	if err != nil {
		return nil, nil, errors.Wrap(err, "When FetchDocWithEncoding client.Get")
	}
	defer res.Body.Close()

	// Convert the designated charset HTML to utf-8 encoded HTML.
	// `charset` being one of the charsets known by the iconv package.
	var enc encoding.Encoding
	switch strings.ToLower(charset) {
	case "gbk", "gb2312":
		enc = simplifiedchinese.GBK
	case "gb18030":
		enc = simplifiedchinese.GB18030
	case "big5":
		enc = traditionalchinese.Big5
	default:
		return nil, nil, fmt.Errorf("unsupported charset: %s", charset)
	}
	utfBody := transform.NewReader(res.Body, enc.NewDecoder())

	// use utfBody using goquery
	doc, err := goquery.NewDocumentFromReader(utfBody)
	if err != nil {
		return nil, nil, errors.Wrap(err, "When FetchDocWithEncoding")
	}
	return doc, res.Cookies(), nil
}

// GenerateID export
func (f *Fetcher) GenerateID(url *string) *string {
	hash := md5.Sum([]byte(*url))
	md5str2 := fmt.Sprintf("%x", hash)
	id, _ := baseconv.Convert(md5str2, baseconv.DigitsHex, baseconv.Digits62)
	id = id[:7]
	return &id
}
