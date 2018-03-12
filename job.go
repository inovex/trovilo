package main

func compareCMLabels(expected *map[string]string, actual *map[string]string) bool {
	if len((*actual)) == 0 {
		// immediately abort if there are no labels at all
		return false
	}

	for key, expectedValue := range *expected {
		actualValue, found := (*actual)[key]
		if !found || expectedValue != actualValue {
			return false
		}
	}
	return true
}
