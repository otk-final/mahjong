package ploy

import "testing"

func TestA(t *testing.T) {
	count := make(map[int]int, 0)
	count[1] = count[1] + 1
	t.Log(count[1])
}
