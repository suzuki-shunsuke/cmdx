package handler

import "fmt"

type StrList []string

func (list *StrList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var val interface{}
	if err := unmarshal(&val); err != nil {
		return err
	}
	if s, ok := val.(string); ok {
		*list = []string{s}
		return nil
	}
	if intfArr, ok := val.([]interface{}); ok {
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
