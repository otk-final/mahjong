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
var ui embed.FS

func main() {

	subUI, _ := fs.Sub(ui, "ui")

	muxRouter := server.NewApiRouter()
	muxRouter.PathPrefix("/").Handler(http.FileServer(http.FS(subUI)))

	//服务
	srv := &http.Server{
		Handler:      muxCors.Handler(muxRouter),
		Addr:         *httpAddr,
		WriteTimeout: time.Duration(15) * time.Second,
		ReadTimeout:  time.Duration(15) * time.Second,
	}
	log.Println("api sever start")
	log.Fatal(srv.ListenAndServe())
}
