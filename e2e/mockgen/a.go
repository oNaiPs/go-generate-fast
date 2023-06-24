package mockgen

//go:generate go run go.uber.org/mock/mockgen -package=mockgen -source=a.go -destination=a.mock.go

type A interface {
	Bar(x int) int
}

func MAINA(f A) {
	// ...
}
