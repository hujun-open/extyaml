package extyaml

import (
	"fmt"
	"reflect"
)

type registeredType struct {
	origType, exType reflect.Type
	fromStr          FromStr
	toStr            ToStr
}

// GetTypeName returns a name string for the type t
func GetTypeName(t reflect.Type) string {
	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}

// Registry maintains all registed types
type Registry struct {
	origToExtTypeList map[string]*registeredType
}

// RegisteredTypes is the global Registry
var RegisteredTypes = Registry{origToExtTypeList: make(map[string]*registeredType)}

func (reg *Registry) isSupportedType(t reflect.Type, isCheckingOrig bool) (ok bool) {
	if isCheckingOrig {
		_, ok = reg.origToExtTypeList[GetTypeName(t)]
		return ok
	} else {

		for _, regt := range reg.origToExtTypeList {
			if GetTypeName(regt.exType) == GetTypeName(t) {
				return true
			}
		}
		return false
	}
}

// Get returns a registered type, nil if no such type;
// typename is the string returned by GetTypeName()
func (reg *Registry) Get(typename string) *registeredType {
	return reg.origToExtTypeList[typename]
}

// RegisterExt register a new type, with supplied to,from function,
// should be called in init()
func RegisterExt[T any](to ToStr, from FromStr) {
	ext := newGeneralExt[T]()
	origType := reflect.TypeOf(*new(T))
	exType := reflect.TypeOf(ext)
	RegisteredTypes.origToExtTypeList[GetTypeName(origType)] = &registeredType{
		origType: origType,
		exType:   exType,
		toStr:    to,
		fromStr:  from,
	}
}
