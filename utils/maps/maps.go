package maps

// Merge bulabula
func Merge(maps ...map[string]interface{}) map[string]interface{} {
	if len(maps) == 0 {
		return nil
	}
	u := make(map[string]interface{})
	for _, m := range maps {
		if m != nil {
			for k, v := range m {
				u[k] = v
			}
		}
	}
	return u
}

// MergeString bulabula
func MergeString(maps ...map[string]string) map[string]string {
	if len(maps) == 0 {
		return nil
	}
	u := make(map[string]string)
	for _, m := range maps {
		if m != nil {
			for k, v := range m {
				u[k] = v
			}
		}
	}
	return u
}
