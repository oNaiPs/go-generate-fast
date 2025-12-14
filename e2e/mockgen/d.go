package mockgen

//go:generate go run go.uber.org/mock/mockgen -package $GOPACKAGE -source $GOFILE -destination d.mock.go

type D interface {
	Qux(x int) int
}

func MAIND(f D) {
	// ...
}
