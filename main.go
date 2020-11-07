package main

import (
	"embed"
	_ "embed"
	"github.com/pkg/browser"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
)

//go:generate go run ./builder
//go:embed build/for-build
var asset embed.FS

func main() {
	pfxFs := WithPrefix(asset, "build/for-build")
	go func() {
		err := http.ListenAndServe("localhost:6280", http.FileServer(http.FS(pfxFs)))
		if err != nil {
			log.Fatal(err)
		}
	}()
	log.Println("you can now open browser with: http://localhost:6280 to access this tool.")
	err := browser.OpenURL("http://localhost:6280")
	if err != nil {
		log.Println("error while opening browser:", err)
	}
	sign := make(chan os.Signal, 1)
	signal.Notify(sign, os.Interrupt, os.Kill)
	<-sign
}

type WrappedFS struct {
	inner  fs.FS
	prefix string
}

func (f WrappedFS) Open(p string) (fs.File, error) {
	p = path.Join(f.prefix, p)
	log.Printf("serving: %s\n", p)
	return f.inner.Open(p)
}

// same thing but for ReadFile, Stat, ReadDir, Glob, and maybe Rename, OpenFile?

func WithPrefix(f fs.FS, p string) WrappedFS {
	return WrappedFS{f, p}
}
