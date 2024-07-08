package mustjson

import "encoding/json"

// Encode returns the JSON encoding of v. It panics if the error returned by json.Marshal is not nil.
func Encode(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}
