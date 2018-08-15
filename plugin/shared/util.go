package shared

import "reflect"

func ExpandStringMap(sr reflect.Value) map[string]reflect.Value {
	keys := sr.MapKeys()
	m := make(map[string]reflect.Value, len(keys))
	for _, key := range keys {
		m[key.String()] = sr.MapIndex(key)
	}
	return m
}
