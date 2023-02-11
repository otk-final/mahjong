package ploy

import (
	"log"
	"testing"
	"time"
)

func TestA(t *testing.T) {

	ch := make(chan int, 0)

	go func() {
		defer close(ch)
		for {
			select {
			case a, ok := <-ch:
				if !ok {
					return
				}
				log.Printf("rev %d", a)
				if a == 10 {
					return
				}
			case <-time.After(3 * time.Second):

			}
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		d := 0
		for {
			<-ticker.C
			d++
			ch <- d
		}
	}()
	select {}
}
