package mockgen

//go:generate go run go.uber.org/mock/mockgen -package=mockgen -destination=b.mock.go . B

type B interface {
	Bar(x int) int
}

func MAINB(f B) {
	// ...
}
