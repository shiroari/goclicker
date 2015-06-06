package bro

import (
	"log"
	"time"
 	"net/http"
 	"net/http/httputil"
	"net/http/cookiejar"	
	"container/list"

	"golang.org/x/net/html"
	//"code.google.com/p/go-html-transform/css/selector"
)

/*  Client */

type Client struct {
	client http.Client
	logLevel int
}

func New(logLevel int) *Client{

	jar, err := cookiejar.New(nil)

	client := http.Client{}

	if err != nil {
		log.Fatal(err)
	}

    client.Jar = jar

    return &Client{client, logLevel}
}

func (self *Client) GetUrls(uri string) ([]string, error) {

	timer := time.Now()

	request, err := http.NewRequest("GET", uri, nil)
	request.SetBasicAuth("system", "manager")

	if self.logLevel > 0 {
		log.Printf("-> GET %s", uri)	
		if self.logLevel > 1 {
			dump, err := httputil.DumpRequestOut(request, (self.logLevel > 2))
			if err == nil {
				log.Printf("-> %s", dump)
			}
		}		
	}

	resp, err := self.client.Do(request)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if self.logLevel > 0 {
		log.Printf("<- %s, Time: %s, Length: %d :: %s", resp.Status, time.Since(timer), resp.ContentLength, uri)
		if self.logLevel > 1 {
			dump, err := httputil.DumpResponse(resp, (self.logLevel > 2))
			if err == nil {
				log.Printf("<- %s", dump)
			}		
		}		
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Panic(err)
		return nil, err
	}

	links := getElementsByTag(doc, "a")	
	res := make([]string, links.Len())

	i := 0
	for e := links.Front(); e != nil; e = e.Next() {
		res[i] = getAttribute(e.Value.(*html.Node), "href")
		i++
	}

	return res, nil
}

/*  Helpers */

type filterFunc func(n *html.Node) bool
type actionFunc func(n *html.Node)

func visit(n *html.Node, filter filterFunc, action actionFunc) {
	if filter(n) {
		action(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		visit(c, filter, action)
	}
}

func getElementsByTag(node *html.Node, tag string) *list.List {

	selected := list.New()

	visit(node,
		func(n *html.Node) bool {
			if n.Type == html.ElementNode {
				if n.Data == tag {
					return true
				}
			}
			return false
		},
		func(n *html.Node) {
			selected.PushBack(n)
		})

	return selected
}

func getElementsByClass(node *html.Node, name string) *list.List {

	selected := list.New()

	visit(node,
		func(n *html.Node) bool {
			if name == getAttribute(n, "class") {
				return true
			}
			return false
		},
		func(n *html.Node) {
			selected.PushBack(n)
		})

	return selected
}

func getText(node *html.Node) string {

	res := ""

	visit(node,
		func(n *html.Node) bool {
			return n.Type == html.TextNode
		},
		func(n *html.Node) {
			res += n.Data
		})

	return res
}

func getAttribute(node *html.Node, name string) string {
	if node.Type != html.ElementNode {
		return ""
	}
	for _, attr := range node.Attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}
