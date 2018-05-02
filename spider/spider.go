package spider

import (
	"fmt"
	"strconv"
	"sync"
	"unicode/utf8"

	"strings"

	"github.com/crackcomm/crawl"
	"golang.org/x/net/context"
)

// List - fin company list
var List = "fin_list"

// Outlist slice
var Outlist [][]string

// CheckMap map
var CheckMap map[string]bool

// WG - Waitgroup for concurrency
var WG *sync.WaitGroup
var count int

// Spider - Registers spider.
func Spider(c crawl.Crawler) {
	spider := &fSpider{Crawler: c}
	c.Register(List, spider.List)
}

type fSpider struct {
	crawl.Crawler
}

func (spider *fSpider) List(ctx context.Context, resp *crawl.Response) (err error) {
	if err := spider.checkError(resp); err != nil {
		WG.Done()
		return err
	}
	var bodystring string
	var author, cover, isbn, publisher, price string
	var n int
	var pricef float64
	defer WG.Done()

	err = resp.ParseHTML()
	bytes, err := resp.Bytes()
	bodystring = string(bytes)

	n = strings.Index(bodystring, "by <strong><span itemprop=\"author\">")
	if n != -1 {
		n = n + utf8.RuneCountInString("by <strong><span itemprop=\"author\">")
		ls := strings.SplitAfter(bodystring[n:], "</span></strong>\n")
		author = ls[0][0 : len(ls[0])-utf8.RuneCountInString("</span></strong>\n")]
	}

	cover = ""
	n = strings.Index(bodystring, "<span class=\"describe-isbn\"><link itemprop=\"bookformat\" href=\"http://schema.org/Paperback\">")
	if n != -1 {
		n = n + utf8.RuneCountInString("<span class=\"describe-isbn\"><link itemprop=\"bookformat\" href=\"http://schema.org/Paperback\">")
		ls := strings.SplitAfter(bodystring[n:], "\n")
		cover = ls[0][0 : len(ls[0])-utf8.RuneCountInString("</link></span>\n")-1]
	}

	n = strings.Index(bodystring, "<span class=\"describe-isbn\"><link itemprop=\"bookformat\" href=\"http://schema.org/Hardcover\">")
	if n != -1 {
		n = n + utf8.RuneCountInString("<span class=\"describe-isbn\"><link itemprop=\"bookformat\" href=\"http://schema.org/Hardcover\">")
		ls := strings.SplitAfter(bodystring[n:], "\n")
		cover = ls[0][0 : len(ls[0])-utf8.RuneCountInString("</link></span>\n")-1]
	}

	publisher = ""
	n = strings.Index(bodystring, "<span class=\"describe-isbn-h\">Publisher:</span><span itemprop=\"publisher\" class=\"describe-isbn\">")
	if n != -1 {
		n = n + utf8.RuneCountInString("<span class=\"describe-isbn-h\">Publisher:</span><span itemprop=\"publisher\" class=\"describe-isbn\">")
		ls := strings.SplitAfter(bodystring[n:], "\n")
		publisher = ls[0][0 : len(ls[0])-utf8.RuneCountInString("</span>\n")-1]
	}

	price = ""
	n1 := strings.Index(bodystring, "<h3 class=\"results-section-heading\" style=\"color: black\">Used books:")
	if n1 != -1 {
		n = strings.Index(bodystring[n1:], "<tr valign=\"top\" class=\"results-table-first-LogoRow has-data\" data-price=\"")
		if n != -1 {
			n = n + utf8.RuneCountInString("<tr valign=\"top\" class=\"results-table-first-LogoRow has-data\" data-price=\"")
			ls := strings.SplitAfter(bodystring[n+n1:], "\"")
			price = ls[0][0 : len(ls[0])-utf8.RuneCountInString("\"")]

			if pricef, err = strconv.ParseFloat(price, 64); err == nil {
				price = fmt.Sprintf("%.2f", pricef)
			}
		}
	}

	n = strings.Index(bodystring, "<a href=\"//www.bookfinder.com/search/?keywords=")
	if n != -1 {
		n = n + utf8.RuneCountInString("<a href=\"//www.bookfinder.com/search/?keywords=")
		isbn = bodystring[n : n+13]
	}
	if isbn != "" {
		Outlist = append(Outlist, []string{isbn, author, cover, publisher, price})
		CheckMap[isbn] = true
	}

	spider.Close()
	count++
	fmt.Printf("%d.", count)
	return
}

func (spider *fSpider) checkError(resp *crawl.Response) (err error) {
	if crawl.Text(resp, "h1") == "D'oh!" {
		return fmt.Errorf("Error: %q", crawl.Text(resp, "body"))
	}
	return
}
