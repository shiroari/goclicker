package bro

import (
	"container/list"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"time"

	"golang.org/x/net/html"
	//"code.google.com/p/go-html-transform/css/selector"
)

/*  Client */

type DocumentCallbackFunc func(status int, url string, doc *html.Node)

type Client struct {
	client   http.Client
	user     string
	password string
	logLevel int
}

func New(user, password string, logLevel int) *Client {

	jar, err := cookiejar.New(nil)

	client := http.Client{}

	if err != nil {
		log.Fatal(err)
	}

	client.Jar = jar

	return &Client{client, user, password, logLevel}
}

func (self *Client) RequestUrl(url string, callback DocumentCallbackFunc) (int, error) {

	timer := time.Now()

	request, err := http.NewRequest("GET", url, nil)
	request.SetBasicAuth(self.user, self.password)

	if self.logLevel > 0 {
		log.Printf("-> GET %s", url)
		if self.logLevel > 1 {
			dump, err := httputil.DumpRequestOut(request, (self.logLevel > 2))
			if err == nil {
				log.Printf("-> %s", dump)
			}
		}
	}

	resp, err := self.client.Do(request)

	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	if self.logLevel > 0 {
		log.Printf("<- %s, Time: %s, Length: %d :: %s", resp.Status, time.Since(timer), resp.ContentLength, url)
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
		return -1, err
	}

	if callback != nil {
		callback(resp.StatusCode, url, doc)
	}

	return resp.StatusCode, nil
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

func GetElementsByTag(node *html.Node, tag string) *list.List {

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

func GetElementsByClass(node *html.Node, name string) *list.List {

	selected := list.New()

	visit(node,
		func(n *html.Node) bool {
			if name == GetAttribute(n, "class") {
				return true
			}
			return false
		},
		func(n *html.Node) {
			selected.PushBack(n)
		})

	return selected
}

func GetElementsById(node *html.Node, name string) *list.List {

	selected := list.New()

	visit(node,
		func(n *html.Node) bool {
			if name == GetAttribute(n, "id") {
				return true
			}
			return false
		},
		func(n *html.Node) {
			selected.PushBack(n)
		})

	return selected
}

func GetText(node *html.Node) string {

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

func GetAttribute(node *html.Node, name string) string {
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

func First(nodes *list.List) *html.Node {
	if nodes.Len() > 0 {
		return nodes.Front().Value.(*html.Node)
	}
	return nil
}
