package moq

//go:generate moq -out a_moq_test.go . MyInterface

type MyInterface interface {
	Method1() error
	Method2(i int)
}
