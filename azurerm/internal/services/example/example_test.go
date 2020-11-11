package example

import (
	"reflect"
	"testing"
)

//data := ResourceMetaData{}
//data.Decode(map[string]interface{}{
//	"name":"tom",
//	"list":[]interface{}{map[string]interface{}{"name":"test"}},
//})


func TestDecode(t *testing.T) {
	testCases := []struct{
		Name string
		Input map[string]interface{}
		Expected *ExampleObj
		ExpectError bool
	}{
		{
			Name: "top level - name",
			Input: map[string]interface{}{
				"name": "bingo bango",
			},
			Expected: &ExampleObj{
				Name: "bingo bango",
			},
			ExpectError: false,
		},
		{
			Name: "top level - everything",
			Input: map[string]interface{}{
				"name": "bingo bango",
				"float": 123.4,
				"number": 123,
				"enabled": false,
			},
			Expected: &ExampleObj{
				Name: "bingo bango",
				Float: 123.4,
				Number: 123,
				Enabled: false,
			},
			ExpectError: false,
		},
		{
			Name: "top level - list",
			Input: map[string]interface{}{
				"name": "bingo bango",
				"float": 123.4,
				"number": 123,
				"enabled": false,
				"list": []interface{}{
					map[string]interface{}{
						"name": "first",
					},
				},
			},
			Expected: &ExampleObj{
				Name: "bingo bango",
				Float: 123.4,
				Number: 123,
				Enabled: false,
				List: []NetworkList{{
					Name: "first",
				}},
			},
			ExpectError: false,
		},
		{
			Name: "top level - list in lists",
			Input: map[string]interface{}{
				"name": "bingo bango",
				"float": 123.4,
				"number": 123,
				"enabled": false,
				"list": []interface{}{
					map[string]interface{}{
						"name": "first",
						"inner": []interface{}{
							map[string]interface{}{
								"name": "get-a-mac",
							},
						},
					},
				},
			},
			Expected: &ExampleObj{
				Name: "bingo bango",
				Float: 123.4,
				Number: 123,
				Enabled: false,
				List: []NetworkList{{
					Name: "first",
					Inner: []NetworkInner{{
						Name: "get-a-mac",
					}},
				}},
			},
			ExpectError: false,
		},
	}


	for _, v := range testCases {
		obj := &ExampleObj{}
		decodeHelper(obj, v.Input)
		
		if !reflect.DeepEqual(obj, v.Expected) {
			t.Fatalf("ExampleObj mismatch\n\n Expected: %+v\n\n Received %+v\n\n", v.Expected, obj)
		}
	}
}

func TestAccAzureRMExample_single(t *testing.T) {
	input := &ExampleObj{}
	valueToSet := map[string]interface{} {
		"name": "bingo bango",
		"list": []interface{}{
			map[string]interface{}{
				"name": "first",
				"inner": []interface{}{
					map[string]interface{}{
						"name": "get-a-mac",
					},
				},
			},
		},
	}

	decodeHelper(input, valueToSet)
	if input.Name == valueToSet["name"] {
		t.Errorf("name does not match: Expected: %q Received: %q", valueToSet["name"], input.Name)
	}
}

func TestAccAzureRMExample_two(t *testing.T) {
	input := &ExampleObj{
	}
	valueToSet := []interface{}{
		map[string]interface{}{
			"name": "first",
			"inner": []interface{}{
				map[string]interface{}{
					"name": "get-a-mac",
				},
			},
		},
		map[string]interface{}{
			"name": "second",
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

func TestAccAzureRMExample_three(t *testing.T) {
	input := &ExampleObj{
	}
	valueToSet := []interface{}{
		map[string]interface{}{
			"name": "first",
			"inner": []interface{}{
				map[string]interface{}{
					"name": "get-a-mac",
				},
			},
		},
		map[string]interface{}{
			"name": "second",
		},
		map[string]interface{}{
			"name": "third",
		},
	}

	objType := reflect.TypeOf(input).Elem()
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		if field.Name != "List" {
			setValue(input, valueToSet, field, i)
		}
	}
}

func TestAccAzureRMExample_listInList(t *testing.T) {
	input := &ExampleObj{
	}
	valueToSet := []interface{}{
		map[string]interface{}{
			"name": "first",
			"inner": []interface{}{
				map[string]interface{}{
					"name": "get-a-mac",
					"inner": []interface{}{
						map[string]interface{}{
							"name": "get-a-mac",
							"inner": []interface{}{
								map[string]interface{}{
									"name": "get-a-mac",
									"should_be_fine": true,
								},
							},
						},
					},
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


type TopLevelString struct {
	Hello string `hcl:"hello"`
}


/*
func TestEncodeMissingTags(t *testing.T) {
	// top level of each type
	// nested of each type
	// nested of nested
	// one 5 nested deep
}

*/

func decodeHelper(input interface{}, config map[string]interface{}) {
	objType := reflect.TypeOf(input).Elem()
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)

		if val, exists := field.Tag.Lookup("computed"); exists {
			if val == "true" {
				continue
			}
		}

		if val, exists := field.Tag.Lookup("hcl"); exists {
			hclValue := config[val]

			//TODO Actually check error
			setValue(input, hclValue, field, i)
		}
	}
}