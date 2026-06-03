package struct_tools

import (
	"errors"
	"fmt"
	"reflect"
)

func setField(v interface{}, name string) (error, reflect.Value) {
	// v must be a pointer to a struct
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return errors.New("v must be pointer to struct"), reflect.Value{}
	}

	// Dereference pointer
	rv = rv.Elem()

	// Lookup field by name
	fv := rv.FieldByName(name)

	if !fv.IsValid() {
		return fmt.Errorf("not a field name: %s", name), fv
	}

	// Field must be exported
	if !fv.CanSet() {
		return fmt.Errorf("cannot set field %s", name), fv
	}

	return nil, fv

}

func SetStringField(v interface{}, name string, value string) error {

	err, fv := setField(v, name)
	if err != nil {
		return err
	}
	// We expect a string field
	if fv.Kind() != reflect.String {
		return fmt.Errorf("%s is not a string field", name)
	}

	// Set the value
	fv.SetString(value)

	return nil
}

func SetIntField(v interface{}, name string, value int64) error {
	err, fv := setField(v, name)
	if err != nil {
		return err
	}

	// We expect a string field
	if fv.Kind() != reflect.Int &&
		fv.Kind() != reflect.Int32 &&
		fv.Kind() != reflect.Int64 {
		return fmt.Errorf("%s is not a string field", name)
	}

	// Set the value
	fv.SetInt(value)
	return nil
}

func SetUintField(v interface{}, name string, value uint64) error {
	err, fv := setField(v, name)
	if err != nil {
		return err
	}

	// We expect a string field
	if fv.Kind() != reflect.Uint &&
		fv.Kind() != reflect.Uint32 &&
		fv.Kind() != reflect.Uint64 {
		return fmt.Errorf("%s is not a string field", name)
	}

	// Set the value
	fv.SetUint(value)
	return nil
}

func SetFloatField(v interface{}, name string, value float64) error {
	err, fv := setField(v, name)
	if err != nil {
		return err
	}

	// We expect a string field
	if fv.Kind() != reflect.Float64 {
		return fmt.Errorf("%s is not a string field", name)
	}

	// Set the value
	fv.SetFloat(value)
	return nil
}

func SetBoolField(v interface{}, name string, value bool) error {
	err, fv := setField(v, name)
	if err != nil {
		return err
	}

	// We expect a string field
	if fv.Kind() != reflect.Bool {
		return fmt.Errorf("%s is not a string field", name)
	}

	// Set the value
	fv.SetBool(value)
	return nil
}
