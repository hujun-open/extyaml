package extyaml

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// addSkipTag add SkipTag for the field that has equal value between in and def
// note: exported field pkgpath must be "", while unexported field pkgpath can't be ""
func addSkipTag(in, def reflect.Value) reflect.Type {
	inT := in.Type()
	if inT.Kind() == reflect.Pointer {
		inT = inT.Elem()
		in = in.Elem()
		def = def.Elem()
	}

	var list = []reflect.StructField{}
	for i := 0; i < inT.NumField(); i++ {
		field := inT.Field(i)
		newField := reflect.StructField{
			Name:    field.Name,
			Type:    field.Type,
			PkgPath: inT.PkgPath(),
			Tag:     field.Tag,
			Index:   field.Index,
		}
		//inT is anonymous type like an ebeded struct
		if newField.PkgPath == "" {
			newField.PkgPath = "anonymous"
		}

		if !field.IsExported() {
			//a non-export field
			list = append(list, newField)
			continue
		}

		if _, exists := field.Tag.Lookup(SkipTag); exists {
			//if field has skip tag
			list = append(list, newField)
			continue
		}
		fieldType := field.Type
		inFieldVal := in.Field(i)
		defFieldVal := def.Field(i)
		if field.Type.Kind() == reflect.Pointer {
			if inFieldVal.IsNil() && defFieldVal.IsNil() {
				//both nil
				newField.Tag += reflect.StructTag(fmt.Sprintf(` %v:" "`, SkipTag))
				list = append(list, newField)
				continue
			} else {
				//at least one of them is not nil
				if inFieldVal.IsNil() != defFieldVal.IsNil() {
					//only one of them is nil
					skip := false

					if inFieldVal.IsNil() {
						switch defFieldVal.Elem().Kind() {
						case reflect.Slice:
							if defFieldVal.Elem().Len() == 0 {
								skip = true
							}
						}
					} else {
						switch inFieldVal.Elem().Kind() {
						case reflect.Slice:
							if inFieldVal.Elem().Len() == 0 {
								skip = true
							}
						}
					}
					if !skip {
						newField.PkgPath = ""
					} else {
						newField.Tag += reflect.StructTag(fmt.Sprintf(` %v:" "`, SkipTag))
					}
					list = append(list, newField)
					continue
				}
			}
			//none of them are nil, can't decide yet, de-reference them
			fieldType = fieldType.Elem()
			inFieldVal = inFieldVal.Elem()
			defFieldVal = defFieldVal.Elem()

		}
		//now non-pointer here, even point has been de-referenced
		switch fieldType.Kind() {
		case reflect.Slice, reflect.Map:
			//empty slice or map are consider equal
			if inFieldVal.Len() == 0 && defFieldVal.Len() == 0 {
				newField.Tag += reflect.StructTag(fmt.Sprintf(` %v:" "`, SkipTag))
				list = append(list, newField)
				continue
			}
		}

		if fieldType.Kind() != reflect.Struct || (fieldType.Implements(textMarshalerInt) || field.Type.Implements(yamlMarshalerInt)) {
			// not struct OR struct but implement marshal interface
			if reflect.DeepEqual(inFieldVal.Interface(), defFieldVal.Interface()) {
				//same value
				newField.Tag += reflect.StructTag(fmt.Sprintf(` %v:" "`, SkipTag))
				list = append(list, newField)
				continue
			} else {
				//not equal
				newField.PkgPath = ""
				list = append(list, newField)
				continue
			}

		} else {
			//struct and it doesn't implment marshal interface
			if fieldType.Kind() == reflect.Struct {
				list = append(list, reflect.StructField{
					Name: field.Name,
					Type: addSkipTag(inFieldVal, defFieldVal),
					// PkgPath: inT.PkgPath(),
					Tag:   field.Tag,
					Index: field.Index,
				})
			}
		}
	}
	return reflect.StructOf(list)
}

// MarshalExtDefault marshal in struct into YAML bytes, any field that has same corresponding value as def will be omitted in output.sss
// in and def must be same type of struct
func MarshalExtDefault(in, def any) ([]byte, error) {
	if reflect.TypeOf(in) != reflect.TypeOf(def) {
		return nil, fmt.Errorf("in and def are not same type")
	}
	inT := reflect.TypeOf(in)
	if inT.Kind() == reflect.Pointer {
		inT = inT.Elem()
	}
	if inT.Kind() != reflect.Struct {
		return nil, fmt.Errorf("in and def are not struct")
	}

	inV := reflect.ValueOf(in)
	if inV.Kind() == reflect.Pointer {
		inV = inV.Elem()
	}
	defV := reflect.ValueOf(def)
	if defV.Kind() == reflect.Pointer {
		defV = defV.Elem()
	}
	newT := addSkipTag(inV, defV)
	newType := convertStructType(newT)
	newVal := reflect.New(newType)
	translateStructInline(inV.Interface(), newVal.Interface(), "", true)
	return yaml.Marshal(newVal.Interface())
}
