package utils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"reflect"
	"testing"
)

type myInt int

var (
	myIntType = reflect.TypeOf(myInt(1))
)

type Test struct {
	J int
	S string
}

type Src struct {
	//I []*Test
	M map[string]int
}

type Dst struct {
	//I []Test
	M map[string]int
}

type Person struct {
	Name string
	Age  int
}

func TestConvertError_Error(T *testing.T) {
	///*	ss := "ssss"
	//	src := Src{
	//		I: &Test{
	//			J:1,
	//			S:ss,
	//		},
	//	}
	//	var dst Dst
	//	_ = Convert_test(src, dst)
	//	c, _ := json.Marshal(src)
	//	fmt.Println(c)*/
	//
	//	val := Src{
	//		I: &Test{
	//			J:1,
	//			S:"222",
	//		},
	//	}
	//	var vv interface{}
	//	vv = val
	//	v := reflect.ValueOf(&vv).Elem()
	///*	t := v.Type()
	//	obj := reflect.Zero(t).Interface()
	//	o := obj
	//	fmt.Println(o)
	//	ov := reflect.ValueOf(&obj)
	//	ov = ov.Elem()
	//	fmt.Println()
	//	cv := reflect.ValueOf(&obj)
	//	cv = cv.Elem()
	//	cv.Set(reflect.ValueOf(val))
	//	oo := obj.(Src)
	//	fmt.Println(cv, oo)
	//	f := v.FieldByName("I")
	//	fmt.Println(f.IsValid())*/
	//

	//field := v.Elem().Field(0)
	//ft := v.Type().Elem().Field(0).Type
	//fmt.Println(ft)
	//newF := reflect.Zero(ft)
	//field.Set(newF)
	//fmt.Println(val, val.I)

	//person := Person{}
	//fmt.Println(person) // 修改前 { 0}
	//pp := reflect.ValueOf(&person).Elem() // 取得struct变量的指针
	//fmt.Println("num", pp.NumField())
	//for i := 0; i < pp.NumField(); i++ {
	//	fmt.Println(pp.Field(i))
	//}
	//field := pp.FieldByName("Name") //获取指定Field
	//field.SetString("gerrylon") // 设置值
	//
	//field = pp.FieldByName("Age")
	//field.SetInt(26)
	//
	//fmt.Println(person) // 修改后 {gerrylon 26}
	//i := 1
	//src := Src{
	//	//I:[]*Test{
	//	//	{
	//	//		J:1,
	//	//		S:"ss",
	//	//	},
	//	//},
	//	M: map[string]int{
	//		"aa": i,
	//	},
	//}
	//dst, err := convert(src, reflect.TypeOf(Dst{}))
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
	//d := reflect.ValueOf(dst).Elem().Interface()
	//fmt.Println(d.(Dst))
}

type T struct {
	Val *int
	A   *A
	M   map[string]bool
}

type A struct {
	Str string
}

func TestGob(t *testing.T) {
	val := 1
	tt := T{
		Val: &val,
		A: &A{
			Str: "hello",
		},
		M: make(map[string]bool),
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(tt)
	if err != nil {
		log.Fatal(err)
	}
	var ttt T
	dnc := gob.NewDecoder(&buf)
	err = dnc.Decode(&ttt)
	fmt.Println(*ttt.Val, ttt.A.Str)
}
