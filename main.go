package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	log "github.com/cihub/seelog"
	httpr "github.com/julienschmidt/httprouter"
	"github.com/satori/go.uuid"
)

type Person struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
	Size int    `json:"size"`
}

type HelloWorldPayload struct {
	S string
	I int
}

type payload struct{}

var Payload payload

func HelloWorld(ctx context.Context, r *http.Request, p httpr.Params) (int, interface{}) {
	payload := ctx.Value(Payload).(HelloWorldPayload)
	return http.StatusOK,
		Person{Id: uuid.NewV4().String(), Name: payload.S, Age: payload.I, Size: payload.I}
}

func CheckPerson(ctx context.Context, r *http.Request, p httpr.Params) (int, interface{}) {
	payload := ctx.Value(Payload).(struct{ Person })

	if payload.Age > 18 {
		return http.StatusOK,
			struct{ Data string }{Data: "you can acces this place ..."}
	}

	return http.StatusBadRequest,
		struct{ Error string }{Error: "wowowow, you need to be at least 18 to access"}
}

func Hello(r *http.Request, p httpr.Params) (int, interface{}) {
	return http.StatusNoContent, nil
}

func WriteJsonResponse(w http.ResponseWriter, status int, body interface{}) {
	rawBody := []byte{}
	if body != nil {
		rawBody, _ = json.Marshal(body)
	}
	WriteResponse(w, status, string(rawBody))
}

func WriteResponse(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Length", fmt.Sprintf("%v", len(body)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `%v`, body)
}

func Json(
	f func(context.Context, *http.Request, httpr.Params) (int, interface{}),
	i interface{},
) func(http.ResponseWriter, *http.Request, httpr.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httpr.Params) {
		if body, err := ioutil.ReadAll(r.Body); err != nil {
			WriteResponse(w, http.StatusBadRequest, "wrapper: enable to read payload")
			return
		} else {
			v := reflect.New(reflect.TypeOf(i))
			iNew := v.Interface()
			if err := json.Unmarshal(body, iNew); err != nil {
				// json error
				log.Errorf(
					"[Wrapper] invalid request, %s, input was %s",
					err.Error(), string(body))
				WriteResponse(w, http.StatusBadRequest, "wrapper: invalid json request")
				return
			}
			ctx := context.WithValue(context.Background(), Payload, reflect.Indirect(v).Interface())
			// lets process the handler
			status, res := f(ctx, r, p)
			// status, res := f(reflect.Indirect(v).Interface(), r, p)
			WriteJsonResponse(w, status, res)
		}

	}
}

func EmptyJson(
	f func(*http.Request, httpr.Params) (int, interface{}),
) func(http.ResponseWriter, *http.Request, httpr.Params) {
	return func(w http.ResponseWriter, r *http.Request, p httpr.Params) {
		status, res := f(r, p)
		WriteJsonResponse(w, status, res)
	}
}

func main() {
	r := httpr.New()

	r.POST("/hello/world", Json(HelloWorld, HelloWorldPayload{}))
	r.POST("/check/person", Json(CheckPerson, struct{ Person }{}))
	r.GET("/hello", EmptyJson(Hello))

	log.Info("Starting http server")
	log.Critical(http.ListenAndServe(fmt.Sprintf(":%v", 1492), r))
}
