package main

import (
  "flag"
  "log"
  "os"
  "bazil.org/fuse"
  "bazil.org/fuse/fs"
)

var rootNode Node;

func main() {
  flag.Parse();

  if flag.NArg() != 2 {
    os.Exit(2)
  }

  mountpoint := flag.Arg(0)
  url := flag.Arg(1)

  rootNode.Files = make(map[string]*Node)
  rootNode.Type = fuse.DT_Dir
  rootNode.Name = "/"
  go Crawl(url)

  c, err := fuse.Mount(mountpoint)
  if err != nil {
    log.Fatal(err)
  }

  fs.Serve(c, FS{})
}

