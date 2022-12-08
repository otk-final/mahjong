package ops

type TakeHead struct{}

func (t TakeHead) Take() int {
	return 0
}

type TakeTail struct {
}

func (t TakeTail) Take() int {
	return 0
}
