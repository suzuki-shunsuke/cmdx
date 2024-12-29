package domain

import (
	"fmt"

	"github.com/invopop/jsonschema"
)

type StrList []string

func (StrList) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		OneOf: []*jsonschema.Schema{
			{
				Type: "string",
			},
			{
				Type: "array",
				Items: &jsonschema.Schema{
					Type: "string",
				},
			},
		},
	}
}

func (list *StrList) UnmarshalYAML(unmarshal func(any) error) error {
	var val any
	if err := unmarshal(&val); err != nil {
		return err
	}
	if s, ok := val.(string); ok {
		*list = []string{s}
		return nil
	}
	if intfArr, ok := val.([]any); ok {
		strArr := make([]string, len(intfArr))
		for i, intf := range intfArr {
			if s, ok := intf.(string); ok {
				strArr[i] = s
				continue
			}
			return fmt.Errorf("the type of the value must be string: %v", intf)
		}
		*list = strArr
	}
	return nil
}
