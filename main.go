package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	//root := server.ApiRegister()
	////服务
	//srv := &http.Server{
	//	Handler:      root,
	//	Addr:         ":7070",
	//	WriteTimeout: time.Duration(15) * time.Second,
	//	ReadTimeout:  time.Duration(15) * time.Second,
	//}
	//
	//log.Println("api sever start")
	//log.Fatal(srv.ListenAndServe())
	//var pidx int
	//var event int
	//var tile int

	sr := bufio.NewScanner(os.Stdin)
	for sr.Scan() {
		fmt.Printf("输入>> %s \n", sr.Text())
	}

	//fmt.Scanln("%d %d %d", &pidx, &event, &tile)
	//fmt.Printf("玩家：%d 事件 %d 内容 %d", pidx, event, tile)

}
