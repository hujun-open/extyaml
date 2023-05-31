package extyaml

import (
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"
)

// FromStr is the function convert a string into a instance of to-be-supported-type
type FromStr func(s string) (any, error)

// ToStr is the function convert a instance of to-be-supported-type into string
type ToStr func(in any) (string, error)

type generalExt[T any] struct {
	//all fields must be comparable, otherwise this won't work as map key
	origV *T
}

func newGeneralExt[T any]() generalExt[T] {
	return generalExt[T]{origV: new(T)}
}

func (ext *generalExt[T]) setVal(in any) {
	*ext.origV = in.(T)
}

func (ext generalExt[T]) toOrig() any {
	return *ext.origV
}

func (ext *generalExt[T]) init() {
	ext.origV = new(T)
}

func (ext generalExt[T]) MarshalYAML() (interface{}, error) {
	pkgName := GetTypeName(reflect.TypeOf(new(T)).Elem())
	toFunc := RegisteredTypes.Get(pkgName).toStr
	if toFunc == nil {
		return nil, fmt.Errorf("can't find %v toStr Func, it is not registed?", pkgName)
	}
	return toFunc(*ext.origV)
}

func (ext *generalExt[T]) UnmarshalYAML(value *yaml.Node) error {
	pkgName := GetTypeName(reflect.TypeOf(new(T)).Elem())
	fromFunc := RegisteredTypes.Get(pkgName).fromStr
	if fromFunc == nil {
		return fmt.Errorf("can't find %v fromFunc, it is not registed?", pkgName)
	}
	val, err := fromFunc(value.Value)
	if err != nil {
		return err
	}
	ext.origV = new(T)
	*ext.origV = val.(T)
	return nil
}
