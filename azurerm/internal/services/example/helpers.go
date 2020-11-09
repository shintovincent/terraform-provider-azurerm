package example

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/resourceid"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
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
		log.Printf("[MATTHEWMATTHEW] Field", field)

		if val, exists := field.Tag.Lookup("computed"); exists {
			if val == "true" {
				continue
			}
		}

		if val, exists := field.Tag.Lookup("hcl"); exists {
			hclValue := rmd.ResourceData.Get(val)

			log.Printf("[MATTHEWMATTHEW] HCLValue: ", hclValue)
			log.Printf("[MATTHEWMATTHEW] Input Type: ", reflect.ValueOf(input).Elem().Field(i).Type())

			setValue(input, hclValue, field, i)

/*
			if v, ok := hclValue.(string); ok {
				reflect.ValueOf(input).Elem().Field(i).SetString(v)
				continue
			}
			if v, ok := hclValue.(int); ok {
				reflect.ValueOf(input).Elem().Field(i).SetInt(int64(v))
				continue
			}

			// Doesn't work for empty bools?
			if v, ok := hclValue.(bool); ok {
				log.Printf("[BOOL] Decode %+v", v)

				reflect.ValueOf(input).Elem().Field(i).SetBool(v)
				continue
			}

			if v, ok := hclValue.(*schema.Set); ok {
				switch fieldType := reflect.ValueOf(input).Elem().Field(i).Type(); fieldType {
				// TODO do I have to do it this way for the rest of the types?
				case reflect.TypeOf([]string{}):
					list := v.List()
					log.Printf("[MATTHEWMATTHEW] Sets!!: ", reflect.TypeOf([]string{}).Kind())
					stringSlice := reflect.MakeSlice(reflect.TypeOf([]string{}), len(list), len(list))
					for j, stringVal := range list {
						stringSlice.Index(j).SetString(stringVal.(string))
					}
					log.Printf("[MATTHEWMATTHEW] Set StringSlice ", stringSlice)
					reflect.ValueOf(input).Elem().Field(i).Set(stringSlice)
					continue
				default:
					return fmt.Errorf("unknown type %+v for key %q", field.Type.Kind(), fieldType)
				}
			}

			if v, ok := hclValue.([]interface{}); ok {
				switch fieldType := reflect.ValueOf(input).Elem().Field(i).Type(); fieldType {
				// TODO do I have to do it this way for the rest of the types?
				case reflect.TypeOf([]string{}):
					log.Printf("[MATTHEWMATTHEW] Lists!: ", reflect.TypeOf([]string{}).Kind())
					stringSlice := reflect.MakeSlice(reflect.TypeOf([]string{}), len(v), len(v))
					for i, stringVal := range v {
						stringSlice.Index(i).SetString(stringVal.(string))
					}
					log.Printf("[MATTHEWMATTHEW] List StringSlice ", stringSlice)
					reflect.ValueOf(input).Elem().Field(i).Set(stringSlice)
					continue
				default:
					//  []example.NetworkList
					arrayList := reflect.New(fieldType.Elem())
					log.Printf("[MATTHEWMATTHEW] List", arrayList.Interface())
					// log.Printf("[MATTHEWMATTHEW] Info ", reflect.TypeOf(obj).Elem().NumField())

					elem := reflect.New(fieldType.Elem())
					log.Printf("[MATTHEWMATTHEW] element ", elem.Interface())

					for j := 0; j < elem.Type().Elem().NumField(); j ++ {
						log.Printf("[MATTHEWMATTHEW] nestedField ", elem.Type().Elem().Field(j).Name)
						log.Printf("[MATTHEWMATTHEW] nestedFieldType ", elem.Type().Elem().Field(j).Type)
					}


						}
					}


					// field, ok := reflect.TypeOf(fieldType).Elem().FieldByName("name")

					return fmt.Errorf("unknown type %+v for key %q", field.Type.Kind(), fieldType)
				}
			} */

			// TODO: other types
		}
	}
	return nil
}

func setValue(input, hclValue interface{}, field reflect.StructField, index int) error{
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
		switch fieldType := reflect.ValueOf(input).Elem().Field(index).Type(); fieldType {
		// TODO do I have to do it this way for the rest of the types?
		case reflect.TypeOf([]string{}):
			list := v.List()
			log.Printf("[MATTHEWMATTHEW] Sets!!: ", reflect.TypeOf([]string{}).Kind())
			stringSlice := reflect.MakeSlice(reflect.TypeOf([]string{}), len(list), len(list))
			for j, stringVal := range list {
				stringSlice.Index(j).SetString(stringVal.(string))
			}
			log.Printf("[MATTHEWMATTHEW] Set StringSlice ", stringSlice)
			reflect.ValueOf(input).Elem().Field(index).Set(stringSlice)
			return nil
		default:
			list := v.List()
			arrayList := reflect.New(reflect.ValueOf(input).Elem().Field(index).Type())
			log.Printf("[MATTHEWMATTHEW] List Type", arrayList.Type())

			for _, mapVal := range list {
				if test := mapVal.(map[string]interface{}); test != nil {
					elem := reflect.New(fieldType.Elem())
					log.Printf("[MATTHEWMATTHEW] element ", elem)
					for j := 0; j < elem.Type().Elem().NumField(); j++ {
						nestedField := elem.Type().Elem().Field(j)
						log.Printf("[MATTHEWMATTHEW] nestedField ", nestedField)
						if val, exists := nestedField.Tag.Lookup("computed"); exists {
							if val == "true" {
								continue
							}
						}

						if val, exists := nestedField.Tag.Lookup("hcl"); exists {
							nestedHCLValue := test[val]
							log.Printf("[MATTHEWMATTHEW] HCLValue: ", nestedHCLValue)
							setValue(elem.Interface(), nestedHCLValue, nestedField, j)
						}
					}
					if !arrayList.CanSet() {
						log.Printf("list can set before", arrayList.CanSet())
						arrayList = arrayList.Elem()
						log.Printf("list can set after", arrayList.CanSet())
					}

					if !elem.CanSet() {
						log.Printf("elem can set before", elem.CanSet())
						elem = elem.Elem()
						log.Printf("elem can set after", elem.CanSet())
					}

					arrayList = reflect.Append(arrayList, elem)
				}
			}
			log.Printf("[Set] Setting list: ", arrayList)
			reflect.ValueOf(input).Elem().Field(index).Set(arrayList)

			return fmt.Errorf("unknown type %+v for key %q", field.Type.Kind(), fieldType)
		}
	}

	if v, ok := hclValue.([]interface{}); ok {
		switch fieldType := reflect.ValueOf(input).Elem().Field(index).Type(); fieldType {
		// TODO do I have to do it this way for the rest of the types?
		case reflect.TypeOf([]string{}):
			log.Printf("[MATTHEWMATTHEW] Lists!: ", reflect.TypeOf([]string{}).Kind())
			stringSlice := reflect.MakeSlice(reflect.TypeOf([]string{}), len(v), len(v))
			for i, stringVal := range v {
				stringSlice.Index(i).SetString(stringVal.(string))
			}
			log.Printf("[MATTHEWMATTHEW] List StringSlice ", stringSlice)
			reflect.ValueOf(input).Elem().Field(index).Set(stringSlice)
			return nil
		default:
			//  []example.NetworkList
			arrayList := reflect.New(reflect.ValueOf(input).Elem().Field(index).Type())
			log.Printf("[MATTHEWMATTHEW] List Type", arrayList.Type())

			for _, mapVal := range v {
				if test := mapVal.(map[string]interface{}); test != nil {
					elem := reflect.New(fieldType.Elem())
					log.Printf("[MATTHEWMATTHEW] element ", elem)
					for j := 0; j < elem.Type().Elem().NumField(); j++ {
						nestedField := elem.Type().Elem().Field(j)
						log.Printf("[MATTHEWMATTHEW] nestedField ", nestedField)
						if val, exists := nestedField.Tag.Lookup("computed"); exists {
							if val == "true" {
								continue
							}
						}

						if val, exists := nestedField.Tag.Lookup("hcl"); exists {
							nestedHCLValue := test[val]
							log.Printf("[MATTHEWMATTHEW] HCLValue: ", nestedHCLValue)
							setValue(elem.Interface(), nestedHCLValue, nestedField, j)
						}
					}
					if !arrayList.CanSet() {
						log.Printf("list can set before", arrayList.CanSet())
						arrayList = arrayList.Elem()
						log.Printf("list can set after", arrayList.CanSet())
					}

					if !elem.CanSet() {
						log.Printf("elem can set before", elem.CanSet())
						elem = elem.Elem()
						log.Printf("elem can set after", elem.CanSet())
					}

					arrayList = reflect.Append(arrayList, elem)
				}
			}
			log.Printf("[List] Setting list: ", arrayList)
			reflect.ValueOf(input).Elem().Field(index).Set(arrayList)


			return nil // fmt.Errorf("unknown type %+v for key %q", field.Type.Kind(), fieldType)
		}
	}


	return nil
}

func (rmd *ResourceMetaData) Encode(input interface{}) error {
	objType := reflect.TypeOf(input).Elem()
	objVal := reflect.ValueOf(input).Elem()
	for i := 0; i < objType.NumField(); i++ {
		field := objType.Field(i)
		fieldVal := objVal.Field(i)
		if hclTag, exists := field.Tag.Lookup("hcl"); exists {
			// TODO: make this better
			switch field.Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				iv := fieldVal.Int()
				log.Printf("[TOMTOM] Setting %q to %d", hclTag, iv)

				if err := rmd.ResourceData.Set(hclTag, iv); err != nil {
					return err
				}

			case reflect.Float32, reflect.Float64:
				fv := fieldVal.Float()
				log.Printf("[TOMTOM] Setting %q to %d", hclTag, fv)

				if err := rmd.ResourceData.Set(hclTag, fv); err != nil {
					return err
				}

			case reflect.String:
				sv := fieldVal.String()
				log.Printf("[TOMTOM] Setting %q to %q", hclTag, sv)
				if err := rmd.ResourceData.Set(hclTag, sv); err != nil {
					return err
				}

			case reflect.Bool:
				bv := fieldVal.Bool()
				log.Printf("[BOOL] Setting %q to %q", hclTag, bv)
				if err := rmd.ResourceData.Set(hclTag, bv); err != nil {
					return err
				}

			case reflect.Slice:
				sv := fieldVal.Slice(0, fieldVal.Len())
				attr := make([]interface{}, sv.Len())
				for i := 0; i < sv.Len(); i++ {
					log.Printf("[SLICE] Index %d is %q", i, sv.Index(i).Interface())
					log.Printf("[SLICE] Type %d is %q", i, sv.Index(i))
					var recurse = func(input reflect.Value) map[string]interface{} {

					}

					serialized := recurse(sv.Index(i))
					attr[i] = serialized
				}
				log.Printf("[SLICE] Setting %q to %+v", hclTag, attr)

				if err := rmd.ResourceData.Set(hclTag, attr); err != nil {
					return err
				}
			default:
				// TODO take this back
				return fmt.Errorf("unknown type %+v for key %q", field.Type.Kind(), hclTag)
			}
		}
	}
	return nil
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
