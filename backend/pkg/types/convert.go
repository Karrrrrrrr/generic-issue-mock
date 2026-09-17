package types

// PointerSlice preserves an explicitly supplied value, including its zero value.
func PointerSlice[T any](value *T) []T {
	if value == nil {
		return nil
	}

	return []T{*value}
}

func ConvertPointer[Input any, Output any](value *Input, convert func(Input) Output) *Output {
	if value == nil {
		return nil
	}

	result := convert(*value)
	return &result
}

func BulkConvertSlice[Input any, Output any](items []Input, convert func(Input) Output) []Output {
	result := make([]Output, 0, len(items))
	for _, item := range items {
		result = append(result, convert(item))
	}

	return result
}
