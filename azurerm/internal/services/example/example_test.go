package example

import (
	"reflect"
	"testing"
)

func TestAccAzureRMExample_basic(t *testing.T) {
	//data := ResourceMetaData{}
	//data.Decode(map[string]interface{}{
	//	"name":"tom",
	//	"list":[]interface{}{map[string]interface{}{"name":"test"}},
	//})
	input := &ExampleObj{
	}
	valueToSet := []interface{}{
		map[string]interface{}{
			"name": "test1232",
			"inner": []interface{}{
				map[string]interface{}{
					"name": "get-a-mac",
				},
			},
		},
	}

	objType := reflect.TypeOf(input).Elem()
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		if field.Name == "List" {
			setValue(input, valueToSet, field, i)
		}
	}
}
