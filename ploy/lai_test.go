package ploy

import (
	"log"
	"testing"
	"time"
)

func TestLaiWin(t *testing.T) {
	now := time.Now()

	next := now.Add(30 * time.Second)

	log.Println(next.Sub(now).Seconds())
}
