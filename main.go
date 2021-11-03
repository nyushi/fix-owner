package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var baseuid int
var root string
var skipdirs = []string{"/proc", "/dev", "/run", "/sys"}

func newid(id int) int {
	if id >= baseuid {
		return id - baseuid
	}
	return id
}

func walker(path string, info fs.FileInfo, err error) error {
	for _, prefix := range skipdirs {
		if strings.HasPrefix(path, root+prefix) {
			return filepath.SkipDir
		}
	}
	if err != nil {
		log.Printf("err at walker: path=`%s`, err=%s", path, err)
		return nil
	}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		uid := int(stat.Uid)
		gid := int(stat.Gid)
		if uid >= baseuid || gid >= baseuid {
			newuid := newid(uid)
			newgid := newid(gid)
			os.Chown(path, newuid, newgid)
			log.Printf("%s %d->%d:%d->%d", path, uid, newuid, gid, newgid)
		}
	} else {
		log.Fatalf("can't get info from %s", path)
	}
	return nil
}

func main() {
	flag.Parse()
	root = flag.Arg(0)
	u, err := strconv.ParseInt(flag.Arg(1), 10, 64)
	if err != nil {
		log.Fatalf("invalid baseuid: %s, %s", flag.Arg(1), err)
	}
	baseuid = int(u)
	if err := filepath.Walk(root, walker); err != nil {
		log.Fatalf("err: %s", err)
	}
	log.Println("OK")
}
