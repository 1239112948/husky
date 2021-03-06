package util

import (
	// "fmt"
	"reflect"
)

// 深拷贝
// 结构体、切片之间递归深拷贝
// 整数、浮点数、字符串、布尔类型直接拷贝，其他类型忽略
func DeepCopy(dst, src interface{}) {
	sval := reflect.ValueOf(src)
	dval := reflect.ValueOf(dst)
	doCopy(dval, sval)
}

func doCopy(dval, sval reflect.Value) {
	if !sval.IsValid() {
		return
	}
	if dval.Kind() == reflect.Ptr && dval.IsNil() && dval.CanSet() {
		dval.Set(reflect.New(dval.Type().Elem()))
	}
	sval = reflect.Indirect(sval)
	dval = reflect.Indirect(dval)
	if !dval.CanSet() {
		return
	}
	// fmt.Println(sval.IsValid(), dval.CanSet())
	if testKind(sval.Kind()) != testKind(dval.Kind()) {
		return
	}
	switch testKind(sval.Kind()) {
	case reflect.Int64:
		dval.SetInt(sval.Int())
	case reflect.Uint64:
		dval.SetUint(sval.Uint())
	case reflect.Float64:
		dval.SetFloat(sval.Float())
	case reflect.Bool, reflect.String:
		dval.Set(sval)
	case reflect.Struct:
		for i := 0; i < sval.NumField(); i++ {
			sfield := sval.Field(i)
			sname := sval.Type().Field(i).Name
			dfield := dval.FieldByName(sname)
			// fmt.Println("==", sname, dfield.Kind(), dfield.CanSet())
			// sfield = reflect.Indirect(sfield)
			// dfield = reflect.Indirect(dfield)
			// fmt.Println("====", sname, dfield.Kind())
			doCopy(dfield, sfield)
		}
	case reflect.Slice:
		if size := sval.Len(); size > 0 {
			newval := reflect.MakeSlice(dval.Type(), size, size)
			for i := 0; i < size; i++ {
				v1, v2 := newval.Index(i), sval.Index(i)
				doCopy(v1, v2)
			}
			dval.Set(newval)
		}
	}
}

func testKind(k reflect.Kind) reflect.Kind {
	switch k {
	case reflect.Float32, reflect.Float64:
		return reflect.Float64
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Int:
		return reflect.Int64
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Uint:
		return reflect.Uint64
	}
	return k
}
