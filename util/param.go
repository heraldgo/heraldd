package util

// GetStringParam get the string param from the map
func GetStringParam(value *string, param map[string]interface{}, name string) {
	strParam, ok := param[name]
	if !ok {
		return
	}

	strValue, ok := strParam.(string)
	if !ok {
		return
	}

	*value = strValue
}

// GetIntParam get the int param from the map
func GetIntParam(value *int, param map[string]interface{}, name string) {
	intParam, ok := param[name]
	if !ok {
		return
	}

	intValue, ok := intParam.(int)
	if !ok {
		return
	}

	*value = intValue
}
