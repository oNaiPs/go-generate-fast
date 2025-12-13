package gogenerateng

//go:generate go run golang.org/x/tools/cmd/stringer -type A -trimprefix A
type A int

const (
	A1 A = 1
	A2 A = 2
	A3 A = 3
)
