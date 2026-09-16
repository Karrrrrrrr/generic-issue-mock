package types

func BulkConvertSlice[Input any, Output any](items []Input, convert func(Input) Output) []Output {
	result := make([]Output, 0, len(items))
	for _, item := range items {
		result = append(result, convert(item))
	}

	return result
}
