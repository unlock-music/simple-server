package main

import (
	"embed"
	"fmt"
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
var listen = "localhost:8086"

func main() {
	if tmp := os.Getenv("UM_LISTEN"); tmp != "" {
		listen = tmp
	}
	pfxFs := WithPrefix(asset, "build/for-build")
	go func() {
		err := http.ListenAndServe(listen, http.FileServer(http.FS(pfxFs)))
		if err != nil {
			log.Println("启动出错，请检查是否有其他程序占用了端口：" + listen)
			_, _ = fmt.Scanln()
			log.Fatal(err)
		}
	}()
	log.Printf("使用浏览器打开: %s 即可访问。", listen)
	err := browser.OpenURL("http://" + listen)
	if err != nil {
		log.Println("自动打开浏览器错误，需要手动打开: "+listen, err)
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
