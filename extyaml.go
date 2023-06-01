package extyaml

import (
	"encoding"
	"fmt"
	"log"
	"reflect"

	"gopkg.in/yaml.v3"
)

type extType interface {
	init()
	//setVal own value to from
	setVal(from any)
}

type toOrigType interface {
	//return value of orig type
	toOrig() any
}

func convertStructType(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Pointer {
		//a real pointer, not pointer like type like slice
		t = t.Elem()
	}
	if RegisteredTypes.isSupportedType(t, true) {
		//input is a supported type
		return RegisteredTypes.Get(GetTypeName(t)).exType
	}
	//check if the input has supported marshaling method
	if t.Implements(textMarshalerInt) || t.Implements(yamlMarshalerInt) {

		return t
	}

	switch t.Kind() {
	case reflect.Array:
		return reflect.ArrayOf(t.Len(), convertStructType(t.Elem()))
	case reflect.Slice:
		return reflect.SliceOf(convertStructType(t.Elem()))
	case reflect.Map:
		return reflect.MapOf(convertStructType(t.Key()), convertStructType(t.Elem()))
	case reflect.Struct:
		var list = []reflect.StructField{}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				//a non-export field
				list = append(list, field)
				continue
			}
			list = append(list, reflect.StructField{
				Name:    field.Name,
				Type:    convertStructType(field.Type),
				PkgPath: field.PkgPath,
				Tag:     field.Tag,
				Index:   field.Index,
			})
		}
		return reflect.StructOf(list)
	default:
		return t
	}
}

var (
	textMarshalerInt   = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	textUnmarshalerInt = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	yamlMarshalerInt   = reflect.TypeOf((*yaml.Marshaler)(nil)).Elem()
	yamlUnmarshalerInt = reflect.TypeOf((*yaml.Unmarshaler)(nil)).Elem()
)

// translateStructInline out = in (convert to out's type), out MUST be a pointer
// NOTE: the tag here is for future use
func translateStructInline(in, out any, tag reflect.StructTag, toExt bool) {
	rV := reflect.ValueOf(out)
	if rV.Kind() != reflect.Pointer {
		log.Fatalf("%v is not a pointer, but a %v", out, rV.Kind())
	}
	inT := reflect.TypeOf(in)
	inV := reflect.ValueOf(in)
	isNil := false
	if inT.Kind() == reflect.Pointer {
		//a real pointer, not pointer like type like slice

		isNil = inV.IsNil()
		inT = inT.Elem()
		inV = inV.Elem()

	}
	if rV.Elem().Kind() == reflect.Pointer {
		if rV.Elem().IsNil() {
			rV.Elem().Set(reflect.New(rV.Type().Elem().Elem()))
		}
	}

	//setFunc set a=b,  b is type T, a could be either *T or **T,
	setFunc := func(a, b reflect.Value) {
		if !b.IsValid() {
			//if b is not valid, skip
			return
		}

		// fmt.Println("a", a.Interface(), ",b", b.Interface())
		// fmt.Println("a", typeToStr(a.Interface()), ",b", typeToStr(b.Interface()))

		if a.Kind() == reflect.Pointer {
			if a.Elem().Kind() == reflect.Pointer {
				a.Elem().Elem().Set(b)
			} else {
				a.Elem().Set(b)
			}
		} else {
			log.Fatal("a is not a pointer")
			a.Set(b)
		}
	}
	if RegisteredTypes.isSupportedType(inT, toExt) {
		//input is a supported type
		if toExt {
			ext := RegisteredTypes.Get(GetTypeName(inT))
			newV := reflect.New(ext.exType)
			newV.Interface().(extType).init()
			if !isNil {
				newV.Interface().(extType).setVal(inV.Interface())
			}
			setFunc(rV, newV.Elem())
			return
		} else {
			setFunc(rV, reflect.ValueOf(inV.Interface().(toOrigType).toOrig()))
			return
		}
	}
	//check if there is supported marshaling method
	if toExt {
		if inT.Implements(textMarshalerInt) || inT.Implements(yamlMarshalerInt) {
			setFunc(rV, inV)
			return
		}
	} else {
		typeToCheck := rV.Type()
		if rV.Type().Elem().Kind() == reflect.Pointer {
			typeToCheck = rV.Type().Elem()
		}
		if typeToCheck.Implements(textUnmarshalerInt) || typeToCheck.Implements(yamlUnmarshalerInt) {
			setFunc(rV, inV)
			return
		}
	}

	//not a supported type
	if inT.Kind() != reflect.Struct {
		//not a struct
		switch inT.Kind() {
		case reflect.Array, reflect.Slice, reflect.Map:
			switch inT.Kind() {
			case reflect.Array:
				for i := 0; i < inV.Len(); i++ {
					translateStructInline(inV.Index(i).Interface(), rV.Elem().Index(i).Addr().Interface(), tag, toExt)
				}
				return
			case reflect.Slice:
				for i := 0; i < inV.Len(); i++ {
					if i <= rV.Elem().Len()-1 {
						translateStructInline(inV.Index(i).Interface(), rV.Index(i).Addr().Interface(), tag, toExt)
					} else {
						//the current rV len is smaller than input
						newElement := reflect.New(rV.Type().Elem().Elem()).Elem()
						translateStructInline(inV.Index(i).Interface(), newElement.Addr().Interface(), tag, toExt)
						rV.Elem().Set(reflect.Append(rV.Elem(), newElement))
					}

				}
				return
			case reflect.Map:
				//in is map
				iter := inV.MapRange()
				if rV.Elem().IsNil() {
					rV.Elem().Set(reflect.MakeMap(rV.Type().Elem()))
				}
				for iter.Next() {
					newkey := reflect.New(rV.Type().Elem().Key())
					newval := reflect.New(rV.Type().Elem().Elem())
					translateStructInline(iter.Key().Interface(), newkey.Interface(), tag, toExt)
					translateStructInline(iter.Value().Interface(), newval.Interface(), tag, toExt)
					rV.Elem().SetMapIndex(newkey.Elem(), newval.Elem())
				}
				return
			}
		default:
			//in is not supported type and is not a struct, map,slice, array,
			setFunc(rV, inV)
			return
		}
	} else {
		//t is a struct
		if rV.Kind() == reflect.Pointer {
			rV = rV.Elem()
		}
		for i := 0; i < inT.NumField(); i++ {
			if !inT.Field(i).IsExported() {
				continue
			}
			translateStructInline(inV.Field(i).Interface(), rV.Field(i).Addr().Interface(), inT.Field(i).Tag, toExt)
		}
	}
}

// UnmarshalExt unmarshal YAML bytes buf into out, out must be a pointer
func UnmarshalExt(buf []byte, out any) error {
	if reflect.TypeOf(out).Kind() != reflect.Pointer {
		return fmt.Errorf("the object unmarhsal into is not a pointer")
	}
	exType := convertStructType(reflect.TypeOf(out))
	extVal := reflect.New(exType)
	translateStructInline(out, extVal.Interface(), "", true)
	err := yaml.Unmarshal(buf, extVal.Interface())
	if err != nil {
		return err
	}
	translateStructInline(extVal.Interface(), out, "", false)
	return nil
}

// MarshalExt marshal in into YAML bytes
func MarshalExt(in any) ([]byte, error) {
	newType := convertStructType(reflect.TypeOf(in))
	newVal := reflect.New(newType)
	translateStructInline(in, newVal.Interface(), "", true)
	return yaml.Marshal(newVal.Interface())
}
