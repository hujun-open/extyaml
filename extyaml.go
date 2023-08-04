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

const (
	//the first letter of the prefix must be lower case to make it non-export
	SkipNamingPrefix = "skippedExtYAMLField"
	SkipTag          = "skipyamlmarshal"
)

func convertStructType(t reflect.Type) reflect.Type {
	isPtr := false
	if t.Kind() == reflect.Pointer {
		//a real pointer, not pointer like type like slice
		t = t.Elem()
		//if it is pointer, then also return a pointer type
		isPtr = true
	}
	if RegisteredTypes.isSupportedType(t, true) {
		//input is a supported type
		if isPtr {
			return reflect.PointerTo(RegisteredTypes.Get(GetTypeName(t)).exType)
		}
		return RegisteredTypes.Get(GetTypeName(t)).exType
	}
	//check if the input has supported marshaling method
	if t.Implements(textMarshalerInt) || t.Implements(yamlMarshalerInt) {
		if isPtr {
			return reflect.PointerTo(t)
		}
		return t
	}

	switch t.Kind() {
	case reflect.Array:
		if isPtr {
			return reflect.PointerTo(reflect.ArrayOf(t.Len(), convertStructType(t.Elem())))
		}
		return reflect.ArrayOf(t.Len(), convertStructType(t.Elem()))
	case reflect.Slice:
		if isPtr {
			return reflect.PointerTo(reflect.SliceOf(convertStructType(t.Elem())))
		}
		return reflect.SliceOf(convertStructType(t.Elem()))
	case reflect.Map:
		if isPtr {
			return reflect.PointerTo(reflect.MapOf(convertStructType(t.Key()), convertStructType(t.Elem())))
		}
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

			if _, exists := field.Tag.Lookup(SkipTag); exists {
				//if field has skip tag, then prepending a naming prefix that make it non-export
				list = append(list, reflect.StructField{
					Name:    SkipNamingPrefix + field.Name,
					Type:    field.Type,
					PkgPath: t.PkgPath(),
					Tag:     field.Tag,
					Index:   field.Index,
				})
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
		if isPtr {
			return reflect.PointerTo(reflect.StructOf(list))
		}
		return reflect.StructOf(list)
	default:
		if isPtr {
			return reflect.PointerTo(t)
		}
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
	wipeFunc := func(r interface{}) {
		v := reflect.ValueOf(r)
		p := v.Elem()
		p.Set(reflect.Zero(p.Type()))
	}
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
		if isNil {
			wipeFunc(rV.Interface())
			return
		}

	}
	if rV.Elem().Kind() == reflect.Pointer {
		if rV.Elem().IsNil() {
			rV.Elem().Set(reflect.New(rV.Type().Elem().Elem()))
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
			if rV.Elem().Kind() == reflect.Pointer {
				//this is for pointer to arrary/slice/map
				rV = rV.Elem()
			}
			switch inT.Kind() {
			case reflect.Array:
				if inV.IsValid() {
					for i := 0; i < inV.Len(); i++ {
						translateStructInline(inV.Index(i).Interface(), rV.Elem().Index(i).Addr().Interface(), tag, toExt)
					}
				}
				return
			case reflect.Slice:
				for i := 0; i < inV.Len(); i++ {
					if i <= rV.Elem().Len()-1 {
						translateStructInline(inV.Index(i).Interface(), rV.Elem().Index(i).Addr().Interface(), tag, toExt)
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
		for rV.Kind() == reflect.Pointer {
			rV = rV.Elem()
		}

		for i := 0; i < inT.NumField(); i++ {
			// fmt.Println("waling field", inT.Field(i).Name)
			if !inT.Field(i).IsExported() {
				continue
			}
			toChkField := inT.Field(i)
			if toExt {
				toChkField = rV.Type().Field(i)
			}
			if _, exists := toChkField.Tag.Lookup(SkipTag); exists {
				continue
			}
			fieldRint := rV.Field(i).Addr().Interface()
			// if rV.Field(i).Kind() == reflect.Ptr {
			// 	fieldRint = rV.Field(i).Interface()
			// }
			translateStructInline(inV.Field(i).Interface(), fieldRint, inT.Field(i).Tag, toExt)
		}
	}
}

// UnmarshalExt unmarshal YAML bytes buf into out, out must be a pointer
func UnmarshalExt(buf []byte, out any) error {
	if reflect.TypeOf(out).Kind() != reflect.Pointer {
		return fmt.Errorf("the object unmarhsal into is not a pointer")
	}
	exType := convertStructType(reflect.TypeOf(out))
	exType = exType.Elem()
	extVal := reflect.New(exType) //this is needed to avoid pointer to pointer
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
	inV := reflect.ValueOf(in)
	if inV.Kind() == reflect.Pointer {
		inV = inV.Elem()
	}
	newType := convertStructType(reflect.TypeOf(inV.Interface()))
	newVal := reflect.New(newType)
	translateStructInline(inV.Interface(), newVal.Interface(), "", true)
	return yaml.Marshal(newVal.Interface())
}
