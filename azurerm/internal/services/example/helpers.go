package example

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/resourceid"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"log"
	"reflect"
	"time"
)


type Logger interface {
	Info(message string)
	InfoF(format string, args ...interface{})
	Warn(message string)
	WarnF(format string, args ...interface{})
}

type ResourceRunFunc func(ctx context.Context, metadata ResourceMetaData) error

type ResourceFunc struct {
	Func    ResourceRunFunc
	Timeout time.Duration
}

type ResourceMetaData struct {
	Client       *clients.Client
	Logger       Logger
	ResourceData *schema.ResourceData
}

func (rmd ResourceMetaData) Decode(input interface{}) error {
	objType := reflect.TypeOf(input).Elem()
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		log.Print("[MATTHEWMATTHEW] Field", field)

		if val, exists := field.Tag.Lookup("computed"); exists {
			if val == "true" {
				continue
			}
		}

		if val, exists := field.Tag.Lookup("hcl"); exists {
			hclValue := rmd.ResourceData.Get(val)

			log.Print("[MATTHEWMATTHEW] HCLValue: ", hclValue)
			log.Print("[MATTHEWMATTHEW] Input Type: ", reflect.ValueOf(input).Elem().Field(i).Type())

			if err := setValue(input, hclValue, i); err != nil {
				return err
			}
		}
	}
	return nil
}

func setValue(input, hclValue interface{}, index int) error {
	if v, ok := hclValue.(string); ok {
		log.Printf("[String] Decode %+v", v)
		log.Printf("Input %+v", reflect.ValueOf(input))
		log.Printf("Input Elem %+v", reflect.ValueOf(input).Elem())
		reflect.ValueOf(input).Elem().Field(index).SetString(v)
		return nil
	}

	if v, ok := hclValue.(int); ok {
		log.Printf("[INT] Decode %+v", v)
		reflect.ValueOf(input).Elem().Field(index).SetInt(int64(v))
		return nil
	}

	if v, ok := hclValue.(float64); ok {
		log.Printf("[Float] Decode %+v", v)
		reflect.ValueOf(input).Elem().Field(index).SetFloat(v)
		return nil
	}

	// Doesn't work for empty bools?
	if v, ok := hclValue.(bool); ok {
		log.Printf("[BOOL] Decode %+v", v)

		reflect.ValueOf(input).Elem().Field(index).SetBool(v)
		return nil
	}

	if v, ok := hclValue.(*schema.Set); ok {
		setListValue(input, index, v.List())
		return nil
	}

	if mapConfig, ok := hclValue.(map[string]interface{}); ok {

		mapOutput := reflect.MakeMap(reflect.TypeOf(map[string]string{}))
		for key, val := range mapConfig {
			mapOutput.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(val))
		}

		reflect.ValueOf(input).Elem().Field(index).Set(mapOutput)
		return nil
	}

	if v, ok := hclValue.([]interface{}); ok {
		setListValue(input, index, v)
		return nil
	}

	return nil
}

func setListValue(input interface{}, index int, v []interface{}) {
	switch fieldType := reflect.ValueOf(input).Elem().Field(index).Type(); fieldType {
	// TODO do I have to do it this way for the rest of the types?
	case reflect.TypeOf([]string{}):
		stringSlice := reflect.MakeSlice(reflect.TypeOf([]string{}), len(v), len(v))
		for i, stringVal := range v {
			stringSlice.Index(i).SetString(stringVal.(string))
		}
		reflect.ValueOf(input).Elem().Field(index).Set(stringSlice)

	case reflect.TypeOf([]int{}):
		iSlice := reflect.MakeSlice(reflect.TypeOf([]int{}), len(v), len(v))
		for i, iVal := range v {
			iSlice.Index(i).SetInt(int64(iVal.(int)))
		}
		reflect.ValueOf(input).Elem().Field(index).Set(iSlice)

	case reflect.TypeOf([]float64{}):
		fSlice := reflect.MakeSlice(reflect.TypeOf([]float64{}), len(v), len(v))
		for i, fVal := range v {
			fSlice.Index(i).SetFloat(fVal.(float64))
		}
		reflect.ValueOf(input).Elem().Field(index).Set(fSlice)

	case reflect.TypeOf([]bool{}):
		bSlice := reflect.MakeSlice(reflect.TypeOf([]bool{}), len(v), len(v))
		for i, bVal := range v {
			bSlice.Index(i).SetBool(bVal.(bool))
		}
		reflect.ValueOf(input).Elem().Field(index).Set(bSlice)

	default:
		valueToSet := reflect.New(reflect.ValueOf(input).Elem().Field(index).Type())
		log.Print("[MATTHEWMATTHEW] List Type", valueToSet.Type())

		for _, mapVal := range v {
			if test, ok := mapVal.(map[string]interface{}); ok && test != nil {
				elem := reflect.New(fieldType.Elem())
				log.Print("[MATTHEWMATTHEW] element ", elem)
				for j := 0; j < elem.Type().Elem().NumField(); j++ {
					nestedField := elem.Type().Elem().Field(j)
					log.Print("[MATTHEWMATTHEW] nestedField ", nestedField)
					if val, exists := nestedField.Tag.Lookup("computed"); exists {
						if val == "true" {
							continue
						}
					}

					if val, exists := nestedField.Tag.Lookup("hcl"); exists {
						nestedHCLValue := test[val]
						setValue(elem.Interface(), nestedHCLValue, j)
					}
				}

				if !elem.CanSet() {
					elem = elem.Elem()
				}

				if valueToSet.Kind() == reflect.Ptr {
					valueToSet.Elem().Set(reflect.Append(valueToSet.Elem(), elem))
				} else {
					valueToSet = reflect.Append(valueToSet, elem)
				}

				log.Print("value to set type after changes", valueToSet.Type())
			}
		}
		fieldToSet := reflect.ValueOf(input).Elem().Field(index)

		if valueToSet.Kind() != reflect.Ptr {
			fieldToSet.Set(valueToSet)
		} else {
			fieldToSet.Set(valueToSet.Elem())
		}
	}
}

func (rmd *ResourceMetaData) Encode(input interface{}) error {
	objType := reflect.TypeOf(input).Elem()
	objVal := reflect.ValueOf(input).Elem()

	serialized,err := recurse(objType, objVal)
	if err != nil {
		return err
	}

	for k, v := range *serialized {
		if err := rmd.ResourceData.Set(k, v); err != nil {
			return fmt.Errorf("settting %q: %+v", k, err)
		}
	}
	return nil
}

func recurse(objType reflect.Type, objVal reflect.Value) (*map[string]interface{}, error) {
	output := make(map[string]interface{}, 0)
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		fieldVal := objVal.Field(i)
		if hclTag, exists := field.Tag.Lookup("hcl"); exists {
			switch field.Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				iv := fieldVal.Int()
				log.Printf("[TOMTOM] Setting %q to %d", hclTag, iv)

				output[hclTag] = iv

			case reflect.Float32, reflect.Float64:
				fv := fieldVal.Float()
				log.Printf("[TOMTOM] Setting %q to %f", hclTag, fv)

				output[hclTag] = fv

			case reflect.String:
				sv := fieldVal.String()
				log.Printf("[TOMTOM] Setting %q to %q", hclTag, sv)
				output[hclTag] = sv

			case reflect.Bool:
				bv := fieldVal.Bool()
				log.Printf("[BOOL] Setting %q to %t", hclTag, bv)
				output[hclTag] = bv

			case reflect.Map:
				iter := fieldVal.MapRange()
				attr := make(map[string]interface{})
				for iter.Next() {
					attr[iter.Key().String()] = iter.Value().Interface()
				}
				output[hclTag] = attr

			case reflect.Slice:
				sv := fieldVal.Slice(0, fieldVal.Len())
				attr := make([]interface{}, sv.Len())
				switch sv.Type() {
				case reflect.TypeOf([]string{}), reflect.TypeOf([]int{}), reflect.TypeOf([]float64{}), reflect.TypeOf([]bool{}):
					log.Printf("[SLICE] Setting %q to %q", hclTag, sv)
					output[hclTag] = sv.Interface()

				default:
					for i := 0; i < sv.Len(); i++ {
						log.Printf("[SLICE] Index %d is %q", i, sv.Index(i).Interface())
						log.Printf("[SLICE] Type %+v", sv.Type())
						nestedType := sv.Index(i).Type()
						nestedValue :=sv.Index(i)
						serialized, err := recurse(nestedType, nestedValue)
						if err != nil {
							panic(err)
						}
						attr[i] = serialized
					}
					log.Printf("[SLICE] Setting %q to %+v", hclTag, attr)
					output[hclTag] = attr
				}
			default:
				return &output, fmt.Errorf("unknown type %+v for key %q", field.Type.Kind(), hclTag)
			}
		}
	}

	return &output, nil
}

func (rmd ResourceMetaData) SetID(formatter resourceid.Formatter) {
	subscriptionId := rmd.Client.Account.SubscriptionId
	rmd.ResourceData.SetId(formatter.ID(subscriptionId))
}

type Resource interface {
	Arguments() map[string]*schema.Schema
	Attributes() map[string]*schema.Schema

	ResourceType() string

	Create() ResourceFunc
	Read() ResourceFunc
	Delete() ResourceFunc
	IDValidationFunc() schema.SchemaValidateFunc
}

type ResourceWithUpdate interface {
	Update() ResourceFunc
}

type ResourceWrapper struct {
	resource Resource
}

func NewResourceWrapper(resource Resource) ResourceWrapper {
	return ResourceWrapper{
		resource: resource,
	}
}

func (rw ResourceWrapper) Resource() (*schema.Resource, error) {
	resourceSchema, err := rw.schema()
	if err != nil {
		return nil, fmt.Errorf("building Schema: %+v", err)
	}

	var d = func(duration time.Duration) *time.Duration {
		return &duration
	}
	logger := ExampleLogger{}

	resource := schema.Resource{
		Schema: *resourceSchema,

		Create: func(d *schema.ResourceData, meta interface{}) error {
			ctx, metaData := rw.runArgs(d, meta, logger)
			err := rw.resource.Create().Func(ctx, metaData)
			if err != nil {
				return err
			}
			return rw.resource.Read().Func(ctx, metaData)
		},

		// looks like these could be reused, easiest if they're not
		Read: func(d *schema.ResourceData, meta interface{}) error {
			ctx, metaData := rw.runArgs(d, meta, logger)
			return rw.resource.Read().Func(ctx, metaData)
		},
		Delete: func(d *schema.ResourceData, meta interface{}) error {
			ctx, metaData := rw.runArgs(d, meta, logger)
			return rw.resource.Delete().Func(ctx, metaData)
		},

		Timeouts: &schema.ResourceTimeout{
			Create: d(rw.resource.Create().Timeout),
			Read:   d(rw.resource.Read().Timeout),
			Delete: d(rw.resource.Delete().Timeout),
		},
		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			fn := rw.resource.IDValidationFunc()
			warnings, errors := fn(id, "id")
			if len(warnings) > 0 {
				for _, warning := range warnings {
					logger.Warn(warning)
				}
			}
			if len(errors) > 0 {
				out := ""
				for _, error := range errors {
					out += error.Error()
				}
				return fmt.Errorf(out)
			}

			return err
		}),
	}

	if v, ok := rw.resource.(ResourceWithUpdate); ok {
		resource.Update = func(d *schema.ResourceData, meta interface{}) error {
			ctx, metaData := rw.runArgs(d, meta, logger)
			err := v.Update().Func(ctx, metaData)
			if err != nil {
				return err
			}
			return rw.resource.Read().Func(ctx, metaData)
		}
		resource.Timeouts.Update = d(v.Update().Timeout)
	}

	return &resource, nil
}

func (rw ResourceWrapper) run(in ResourceFunc, logger Logger) func(d *schema.ResourceData, meta interface{}) error {
	return func(d *schema.ResourceData, meta interface{}) error {
		ctx, metaData := rw.runArgs(d, meta, logger)
		err := in.Func(ctx, metaData)
		// TODO: ensure the logger is drained/processed
		return err
	}
}

func (rw ResourceWrapper) schema() (*map[string]*schema.Schema, error) {
	out := make(map[string]*schema.Schema, 0)

	for k, v := range rw.resource.Arguments() {
		if _, alreadyExists := out[k]; alreadyExists {
			return nil, fmt.Errorf("%q already exists in the schema", k)
		}

		// TODO: if readonly

		out[k] = v
	}

	for k, v := range rw.resource.Attributes() {
		if _, alreadyExists := out[k]; alreadyExists {
			return nil, fmt.Errorf("%q already exists in the schema", k)
		}

		// TODO: if editable

		// every attribute has to be computed
		v.Computed = true
		out[k] = v
	}

	return &out, nil
}

func (rw ResourceWrapper) runArgs(d *schema.ResourceData, meta interface{}, logger Logger) (context.Context, ResourceMetaData) {
	ctx := meta.(*clients.Client).StopContext
	client := meta.(*clients.Client)
	metaData := ResourceMetaData{
		Client:       client,
		Logger:       logger,
		ResourceData: d,
	}

	return ctx, metaData
}
