package crawl

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/net/html/charset"

	"github.com/PuerkitoBio/goquery"
)

// Response - Crawl http response.
// It is expected it to be a HTML response but not required.
// It ALWAYS has to be released using Close() method.
type Response struct {
	*Request
	*http.Response
	doc  *goquery.Document
	body []byte
}

// ParseHTML - Reads response body and parses it as HTML.
func (r *Response) ParseHTML() (err error) {
	body, err := r.Bytes()
	if err != nil {
		return
	}
	r.doc, err = goquery.NewDocumentFromReader(bytes.NewBuffer(body))
	return
}

// Bytes - Reads response body and returns byte array.
func (r *Response) Bytes() (body []byte, err error) {
	if r.body == nil {
		err = r.readBody()
	}
	return r.body, err
}

// Status - Gets response status.
func (r *Response) Status() string {
	return r.Response.Status
}

// URL - Gets response request URL.
func (r *Response) URL() *url.URL {
	return r.Response.Request.URL
}

// Query - Returns goquery.Document.
func (r *Response) Query() *goquery.Document {
	return r.doc
}

// Find - Short for: r.Query().Find(selector).
func (r *Response) Find(selector string) *goquery.Selection {
	return r.doc.Find(selector)
}

// Close - Closes response body.
func (r *Response) Close() error {
	// close response body
	// even though it should be closed after a read
	// but to make sure we can just close again
	return r.Response.Body.Close()
}

// readBody - Reads response body to `r.body`.
func (r *Response) readBody() (err error) {
	if r.body != nil {
		return
	}
	defer r.Response.Body.Close()
	//r.body, err = ioutil.ReadAll(r.Response.Body)

	utf8, err := charset.NewReader(r.Response.Body, r.Response.Header.Get("Content-Type"))
	if err != nil {
		fmt.Println("Encoding error:", err)
		return
	}
	r.body, err = ioutil.ReadAll(utf8)
	if err != nil {
		fmt.Println("IO error:", err)
		return
	}

	return
}
