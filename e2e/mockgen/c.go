package mockgen

//go:generate env TEST_ENV_VAR=value go run go.uber.org/mock/mockgen -package=mockgen -source=c.go -destination=c.mock.go

type C interface {
	Baz(x int) int
}

func MAINC(f C) {
	// ...
}
