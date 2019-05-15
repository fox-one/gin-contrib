package gin_helper

import (
	"github.com/iancoleman/strcase"
)

func lookupJsonKey(data []byte) (int, string) {
	var left, right int

	for idx, b := range data {
		switch b {
		case '"':
			if left == 0 || right > 0 {
				left = idx
				right = 0
			} else if right == 0 {
				right = idx
			}
		case ':':
			if left > 0 && right > 0 {
				return left + 1, string(data[left+1 : right])
			}
		default:
			if right > 0 {
				left = 0
				right = 0
			}
		}
	}

	return -1, ""
}

type JsonKeyTransformer func(string) string

func TransformJsonKeys(data []byte, transformer JsonKeyTransformer) []byte {
	buffer := make([]byte, 0, len(data))

	for {
		idx, key := lookupJsonKey(data)

		if idx == -1 {
			buffer = append(buffer, data...)
			break
		}

		buffer = append(buffer, data[:idx]...)
		data = data[idx+len(key):]

		key = transformer(key)
		buffer = append(buffer, []byte(key)...)
	}

	return buffer
}

var ToCamelKey JsonKeyTransformer = strcase.ToLowerCamel
var ToSnakeKey JsonKeyTransformer = strcase.ToSnake

func TransformWithDict(fn JsonKeyTransformer, dict map[string]string) JsonKeyTransformer {
	if len(dict) == 0 {
		return fn
	}

	return func(key string) string {
		if v, ok := dict[key]; ok {
			return v
		}

		return fn(key)
	}
}

func TransformWithWhitelists(fn JsonKeyTransformer, list ...string) JsonKeyTransformer {
	dict := make(map[string]string, len(list))
	for _, key := range list {
		dict[key] = key
	}

	return TransformWithDict(fn, dict)
}

func (fn JsonKeyTransformer) WithDict(dict map[string]string) JsonKeyTransformer {
	return TransformWithDict(fn, dict)
}

func (fn JsonKeyTransformer) WithWhitelist(list ...string) JsonKeyTransformer {
	return TransformWithWhitelists(fn, list...)
}
