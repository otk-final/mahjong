package main

import (
	"embed"
	"flag"
	"github.com/rs/cors"
	"io/fs"
	"log"
	"mahjong/server"
	"net/http"
	"time"
)

//跨域规则
var muxCors = cors.AllowAll()
var httpAddr = flag.String("addr", ":7070", "")

//go:embed ui
var webUi embed.FS

func main() {
	flag.Parse()

	//静态文件
	subUI, _ := fs.Sub(webUi, "ui")

	muxRouter := server.NewApiRouter()
	muxRouter.PathPrefix("/").Handler(http.FileServer(http.FS(subUI)))

	//服务
	srv := &http.Server{
		Handler:      muxCors.Handler(muxRouter),
		Addr:         *httpAddr,
		WriteTimeout: time.Duration(30) * time.Second,
		ReadTimeout:  time.Duration(30) * time.Second,
	}
	log.Printf("api sever start addr %s", *httpAddr)
	log.Fatal(srv.ListenAndServe())
}
