package janet

/*
#cgo CFLAGS: -std=c99
#cgo LDFLAGS: -lm -ldl

#include <stdlib.h>
#include <janet.h>
#include <api.h>
*/
import "C"

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

// Wrap a string as a Janet keyword.
func wrapKeyword(word string) C.Janet {
	str := C.CString(word)
	keyword := C.wrap_keyword(str)
	C.free(unsafe.Pointer(str))
	return keyword
}

// Marshal a Go value into a Janet value.
func marshal(item interface{}) (result C.Janet, err error) {
	result = C.janet_wrap_nil()

	type_ := reflect.TypeOf(item)
	value := reflect.ValueOf(item)

	if value.Kind() == reflect.Pointer {
		// TODO(cfoust): 06/23/23 maybe support this someday
		err = fmt.Errorf("cannot marshal pointer to value")
		return
	}

	switch type_.Kind() {
	case reflect.Int:
		result = C.janet_wrap_integer(C.int(value.Int()))
	case reflect.Float64:
		result = C.janet_wrap_number(C.double(value.Float()))
	case reflect.Bool:
		boolean := value.Bool()
		if boolean {
			result = C.janet_wrap_boolean(1)
		} else {
			result = C.janet_wrap_boolean(0)
		}
	case reflect.String:
		strPtr := C.CString(value.String())
		result = C.janet_wrap_string(C.janet_cstring(strPtr))
		defer C.free(unsafe.Pointer(strPtr))
	case reflect.Struct:
		struct_ := C.janet_struct_begin(C.int(type_.NumField()))
		for i := 0; i < type_.NumField(); i++ {
			field := type_.Field(i)
			fieldValue := value.Field(i)

			key_ := wrapKeyword(field.Name)

			value_, fieldErr := marshal(fieldValue.Interface())
			if fieldErr != nil {
				err = fmt.Errorf("could not marshal value '%s': %s", field.Name, fieldErr.Error())
				return
			}

			C.janet_struct_put(struct_, key_, value_)
		}
		result = C.janet_wrap_struct(C.janet_struct_end(struct_))
	case reflect.Array, reflect.Slice:
		numElements := 0
		if type_.Kind() == reflect.Array {
			numElements = type_.Len()
		} else {
			numElements = value.Len()
		}

		array := C.janet_array(C.int(numElements))
		for i := 0; i < numElements; i++ {
			value_, indexErr := marshal(value.Index(i).Interface())
			if indexErr != nil {
				err = indexErr
				return
			}

			C.janet_array_push(array, value_)
		}
		result = C.janet_wrap_array(array)
	default:
		err = fmt.Errorf("unimplemented type: %s", type_.String())
		return
	}

	return
}

var JANET_TYPE_TO_STRING map[C.JanetType]string = map[C.JanetType]string{
	C.JANET_NUMBER:    "number",
	C.JANET_NIL:       "nil",
	C.JANET_BOOLEAN:   "boolean",
	C.JANET_FIBER:     "fiber",
	C.JANET_STRING:    "string",
	C.JANET_SYMBOL:    "symbol",
	C.JANET_KEYWORD:   "keyword",
	C.JANET_ARRAY:     "array",
	C.JANET_TUPLE:     "tuple",
	C.JANET_TABLE:     "table",
	C.JANET_STRUCT:    "struct",
	C.JANET_BUFFER:    "buffer",
	C.JANET_FUNCTION:  "function",
	C.JANET_CFUNCTION: "cfunction",
	C.JANET_ABSTRACT:  "abstract",
	C.JANET_POINTER:   "pointer",
}

func janetTypeString(type_ C.JanetType) string {
	mapping, ok := JANET_TYPE_TO_STRING[type_]
	if !ok {
		return "unknown"
	}

	return mapping
}

func assertType(value C.Janet, expected C.JanetType) (err error) {
	if C.janet_checktype(value, expected) == 1 {
		return
	}

	actual := C.janet_type(value)
	return fmt.Errorf("expected number, got %s", janetTypeString(actual))
}

func prettyPrint(value C.Janet) string {
	ptr := C._pretty_print(value)
	return strings.Clone(C.GoString(ptr))
}

func unmarshal(source C.Janet, dest interface{}) error {
	type_ := reflect.TypeOf(dest)
	value := reflect.ValueOf(dest)

	if value.Kind() != reflect.Pointer {
		return fmt.Errorf("cannot unmarshal into non-pointer value")
	}

	type_ = type_.Elem()
	value = value.Elem()

	switch type_.Kind() {
	case reflect.Int:
		if err := assertType(source, C.JANET_NUMBER); err != nil {
			return err
		}
		unwrapped := C.janet_unwrap_integer(source)
		value.SetInt(int64(unwrapped))
	case reflect.Float64:
		if err := assertType(source, C.JANET_NUMBER); err != nil {
			return err
		}
		unwrapped := C.janet_unwrap_number(source)
		value.SetFloat(float64(unwrapped))
	case reflect.Bool:
		if err := assertType(source, C.JANET_BOOLEAN); err != nil {
			return err
		}
		unwrapped := C.janet_unwrap_boolean(source)
		if unwrapped == 1 {
			value.SetBool(true)
		} else {
			value.SetBool(false)
		}
	case reflect.String:
		if err := assertType(source, C.JANET_STRING); err != nil {
			return err
		}
		strPtr := C.GoString(C.cast_janet_string(C.janet_unwrap_string(source)))
		value.SetString(strings.Clone(strPtr))
	case reflect.Struct:
		if err := assertType(source, C.JANET_STRUCT); err != nil {
			return err
		}

		struct_ := C.janet_unwrap_struct(source)
		for i := 0; i < type_.NumField(); i++ {
			field := type_.Field(i)
			fieldValue := value.Field(i)

			key_ := wrapKeyword(field.Name)
			value_ := C.janet_struct_get(struct_, key_)
			err := unmarshal(value_, fieldValue.Addr().Interface())
			if err != nil {
				return fmt.Errorf("failed to unmarshal struct field %s: %s", field.Name, err.Error())
			}
		}
	case reflect.Array:
		if err := assertType(source, C.JANET_ARRAY); err != nil {
			return err
		}

		wantElements := type_.Len()
		haveElements := int(C.janet_length(source))
		if haveElements != wantElements {
			return fmt.Errorf("janet array had %d elements, wanted %d", haveElements, wantElements)
		}

		for i := 0; i < type_.Len(); i++ {
			value_ := C.janet_get(source, C.janet_wrap_integer(C.int(i)))
			err := unmarshal(value_, value.Index(i).Addr().Interface())
			if err != nil {
				return fmt.Errorf("failed to unmarshal array index %d: %s", i, err.Error())
			}
		}
	case reflect.Slice:
		if err := assertType(source, C.JANET_ARRAY); err != nil {
			return err
		}

		haveElements := int(C.janet_length(source))

		element := type_.Elem()
		slice := reflect.MakeSlice(type_, 0, 0)

		for i := 0; i < int(haveElements); i++ {
			value_ := C.janet_get(source, C.janet_wrap_integer(C.int(i)))
			entry := reflect.New(element)
			err := unmarshal(value_, entry.Interface())
			if err != nil {
				return fmt.Errorf("failed to unmarshal slice index %d: %s", i, err.Error())
			}
			slice = reflect.Append(slice, entry.Elem())
		}

		value.Set(slice)
	default:
		return fmt.Errorf("unimplemented type: %s", type_.String())
	}

	return nil
}
