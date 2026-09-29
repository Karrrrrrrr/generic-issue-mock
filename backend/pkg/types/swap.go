package types

func Swap[T any](first *T, second *T) {
	value := *first
	*first = *second
	*second = value
}
