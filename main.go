package main

import (
	"log"
	"mahjong/server"
	"net/http"
	"time"
)

func main() {
	root := server.ApiRegister()
	//服务
	srv := &http.Server{
		Handler:      root,
		Addr:         ":7070",
		WriteTimeout: time.Duration(15) * time.Second,
		ReadTimeout:  time.Duration(15) * time.Second,
	}

	log.Println("api sever start")
	log.Fatal(srv.ListenAndServe())
}
