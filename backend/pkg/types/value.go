package types

const DefaultPageSize = 20

func Value[T any](input *T) T {
	if input == nil {
		var zero T
		return zero
	}

	return *input
}

func NormalizePagination(page int, size int) (int, int) {
	if page < 1 {
		page = 1
	}

	if size < 1 {
		size = DefaultPageSize
	}

	return page, size
}
