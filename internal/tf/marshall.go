package tf

func ExpandStringSlicePtr(input []interface{}) *[]string {
	result := make([]string, len(input))
	for i, item := range input {
		result[i] = item.(string)
	}
	return &result
}

func FlattenStringSlicePtr(input *[]string) []interface{} {
	result := make([]interface{}, 0)
	if input != nil {
		for _, item := range *input {
			result = append(result, item)
		}
	}
	return result
}
