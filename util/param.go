package util

import (
	"fmt"
)

// InterfaceToStringMap will only keep map with string keys
func InterfaceToStringMap(param interface{}) interface{} {
	paramSlice, ok := param.([]interface{})
	if ok {
		var resultSlice []interface{}
		for _, value := range paramSlice {
			resultSlice = append(resultSlice, InterfaceToStringMap(value))
		}
		return resultSlice
	}

	paramMap, ok := param.(map[interface{}]interface{})
	if ok {
		resultMap := make(map[string]interface{})
		for key, value := range paramMap {
			keyString, ok := key.(string)
			if !ok {
				continue
			}
			resultMap[keyString] = InterfaceToStringMap(value)
		}
		return resultMap
	}

	return param
}

// GetStringParam get the string param from the map
func GetStringParam(param map[string]interface{}, name string) (string, error) {
	strParam, ok := param[name]
	if !ok {
		return "", fmt.Errorf("Param \"%s\" not found", name)
	}

	strValue, ok := strParam.(string)
	if !ok {
		return "", fmt.Errorf("Param \"%s\" is not a string", name)
	}

	return strValue, nil
}

// GetIntParam get the int param from the map
func GetIntParam(param map[string]interface{}, name string) (int, error) {
	intParam, ok := param[name]
	if !ok {
		return 0, fmt.Errorf("Param \"%s\" not found", name)
	}

	intValue, ok := intParam.(int)
	if !ok {
		return 0, fmt.Errorf("Param \"%s\" is not a string", name)
	}

	return intValue, nil
}

// GetStringSliceParam get the string slice param from the map (string or slice of strings)
func GetStringSliceParam(param map[string]interface{}, name string) ([]string, error) {
	strSliceParam, ok := param[name]
	if !ok {
		return nil, fmt.Errorf("Param \"%s\" not found", name)
	}

	strValue, ok := strSliceParam.(string)
	if ok {
		return []string{strValue}, nil
	}

	sliceValue, ok := strSliceParam.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Param \"%s\" is not a string or slice", name)
	}

	var strSliceValue []string
	for _, value := range sliceValue {
		valueString, ok := value.(string)
		if !ok {
			continue
		}
		strSliceValue = append(strSliceValue, valueString)
	}

	return strSliceValue, nil
}

// UpdateStringParam get the string param from the map
func UpdateStringParam(value *string, param map[string]interface{}, name string) error {
	strValue, err := GetStringParam(param, name)
	if err != nil {
		return err
	}

	*value = strValue
	return nil
}

// UpdateIntParam get the int param from the map
func UpdateIntParam(value *int, param map[string]interface{}, name string) error {
	intValue, err := GetIntParam(param, name)
	if err != nil {
		return err
	}

	*value = intValue
	return nil
}
