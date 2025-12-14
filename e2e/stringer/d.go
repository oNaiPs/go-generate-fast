package gogenerateng

//go:generate go tool stringer -type D -trimprefix D
type D int

const (
	D1 D = 1
	D2 D = 2
	D3 D = 3
)
