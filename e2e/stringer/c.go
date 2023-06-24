package gogenerateng

type C uint8

//go:generate go run golang.org/x/tools/cmd/stringer -type C

const (
	C1 C = iota
	C2
	C3
	C4
)
