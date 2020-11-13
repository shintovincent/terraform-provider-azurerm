package example

import (
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
		{
			Name: "top level - everything",
			Input: map[string]interface{}{
				"name": "bingo bango",
				"float": 123.4,
				"number": 123,
				"enabled": false,
				"networks": []interface{}{"network1", "network2", "network3"},
				"networks_set": []interface{}{"networkset1", "networkset2", "networkset3"},
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
				"set": schema.NewSet(FakeHashSchema(),
					[]interface{}{
						map[string]interface{}{
							"name": "setname",
						},
					}),
			},
			Expected: &ExampleObj{
				Name: "bingo bango",
				Float: 123.4,
				Number: 123,
				Enabled: false,
				Networks: []string{"network1", "network2", "network3"},
				NetworksSet: []string{"networkset1", "networkset2", "networkset3"},
				List: []NetworkList{{
					Name: "first",
					Inner: []NetworkInner{{
						Name: "get-a-mac",
					}},
				}},
				Set: []NetworkSet{{
					Name: "setname",
				}},
			},
			ExpectError: false,
		},
		{
			Name: "nests",
			Input: map[string]interface{}{
				"name": "bingo bango",
				"float": 123.4,
				"number": 123,
				"enabled": false,
				"networks": []interface{}{"network1", "network2", "network3"},
				"networks_set": []interface{}{"networkset1", "networkset2", "networkset3"},
				"list": []interface{}{
					map[string]interface{}{
						"name": "first",
						"inner": []interface{}{
							map[string]interface{}{
								"name": "get-a-mac",
								"inner": []interface{}{
									map[string]interface{}{
										"name": "innerinner",
										"should_be_fine": true,
									},
								},
								"set": schema.NewSet(FakeHashSchema(),
									[]interface{}{
										map[string]interface{}{
											"name": "nestedsetname",
										},
									}),
							},
						},
					},
					map[string]interface{}{
						"name": "second",
						"inner": []interface{}{
							map[string]interface{}{
								"name": "get-a-mac2",
								"inner": []interface{}{
									map[string]interface{}{
										"name": "innerinner2",
										"should_be_fine": true,
									},
								},
								"set": schema.NewSet(FakeHashSchema(),
									[]interface{}{
										map[string]interface{}{
											"name": "nestedsetname2",
										},
									}),
							},
						},
					},
				},
				"set": schema.NewSet(FakeHashSchema(),
					[]interface{}{
						map[string]interface{}{
							"name": "setname",
						},
					}),
			},
			Expected: &ExampleObj{
				Name: "bingo bango",
				Float: 123.4,
				Number: 123,
				Enabled: false,
				Networks: []string{"network1", "network2", "network3"},
				NetworksSet: []string{"networkset1", "networkset2", "networkset3"},
				List: []NetworkList{
					{
						Name: "first",
						Inner: []NetworkInner{{
							Name: "get-a-mac",
							Inner: []InnerInner{{
								Name: "innerinner",
								ShouldBeFine: true,
							}},
							Set: []NetworkListSet{{
								Name: "nestedsetname",
							}},
						}},
					},
					{
						Name: "second",
						Inner: []NetworkInner{{
							Name: "get-a-mac2",
							Inner: []InnerInner{{
								Name: "innerinner2",
								ShouldBeFine: true,
							}},
							Set: []NetworkListSet{{
								Name: "nestedsetname2",
							}},
						}},
					},
				},
				Set: []NetworkSet{{
					Name: "setname",
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

func TestEncode(t *testing.T) {
	testCases := []struct{
		Name string
		Input *ExampleObj
		Expected map[string]interface{}
		ExpectError bool
	}{
		{
			Name: "top level - name",
			Input: &ExampleObj{
				Name: "bingo bango",
			},
			Expected: map[string]interface{}{
				"name": "bingo bango",
				"enabled": false,
				"float": float64(0),
				"list": []interface{}{},
				"networks": []string(nil),
				"networks_set": []string(nil),
				"number": int64(0),
				"output": "",
				"set": []interface{}{},
			},
			ExpectError: false,
		},
		{
			Name: "top level - everything",
			Input: &ExampleObj{
				Name: "bingo bango",
				Float: 123.4,
				Number: 123,
				Enabled: false,
			},
			Expected: map[string]interface{}{
				"name": "bingo bango",
				"enabled": false,
				"float": 123.4,
				"list": []interface{}{},
				"networks": []string(nil),
				"networks_set": []string(nil),
				"number": int64(123),
				"output": "",
				"set": []interface{}{},
			},
			ExpectError: false,
		},
		{
			Name: "top level - list",
			Input: &ExampleObj{
				Name: "bingo bango",
				Float: 123.4,
				Number: 123,
				Enabled: false,
				List: []NetworkList{{
					Name: "first",
				}},
			},
			Expected: map[string]interface{}{
				"name": "bingo bango",
				"enabled": false,
				"float": 123.4,
				"list": []interface{}{
					&map[string]interface{}{
						"name": "first",
						"inner": []interface{}{},
					},
				},
				"networks": []string(nil),
				"networks_set": []string(nil),
				"number": int64(123),
				"output": "",
				"set": []interface{}{},
			},
			ExpectError: false,
		},
		{
			Name: "top level - list in lists",
			Input: &ExampleObj{
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
			Expected: map[string]interface{}{
				"name": "bingo bango",
				"enabled": false,
				"float": 123.4,
				"list": []interface{}{
					&map[string]interface{}{
						"name": "first",
						"inner": []interface{}{
							&map[string]interface{}{
								"name": "get-a-mac",
								"inner": []interface{}{},
								"set": []interface{}{},
							},
						},
					},
				},
				"networks": []string(nil),
				"networks_set": []string(nil),
				"number": int64(123),
				"output": "",
				"set": []interface{}{},
			},
			ExpectError: false,
		},
		{
			Name: "top level - everything",
			Input: &ExampleObj{
				Name: "bingo bango",
				Float: 123.4,
				Number: 123,
				Enabled: false,
				Networks: []string{"network1", "network2", "network3"},
				NetworksSet: []string{"networkset1", "networkset2", "networkset3"},
				List: []NetworkList{{
					Name: "first",
					Inner: []NetworkInner{{
						Name: "get-a-mac",
					}},
				}},
				Set: []NetworkSet{{
					Name: "setname",
				}},
			},
			Expected: map[string]interface{}{
				"name": "bingo bango",
				"enabled": false,
				"float": 123.4,
				"list": []interface{}{
					&map[string]interface{}{
						"name": "first",
						"inner": []interface{}{
							&map[string]interface{}{
								"name": "get-a-mac",
								"inner": []interface{}{},
								"set": []interface{}{},
							},
						},
					},
				},
				"networks": []string{"network1", "network2", "network3"},
				"networks_set": []string{"networkset1", "networkset2", "networkset3"},
				"number": int64(123),
				"output": "",
				"set": []interface{}{
					&map[string]interface{}{
						"name": "setname",
						"inner": []interface{}{},
					},
				},
			},
			ExpectError: false,
		},
		{
			Name: "nests",
			Input: &ExampleObj{
				Name:        "bingo bango",
				Float:       123.4,
				Number:      123,
				Enabled:     false,
				List: []NetworkList{
					{
						Name: "first",
						Inner: []NetworkInner{{
							Name: "get-a-mac",
							Inner: []InnerInner{{
								Name:         "innerinner",
								ShouldBeFine: true,
							}},
							Set: []NetworkListSet{{
								Name: "nestedsetname",
							}},
						}},
					},
					{
						Name: "second",
						Inner: []NetworkInner{{
							Name: "get-a-mac2",
							Inner: []InnerInner{{
								Name:         "innerinner2",
								ShouldBeFine: true,
							}},
							Set: []NetworkListSet{{
								Name: "nestedsetname2",
							}},
						}},
					},
				},
				Set: []NetworkSet{{
					Name: "setname",
				}},
			},
			Expected: map[string]interface{}{
				"name": "bingo bango",
				"enabled": false,
				"float": 123.4,
				"list": []interface{}{
					&map[string]interface{}{
						"name": "first",
						"inner": []interface{}{
							&map[string]interface{}{
								"name": "get-a-mac",
								"inner": []interface{}{
									&map[string]interface{}{
										"name": "innerinner",
										"should_be_fine": true,
									},
								},
								"set": []interface{}{
									&map[string]interface{}{
										"name": "nestedsetname",
									},
								},
							},
						},
					},
					&map[string]interface{}{
						"name": "second",
						"inner": []interface{}{
							&map[string]interface{}{
								"name": "get-a-mac2",
								"inner": []interface{}{
									&map[string]interface{}{
										"name": "innerinner2",
										"should_be_fine": true,
									},
								},
								"set": []interface{}{
									&map[string]interface{}{
										"name": "nestedsetname2",
									},
								},
							},
						},
					},
				},

				"networks": []string(nil),
				"networks_set": []string(nil),
				"number": int64(123),
				"output": "",
				"set": []interface{}{
					&map[string]interface{}{
						"name": "setname",
						"inner": []interface{}{},
					},
				},
			},
			ExpectError: false,
		},
	}
	for _, v := range testCases {
		output, err := encodeHelper(v.Input)
		if err != nil {
			t.Fatalf("encoding error: %+v", err)
		}

		if !cmp.Equal(output, v.Expected) {
			t.Fatalf("Test Failed %q: output mismatch\n\n Expected: %+v\n\n Received: %+v\n\n", v.Name, v.Expected, output)
		}
	}
}

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

func encodeHelper(input interface{}) (map[string]interface{}, error) {
	objType := reflect.TypeOf(input).Elem()
	objVal := reflect.ValueOf(input).Elem()

	serialized,err := recurse(objType, objVal)
	if err != nil {
		return nil, err
	}
	return *serialized, nil
}

// FakeHashSchema is a dummy method to return a fake hash so we can build a Set for testing purposes.
func FakeHashSchema() schema.SchemaSetFunc {
	return func(v interface{}) int {
		return 0
	}
}