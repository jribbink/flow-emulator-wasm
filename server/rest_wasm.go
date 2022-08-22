//go:build wasm
// +build wasm

/*
 * Flow Emulator
 *
 * Copyright 2019-2020 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"syscall/js"

	"github.com/onflow/flow-emulator/server/backend"
	"github.com/onflow/flow-go/engine/access/rest"
	"github.com/onflow/flow-go/model/flow"
	"github.com/rs/zerolog"
)

type RestServer struct {
	server   		*http.Server
	jsInstance 	js.Value
}

func (r *RestServer) Start() error {
	Serve(r.server.Handler, "restHandler", r.jsInstance)
	return nil
}

func (r *RestServer) Stop() {
	_ = r.server.Shutdown(context.Background())
}

func NewRestServer(be *backend.Backend, jsInstanceName string, debug bool) (*RestServer, error) {
	logger := zerolog.Logger{}

	srv, err := rest.NewServer(backend.NewAdapter(be), "127.0.0.1:3333", logger, flow.Emulator.Chain())
	if err != nil {
		return nil, err
	}

	jsInstance := js.Global().Get(jsInstanceName)

	return &RestServer{
		server:   srv,
		jsInstance: jsInstance,
	}, nil
}

func Serve(handler http.Handler, jsHandlerName string, jsHandlerParent js.Value) func() {
	jsHandler := js.FuncOf(func(_ js.Value, args []js.Value) any {
		var resolve func(js.Value)
		var reject func(js.Value)

		promise := js.Global().Get("Promise").New(js.FuncOf(func(_ js.Value, args []js.Value) any {
			resolve = func(value js.Value) {
				args[0].Invoke(value)
			}
	
			reject = func(value js.Value) {
				args[1].Invoke(value)
			}
	
			return js.Undefined()
		}))

		go func() {
			defer func() {
				if r := recover(); r != nil {
					if err, ok := r.(error); ok {
						reject(js.ValueOf(fmt.Sprintf("panic: %+v\n", err)))
					} else {
						reject(js.ValueOf(fmt.Sprintf("panic: %v\n", r)))
					}
				}
			}()

			res := NewResponseRecorder()

			handler.ServeHTTP(res, Request(args[0], args[1]))

			resolve(res.JSResponse())
		}()

		return promise
	})

	jsHandlerParent.Set(jsHandlerName, jsHandler)

	return jsHandler.Release
}

func Request(url js.Value, r js.Value) *http.Request {
	body := r.Get("body").String()

	req := httptest.NewRequest(
		r.Get("method").String(),
		url.String(),
		bytes.NewBuffer([]byte(body)),
	)

	headers := r.Get("headers")
	if(!headers.IsNull() && !headers.IsUndefined()) {
		headersKeys := js.Global().Get("Object").Call("keys", headers)
		for i := 0; i < headersKeys.Length(); i++ {
			key := headersKeys.Get(fmt.Sprint(i)).String()
			req.Header.Set(key, headers.Get(key).String())
		}
	}

	return req
}


// ResponseRecorder uses httptest.ResponseRecorder to build a JS Response
type ResponseRecorder struct {
	*httptest.ResponseRecorder
}

// NewResponseRecorder returns a new ResponseRecorder
func NewResponseRecorder() ResponseRecorder {
	return ResponseRecorder{httptest.NewRecorder()}
}

// JSResponse builds and returns the equivalent JS Response
func (rr ResponseRecorder) JSResponse() js.Value {
	var res = rr.Result()

	var body js.Value = js.Undefined()
	if res.ContentLength != 0 {
		var b, err = ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		body = js.Global().Get("Uint8Array").New(len(b))
		js.CopyBytesToJS(body, b)
	}

	var init = make(map[string]interface{}, 2)

	if res.StatusCode != 0 {
		init["status"] = res.StatusCode
	}

	if len(res.Header) != 0 {
		var headers = make(map[string]interface{}, len(res.Header))
		for k := range res.Header {
			headers[k] = res.Header.Get(k)
		}
		init["headers"] = headers
	}

	return js.Global().Get("Response").New(body, init)
}