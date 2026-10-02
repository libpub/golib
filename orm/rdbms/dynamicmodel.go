package rdbms

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/libpub/golib/logger"
	"github.com/libpub/golib/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type TableInterface interface {
	TableName() string
	SetTableName(tableName string)
}

// CreateDynamicModel Dynamically create a struct by defination structure
func CreateDynamicModel(definition DynamicModelDefination) interface{} {
	// 创建一个包含默认主键的动态结构体
	// type dynamicStruct struct {
	// 	DynamicModel
	// }
	// dynamicModel := dynamicStruct{}
	// dynamicModel.customizedTableName = definition.TableName
	// dynamicModelValue := reflect.ValueOf(&dynamicModel)

	// dynamic struct fields
	fields := []reflect.StructField{}
	for _, col := range definition.Columns {
		colName := col.Name
		colType := col.Type

		field := reflect.StructField{
			Name: cases.Title(language.English).String(colName),
			Type: getFieldType(colType),
		}

		// Add field tag
		tag := fmt.Sprintf(`xorm:"'%s' %s"`, colName, col.TagContent())
		field.Tag = reflect.StructTag(tag)

		// add field struct to field list
		fields = append(fields, field)
	}

	// Embed DynamicModel in the new struct type
	fields = append(fields, reflect.StructField{
		Name: "DynamicModel",
		Type: reflect.TypeOf(DynamicModel{}),
	})
	// tableNameMethod := func(args []reflect.Value) []reflect.Value {
	// 	return []reflect.Value{reflect.ValueOf(definition.TableName)}
	// }
	// methodField := reflect.StructField{
	// 	Name: "TableName",
	// 	// Type: reflect.FuncOf([]reflect.Type{}, []reflect.Type{reflect.TypeOf("")}, false),
	// 	Type: reflect.TypeOf(func() string { return "" }),
	// }
	// fields = append(fields, methodField)
	myStructType := reflect.StructOf(fields)
	// Create an instance of the struct
	instance := reflect.New(myStructType).Elem()
	instance.FieldByName("DynamicModel").Addr().Interface().(*DynamicModel).SetTableName(definition.TableName)
	// instance.Elem().FieldByName("TableName").Set(reflect.MakeFunc(fields[len(fields)-1].Type, tableNameMethod))
	// instance.Elem().MethodByName("SetTableName").Call([]reflect.Value{reflect.ValueOf(definition.TableName)})

	// Dynamically add a method
	// vals := instance.Elem().FieldByName("DynamicModel").MethodByName("TableName").Call([]reflect.Value{})
	fmt.Println("test TableName", instance.FieldByName("DynamicModel").Addr().Interface().(*DynamicModel).TableName())
	// addMethod(&instance, "TableName", tableNameMethod)

	// // 添加一个方法
	// method := reflect.Method{
	// 	Name: "TableName",
	// 	Type: reflect.FuncOf([]reflect.Type{reflect.TypeOf("")}, []reflect.Type{}),
	// 	Func: reflect.MakeFunc(myStructType, func(args []reflect.Value) []reflect.Value {
	// 		// 在这里定义 String 方法的实现
	// 		return []reflect.Value{reflect.ValueOf(definition.TableName)}
	// 	}),
	// }
	// // 将方法添加到结构体类型
	// methods := []reflect.Method{method}
	// myStructType = reflect.StructOf(fields)
	// myStructType = reflect.New(myStructType).Type()
	// myStructType = reflect.InterfaceType(myStructType, methods)
	// inst2 := instance.Elem().Interface()
	// fmt.Println("Dynamic model", inst2)
	// inst, ok := instance.Interface().(TableInterface)
	// if ok {
	// 	fmt.Printf("1 TableName:%s\n", inst.TableName())
	// } else {
	// 	inst, ok = instance.Elem().Interface().(TableInterface)
	// 	if ok {
	// 		fmt.Printf("2 TableName:%s\n", inst.TableName())
	// 	}
	// }
	// // 创建结构体实例
	// return reflect.New(reflect.TypeOf(dynamicStruct{})).Elem().Interface()
	return instance.Interface()
}

// ParseDynamicModelFromJSON parsing dynamic model defination
func ParseDynamicModelFromJSON(jsonDefinition string) (DynamicModelDefination, error) {
	modelDefination := DynamicModelDefination{}
	err := json.Unmarshal([]byte(jsonDefinition), &modelDefination)
	if err != nil {
		logger.Error.Printf("Parsing dynamic model from json failed with error:%v", err)
	}
	return modelDefination, err
}

// getFieldType 根据字符串类型返回反射类型
func getFieldType(fieldType string) reflect.Type {
	typePattern, _ := regexp.Compile(`\w+`)
	typeStr := strings.ToLower(typePattern.FindString(fieldType))
	switch typeStr {
	case "int":
		return reflect.TypeOf(0)
	case "bigint":
		return reflect.TypeOf(int64(0))
	case "varchar", "char", "text":
		return reflect.TypeOf("")
	case "decimal", "float64":
		return reflect.TypeOf(0.1)
	case "float", "float32":
		return reflect.TypeOf(float32(0.1))
	case "double":
		return reflect.TypeOf(float64(0.1))
	case "blob":
		return reflect.TypeOf([]uint8{})
	case "bool":
		return reflect.TypeOf(false)
	case "datetime":
		return reflect.TypeOf(time.Now())
	default:
		return reflect.TypeOf(interface{}(nil))
	}
}

// // addMethod dynamically adds a method to the struct
// func addMethod(instance *reflect.Value, methodName string, methodFunc func(args []reflect.Value) (results []reflect.Value)) {
// 	// Create an anonymous field containing the method
// 	methodField := reflect.FuncOf([]reflect.Type{}, []reflect.Type{reflect.TypeOf("")}, false)
// 	method := reflect.New(methodField).Elem()
// 	method.Set(reflect.MakeFunc(methodField, methodFunc))

// 	// Add the anonymous field to the struct
// 	*instance = reflect.Append(*instance, method)

// 	// Set the anonymous field as the new method
// 	instanceType := instance.Type()
// 	if instanceType.NumField() > 0 {
// 		instanceType.Field(instanceType.NumField() - 1).Name = methodName
// 	}
// }

// addMethod dynamically adds a method to the struct
func addMethod0(instance reflect.Value, methodName string, methodFunc func(args []reflect.Value) (results []reflect.Value)) {
	// Get the type of the struct
	typ := instance.Type()

	// Create an anonymous field containing the method
	methodField := reflect.FuncOf([]reflect.Type{}, []reflect.Type{reflect.TypeOf("")}, false)
	method := reflect.New(methodField).Elem()
	method.Set(reflect.MakeFunc(methodField, methodFunc))

	// Add the anonymous field to the struct
	numFields := typ.NumField()
	fields := make([]reflect.StructField, numFields+1)
	for i := 0; i < numFields; i++ {
		fields[i] = typ.Field(i)
	}
	fields[numFields] = reflect.StructField{
		Name: methodName,
		Type: methodField,
	}
	typ = reflect.StructOf(fields)

	// Create a new instance of the struct
	newInstance := reflect.New(typ).Elem()

	// Copy the values of the original struct's fields to the new instance
	for i := 0; i < numFields; i++ {
		newInstance.Field(i).Set(instance.Field(i))
	}

	// Set the anonymous field as the new method
	newInstance.Field(numFields).Set(method)

	// Update the original struct instance
	instance.Set(newInstance)
}

func (c *ColumnDefination) TagContent() string {
	tagContents := []string{}
	if c.PrimaryKey {
		tagContents = append(tagContents, "pk")
	}
	if c.NotNull {
		tagContents = append(tagContents, "notnull")
	}
	if c.AutoIncrement {
		tagContents = append(tagContents, "autoincr")
	}
	if c.Unique {
		tagContents = append(tagContents, "unique")
	}
	if c.Index {
		tagContents = append(tagContents, "index")
	}
	if c.Default != nil {
		switch c.Default.(type) {
		case string:
			tagContents = append(tagContents, fmt.Sprintf("default('%s')", strings.ReplaceAll(c.Default.(string), "'", "\\'")))
		default:
			tagContents = append(tagContents, fmt.Sprintf("default(%s)", utils.ToString(c.Default)))
		}
	}
	if c.Comment != "" {
		tagContents = append(tagContents, fmt.Sprintf("comment('%s')", strings.ReplaceAll(c.Comment, "'", "\\'")))
	}
	return strings.Join(tagContents, " ")
}
