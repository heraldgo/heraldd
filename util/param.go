package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// DeepCopyParam returns a deep copied json param object
func DeepCopyParam(param interface{}) interface{} {
	paramSlice, ok := param.([]interface{})
	if ok {
		resultSlice := make([]interface{}, 0, len(paramSlice))
		for _, value := range paramSlice {
			resultSlice = append(resultSlice, DeepCopyParam(value))
		}
		return resultSlice
	}

	paramMap, ok := param.(map[string]interface{})
	if ok {
		resultMap := make(map[string]interface{})
		for key, value := range paramMap {
			resultMap[key] = DeepCopyParam(value)
		}
		return resultMap
	}

	return param
}

// DeepCopyMapParam returns a deep copied map param object
func DeepCopyMapParam(param map[string]interface{}) map[string]interface{} {
	paramNew, _ := DeepCopyParam(param).(map[string]interface{})
	return paramNew
}

// MergeMapParam merges the two maps
func MergeMapParam(mapOrigin, mapNew map[string]interface{}) {
	for k, v := range mapNew {
		mapOrigin[k] = DeepCopyParam(v)
	}
}

// InterfaceMapToStringMap will only keep map with string keys
func InterfaceMapToStringMap(param interface{}) interface{} {
	paramSlice, ok := param.([]interface{})
	if ok {
		resultSlice := make([]interface{}, 0, len(paramSlice))
		for _, value := range paramSlice {
			resultSlice = append(resultSlice, InterfaceMapToStringMap(value))
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
			resultMap[keyString] = InterfaceMapToStringMap(value)
		}
		return resultMap
	}

	return param
}

// JSONToMap convert json to map
func JSONToMap(text []byte) (map[string]interface{}, error) {
	var outputJSON interface{}
	err := json.Unmarshal(text, &outputJSON)
	if err != nil {
		return nil, fmt.Errorf("Parse json error (%s)", err)
	}
	outputMap, ok := outputJSON.(map[string]interface{})
	if !ok {
		return nil, errors.New("Input json is not a map")
	}
	return outputMap, nil
}

// GetStringParam get the string param from the map
func GetStringParam(param map[string]interface{}, name string) (string, error) {
	strParam, ok := param[name]
	if !ok {
		return "", fmt.Errorf(`Param "%s" not found`, name)
	}

	strValue, ok := strParam.(string)
	if !ok {
		return "", fmt.Errorf(`Param "%s" is not a string`, name)
	}

	return strValue, nil
}

// GetIntParam get the int param from the map
func GetIntParam(param map[string]interface{}, name string) (int, error) {
	intParam, ok := param[name]
	if !ok {
		return 0, fmt.Errorf(`Param "%s" not found`, name)
	}

	intValue, ok := intParam.(int)
	if !ok {
		return 0, fmt.Errorf(`Param "%s" is not an integer`, name)
	}

	return intValue, nil
}

// GetFloatParam get the float param from the map
func GetFloatParam(param map[string]interface{}, name string) (float64, error) {
	floatParam, ok := param[name]
	if !ok {
		return 0, fmt.Errorf(`Param "%s" not found`, name)
	}

	floatValue, ok := floatParam.(float64)
	if !ok {
		return 0, fmt.Errorf(`Param "%s" is not a float`, name)
	}

	return floatValue, nil
}

// GetBoolParam get the bool param from the map
func GetBoolParam(param map[string]interface{}, name string) (bool, error) {
	boolParam, ok := param[name]
	if !ok {
		return false, fmt.Errorf(`Param "%s" not found`, name)
	}

	boolValue, ok := boolParam.(bool)
	if !ok {
		return false, fmt.Errorf(`Param "%s" is not a bool`, name)
	}

	return boolValue, nil
}

// GetMapParam get the map param from the map
func GetMapParam(param map[string]interface{}, name string) (map[string]interface{}, error) {
	mapParam, ok := param[name]
	if !ok {
		return nil, fmt.Errorf(`Param "%s" not found`, name)
	}

	mapValue, ok := mapParam.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(`Param "%s" is not a map`, name)
	}

	return mapValue, nil
}

// GetStringSliceParam get the string slice param from the map (string or slice of strings)
func GetStringSliceParam(param map[string]interface{}, name string) ([]string, error) {
	strSliceParam, ok := param[name]
	if !ok {
		return nil, fmt.Errorf(`Param "%s" not found`, name)
	}

	strValue, ok := strSliceParam.(string)
	if ok {
		return []string{strValue}, nil
	}

	sliceValue, ok := strSliceParam.([]interface{})
	if !ok {
		return nil, fmt.Errorf(`Param "%s" is not a string or slice`, name)
	}

	strSliceValue := make([]string, 0, len(sliceValue))
	for _, value := range sliceValue {
		valueString, ok := value.(string)
		if !ok {
			continue
		}
		strSliceValue = append(strSliceValue, valueString)
	}

	return strSliceValue, nil
}

// GetNestedMapValue get the value of a nested key from the map
// map["a/b/c"] = map["a"]["b"]["c"]
func GetNestedMapValue(param map[string]interface{}, nestedKey string) (interface{}, error) {
	var currentParam interface{}

	frags := strings.Split(nestedKey, "/")
	currentKeys := make([]string, 0, len(frags))

	currentParam = param
	for _, frag := range frags {
		currentMap, ok := currentParam.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf(`Get nested map value error: "%s" is not a map`, strings.Join(currentKeys, "/"))
		}

		currentKeys = append(currentKeys, frag)

		currentParam, ok = currentMap[frag]
		if !ok {
			return nil, fmt.Errorf(`Get nested map value error: "%s" does not exist`, strings.Join(currentKeys, "/"))
		}
	}

	return currentParam, nil
}
