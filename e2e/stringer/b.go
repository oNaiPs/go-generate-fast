package gogenerateng

//go:generate stringer -type B -output stringer_out.go
type B int

const (
	B1 B = 1
	B2 B = 2
	B3 B = 3
)
