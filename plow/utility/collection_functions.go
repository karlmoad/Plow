package utility

func Index(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

func Include(vs []string, t string) bool {
	return Index(vs, t) >= 0
}

func Any(vs []string, f func(string) bool) bool {
	for _, v := range vs {
		if f(v) {
			return true
		}
	}
	return false
}

func All(vs []string, f func(string) bool) bool {
	for _, v := range vs {
		if !f(v) {
			return false
		}
	}
	return true
}

func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func DeepMapCopy(input map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range input {
		// Handle maps
		mv, isMap := v.(map[string]interface{})
		if isMap {
			result[k] = DeepMapCopy(mv)
			continue
		}

		// Handle slices
		sv, isSlice := v.([]interface{})
		if isSlice {
			result[k] = DeepSliceCopy(sv)
			continue
		}
		result[k] = v
	}
	return result
}

func DeepSliceCopy(input []interface{}) []interface{} {
	result := make([]interface{}, len(input))

	for _, v := range input {
		// Handle maps
		mv, isMap := v.(map[string]interface{})
		if isMap {
			result = append(result, DeepMapCopy(mv))
			continue
		}

		// Handle slices
		sv, isSlice := v.([]interface{})
		if isSlice {
			result = append(result, DeepSliceCopy(sv))
			continue
		}
		result = append(result, v)
	}
	return result
}
