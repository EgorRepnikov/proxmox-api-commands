package utils

import jsoniter "github.com/json-iterator/go"

var (
	jsonIter = jsoniter.Config{
		EscapeHTML:             false,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
	}.Froze()
)

func JsonMarshal(v interface{}) ([]byte, error) {
	return jsonIter.Marshal(v)
}

func JsonUnmarshal(data []byte, v interface{}) error {
	return jsonIter.Unmarshal(data, v)
}
