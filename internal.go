// The code in this file is for dev use only
package extyaml

import (
	"fmt"
	"reflect"
	"strings"
)

func strStructTypeTree(inT reflect.Type, prefix string) (rs string) {
	if inT.Kind() == reflect.Pointer {
		//a real pointer, not pointer like type like slice
		inT = inT.Elem()
	}
	if inT.Kind() != reflect.Struct {
		return "not a struct"
	} else {
		for i := 0; i < inT.NumField(); i++ {
			field := inT.Field(i)
			rs += fmt.Sprintf("%s%s --- %s, type %v, anonymouse %v\n", prefix, field.Name, field.Type.String(), field.Type.Name(), field.Anonymous)

			if field.Type.Kind() == reflect.Struct &&
				GetTypeName(field.Type) != "gopkg.in/yaml.v3.Node" &&
				!strings.HasPrefix(field.Type.String(), "extyaml.GeneralExt[") {
				rs += strStructTypeTree(field.Type, prefix+"    ")
			}
		}
	}
	return rs
}

func typeToStr(v any) string {
	t := reflect.TypeOf(v)
	s := ""
	for t.Kind() == reflect.Pointer {
		s += "pointer to "
		t = t.Elem()
	}
	s += t.String()
	return s

}
