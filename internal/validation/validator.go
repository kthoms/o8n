package validation

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ValidateAndParse validates input string according to inputType and returns parsed value
// Supported inputType: bool,int,number,text,json
func ValidateAndParse(input string, inputType string) (interface{}, error) {
	trimmed := strings.TrimSpace(input)
	switch inputType {
	case "bool", "boolean":
		if strings.EqualFold(trimmed, "true") || trimmed == "1" {
			return true, nil
		}
		if strings.EqualFold(trimmed, "false") || trimmed == "0" {
			return false, nil
		}
		return nil, fmt.Errorf("enter true or false")
	case "int", "integer":
		v, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("enter an integer")
		}
		return v, nil
	case "number", "double", "float":
		v, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return nil, fmt.Errorf("enter a number")
		}
		return v, nil
	case "json":
		var j interface{}
		if err := json.Unmarshal([]byte(trimmed), &j); err != nil {
			return nil, fmt.Errorf("invalid json: %v", err)
		}
		return j, nil
	default:
		// default to raw text
		return input, nil
	}
}
