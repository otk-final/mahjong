package ploy

import (
	"container/ring"
	"fmt"
	"log"
	"testing"
)

func TestRing(t *testing.T) {

	rs := ring.New(10)
	for i := 0; i < rs.Len(); i++ {
		rs.Value = fmt.Sprintf("node-%d", i+1)
		rs = rs.Next()
	}
	rs.Unlink(2)

	rs.Do(func(a any) {
		log.Println(a)
	})

}
