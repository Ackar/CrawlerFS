package main

import (
  "net/http"
  "net/url"
  "fmt"
  "strings"
  "code.google.com/p/go.net/html"
  "bazil.org/fuse"
)

// CreateNode take an URL and return the corresponding Node, creating all
// necessary nodes.
// The function returns a Node and a boolean indicating whether or not the node
// existed already.
func CreateNode(u *url.URL) (resNode *Node, existing bool) {
  fmt.Println("Creating node", u)
  var n *Node = nil
  existing = true
  resNode = rootNode.Files[u.Host]

  // If the root node (the host name) does not exist, we create it
  if resNode == nil {
    existing = false
    n = new (Node)
    n.Type = fuse.DT_Dir
    n.Name = u.Host
    n.Files = make(map[string]*Node)
    rootNode.Files[u.Host] = n
    resNode = n
  }

  // We go through each node on the path, creating them if they don't exist
  folders := strings.Split(u.Path, "/")
  for i, folder := range folders {
    if folder == "" {
      continue
    }
    n = resNode.Files[folder]
    existing = true
    if n == nil {
      existing = false
      n = new (Node)
      if i == len(folders) - 1 {
        n.Type = fuse.DT_File
      } else {
        n.Type = fuse.DT_Dir
      }
      n.Name = folder
      n.Files = make(map[string]*Node)
      resNode.Files[folder] = n
    }
    resNode = n
  }

  return
}

// Crawl takes an url string, get all the links from that url and call
// recursively on each readable link (HTML pages).
func Crawl(source string) {
  fmt.Println("Scraping", source)
  sourceUrl, _ := url.Parse(source)
  _, existing := CreateNode(sourceUrl)

  if existing {
    return
  }

  links := GetLinksFromHtml(source)

  for _, link := range links {
    u, _ := url.Parse(link)
    // FIXME
    if !strings.HasSuffix(u.Path, ".html") && !strings.HasSuffix(u.Path, ".htm") && !strings.HasSuffix(u.Path, ".aspx") {
      fmt.Println("ingoring ", u.Path)
      continue
    }
    if u.IsAbs() {
      if u.Host == sourceUrl.Host {
        Crawl(u.String())
      }
    } else {
      Crawl("http://" + sourceUrl.Host + "/" + u.Path)
    }
  }
}

// GetLinksFromHtml takes an url, retrieve the content and parse the HTML to
// find all embedded links.
func GetLinksFromHtml(url string) []string {
  resp, _ := http.Get(url)
  tokenizer := html.NewTokenizer(resp.Body)
  res := []string{}

  for {
    tokenType := tokenizer.Next()
    if tokenType == html.ErrorToken {
      break
    }

    token := tokenizer.Token()
    switch tokenType {
      case html.StartTagToken:
        for _, attr := range token.Attr {
          if attr.Key == "href" || attr.Key == "src" {
            // If the link is just an anchor, we ignore it
            if len(attr.Val) > 0 && attr.Val[0] == '#' {
              continue;
            }

            res = append(res, attr.Val)
          }
        }
    }
  }

  return res
}
