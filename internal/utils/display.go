package utils

import (
	"encoding/json"
	"fmt"
)

func JsonDumps(data any) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal data to JSON: %v", err))
	}
	return string(jsonData)
}
