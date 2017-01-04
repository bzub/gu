// +build ignore

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func pullDoc(url string, fx func(doc *goquery.Document)) error {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return err
	}

	fx(doc)
	return nil
}

func main() {

	doc, err := goquery.NewDocument("http://www.iana.org/assignments/media-types/media-types.xhtml")
	if err != nil {
		log.Fatalf("Unable to pull Mime types: %s", err)
	}

	var types bytes.Buffer

	tableApp := doc.Find("#table-application tbody tr")
	tableModel := doc.Find("#table-model tbody tr")
	tableMessage := doc.Find("#table-message tbody tr")
	tableMultipart := doc.Find("#table-multipart tbody tr")
	tableText := doc.Find("#table-text tbody tr")
	tableVideo := doc.Find("#table-video tbody tr")

	pullMimes(&types, tableApp)
	pullMimes(&types, tableMessage)
	pullMimes(&types, tableModel)
	pullMimes(&types, tableMultipart)
	pullMimes(&types, tableText)
	pullMimes(&types, tableVideo)

	file, err := os.Create("mimes.gen.go")
	if err != nil {
		panic(err)
	}

	defer file.Close()

	fmt.Fprintf(file, `// Package mimes contains a infile db of mime types generated
// from different sources including www.iana.org/assignments/media-types/media-types.xhtml
// and crafted by hand from http://hg.nginx.org/nginx/raw-file/default/conf/mime.types
// to provide the best comprehensive list of mime types when the mime native go
// package fails.
// Updates to the file can be generated by running: 'go generate'

//go:generate go run generate.go

package mimes

import (
	"mime"
	"sync"
	"strings"
)

var extDB  = struct{
  db map[string]Extension
  dbl sync.RWMutex
}{
  db: make(map[string]Extension),
}

// Extension defines a struct detailing extension information for a giving mime
// type
type Extension struct{
  Name string
  Ext string
  Reference []string
}

// GetByExtensionName returns the mime type by using the extension of the file or
// by matching the extension part without the '.' with the infile library
func GetByExtensionName(ext string) string {
  if mtype := mime.TypeByExtension(ext); mtype != "" {
    return mtype
  }

  extn, has := GetByExtension(ext)
  if !has{
    return ""
  }

  return extn.Name
}

// GetByExtension returns the Extension associated with the giving extension
// either be it a '.html' or 'html' formatted extension name.
func GetByExtension(ext string) (Extension, bool) {
  extnd := strings.TrimPrefix(ext,".")
  extnd2 := strings.ToLower(extnd)
  extnd3 := strings.ToUpper(extnd)

  var end Extension
  var has bool

  extDB.dbl.RLock()
  {
    end, has = extDB.db[extnd]
    if !has{
      end, has = extDB.db[extnd2]

	    if !has{
	      end, has = extDB.db[extnd3]
	    }
    }
  }
  extDB.dbl.RUnlock()

  return end,has
}

// AddExtensionType adds a new extension into the extension record.
func AddExtensionType(ext string, typed string, references ...string){
  extnd := strings.TrimPrefix(ext,".")
  extnd2 := strings.ToLower(extnd)
  extnd3 := strings.ToUpper(extnd)

  var has bool
  var has2 bool
  var has3 bool

  extDB.dbl.RLock()
  {
    _, has = extDB.db[extnd]
    _, has2 = extDB.db[extnd2]
    _, has3 = extDB.db[extnd3]
  }
  extDB.dbl.RUnlock()

  if has || has2 || has3 {
    return
  }

  extDB.dbl.Lock()
  {
    extDB.db[extnd] = Extension{
      Ext: ext,
      Name: typed,
      Reference: references,
    }
  }
  extDB.dbl.Unlock()

}

func init() {
  AddExtensionType("go","application/go+source","http://golang.org")

%+s
}

`, types.Bytes())
}

func pullMimes(w io.Writer, sel *goquery.Selection) {
	sel.Each(func(_ int, s *goquery.Selection) {
		tds := s.Find("td")

		style, ok := tds.Attr("style")
		if ok && strings.Contains(style, "cursor:") {
			return
		}

		extNode := tds.WrapNode(tds.Get(0))
		html, _ := extNode.Html()

		ext := strings.Split(extNode.Text(), "\n")
		if len(ext) < 2 {
			return
		}

		var refs []string

		tds.WrapNode(tds.Get(2)).Find("a").Each(func(_ int, sel *goquery.Selection) {
			href, ok := sel.Attr("href")
			if !ok {
				return
			}

			if !strings.HasPrefix(href, "http") {
				return
			}

			refs = append(refs, fmt.Sprintf("%q", href))
		})

		if len(refs) > 0 {
			fmt.Fprintf(w, " AddExtensionType(%q, %q, %s)\n", html, strings.TrimSpace(ext[1]), strings.Join(refs, ","))
			return
		}

		fmt.Fprintf(w, " AddExtensionType(%q, %q)\n", html, strings.TrimSpace(ext[1]))
	})
}
