package api

import "encoding/json"

func init() {
	jsonUnmarshalImpl = json.Unmarshal
}
