package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	vsAPI "github.com/Percona-Lab/percona-version-service/versionpb/api"
)

// marshal marshals ProductResponse to JSON, ensuring the "critical" field is always included,
// without requiring modifications to the `versionpb/api` package or creating custom types that
// implement the json.Marshaler interface.
// Use protojson.Marshal instead if omitting the "critical" field is acceptable.
func marshal(product *vsAPI.ProductResponse) ([]byte, error) {
	m, err := productToMap(product)
	if err != nil {
		return nil, fmt.Errorf("json conversion: %w", err)
	}

	content, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal product response: %w", err)
	}
	return content, nil
}

// productToMap is a recursive function that converts a ProductResponse into a map.
// The resulting map can be used with json.Marshal, ensuring that fields like "critical" are included.
func productToMap(v any) (any, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() { // skip nil values
			return nil, nil
		}
		val = val.Elem() // dereference
	}

	switch val.Kind() {
	case reflect.Struct:
		if val.NumField() == 0 {
			return nil, nil
		}
		m := make(map[string]any)
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			fieldValue := val.Field(i)

			jsonTag := field.Tag.Get("json")
			if jsonTag == "" {
				continue
			}
			jsonFieldName := strings.Split(jsonTag, ",")[0]
			omitempty := slices.Contains(strings.Split(jsonTag, ","), "omitempty")
			if fieldValue.Kind() == reflect.Bool {
				omitempty = false // do not omit bool values like "critical"
			}

			if omitempty {
				zeroValue := reflect.Zero(fieldValue.Type())

				// check if the value is equal to its zero value
				if reflect.DeepEqual(fieldValue.Interface(), zeroValue.Interface()) {
					continue
				}
			}

			if status, ok := fieldValue.Interface().(vsAPI.Status); ok {
				m[jsonFieldName] = vsAPI.Status_name[int32(status)]
				continue
			}

			fieldVal, err := productToMap(fieldValue.Interface())
			if err != nil {
				return nil, err
			}

			m[jsonFieldName] = fieldVal
		}
		return m, nil
	case reflect.Slice:
		var slice []any
		for j := 0; j < val.Len(); j++ {
			element := val.Index(j)
			elementValue, err := productToMap(element.Interface())
			if err != nil {
				return nil, err
			}
			if elementValue == nil {
				continue
			}
			slice = append(slice, elementValue)
		}
		return slice, nil
	case reflect.Map:
		m := make(map[string]any)
		for _, key := range val.MapKeys() {
			value := val.MapIndex(key)

			elementValue, err := productToMap(value.Interface())
			if err != nil {
				return nil, err
			}
			m[key.Interface().(string)] = elementValue
		}
		return m, nil
	default:
		return val.Interface(), nil
	}
}
