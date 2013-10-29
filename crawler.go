package main

import (
  "net/http"
  "net/url"
  "io"
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

var htmlSuffixes map[string]bool = map[string]bool {
  "html" : true,
  "htm" : true,
  "aspx" : true,
  "" : true,
}

// Crawl takes an url string, get all the links from that url and call
// recursively on each readable link (HTML pages).
func Crawl(source string) {
  fmt.Println("Scraping", source)
  sourceUrl, _ := url.Parse(source)
  node, existing := CreateNode(sourceUrl)

  if existing {
    return
  }

  node.Url = source

  resp, _ := http.Get(source)
  // node.Size = resp.ContentLength
  links, node.Size := GetLinksFromHtml(resp.Body)

  for _, link := range links {
    u, _ := url.Parse(link)
    urlToCrawl := link

    // We check if the url is on the same host, if not we skip it
    if u.IsAbs() {
      if u.Host != sourceUrl.Host {
        continue
      }
    } else {
      urlToCrawl = "http://" + sourceUrl.Host + "/" + u.Path
    }

    // If this is a html page, we parse it, else we just need to get its size
    suffix := GetSuffix(u.Path)
    fmt.Println("suffix", suffix, u.Path)
    if _, ok := htmlSuffixes[suffix]; ok {
      Crawl(urlToCrawl)
    } else {
      InspectRessource(urlToCrawl)
    }
  }
}

func InspectRessource(url string) {
  // TODO
}

func GetSuffix(path string) string {
  res := ""

  for i := len(path) - 1; i >= 0; i-- {
    switch path[i] {
      case '.':
        return res
      case '/':
        return ""
      default:
        res = string(path[i]) + res
    }
  }

  return ""
}

// GetLinksFromHtml takes an url, retrieve the content and parse the HTML to
// find all embedded links.
func GetLinksFromHtml(body io.ReadCloser) []string {
  tokenizer := html.NewTokenizer(body)
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
