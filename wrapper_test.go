package main

import (
	"encoding/json"
	"reflect"
	"testing"

	httpr "github.com/julienschmidt/httprouter"
)

var body = "{\"S\": \"bob\", \"I\": 42}"

func BenchmarkNoReflect(b *testing.B) {
	s := HelloWorldPayload{}
	json.Unmarshal([]byte(body), &s)
	HelloWorld(s, httpr.Params{})
}

func BenchmarkReflect(b *testing.B) {
	var i interface{} = HelloWorldPayload{}
	v := reflect.New(reflect.TypeOf(i))
	iNew := v.Interface()
	json.Unmarshal([]byte(body), iNew)
	HelloWorld(reflect.Indirect(v).Interface(), httpr.Params{})
}
