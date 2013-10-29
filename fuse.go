package main

import (
  "os"
  "bazil.org/fuse"
  "bazil.org/fuse/fs"
)

type FS struct {}

func (FS) Root() (fs.Node, fuse.Error) {
  return Dir{rootNode}, nil
}

type Dir struct {
  Node
}

type File struct {
  Node
}

type Node struct {
  Name    string
  Url     string
  Type    fuse.DirentType
  Size    int64
  Files   map[string]*Node
}

func (n Node) Attr() fuse.Attr {
  if n.Type == fuse.DT_Dir {
    return fuse.Attr { Mode: os.ModeDir, Size: uint64(n.Size) }
  } else {
    return fuse.Attr { Mode: 0777 }
  }
}

func (d Dir) Lookup(name string, intr fs.Intr) (fs.Node, fuse.Error) {
  resNode := d.Files[name]

  if resNode != nil {
    return Dir{*resNode}, nil
  } else {
    return nil, fuse.ENOENT
  }
}

func (d Dir) ReadDir(intr fs.Intr) ([]fuse.Dirent, fuse.Error) {
  res := []fuse.Dirent{}
  for key, value := range d.Files {
    res = append(res, fuse.Dirent{Name: key, Type: value.Type})
  }
  return res, nil
}
