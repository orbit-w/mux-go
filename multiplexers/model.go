package multiplexers

const (
	StateNone = iota
	StateClosed
)

var (
	incrPriority = func(s int) int {
		return s + 1
	}

	decrPriority = func(s int) int {
		return s - 1
	}
)
