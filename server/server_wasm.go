//go:build wasm
// +build wasm

/*
 * Flow Emulator
 *
 * Copyright 2019-2022 Dapper Labs, Inc.
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
	"syscall"
	"syscall/js"

	"github.com/onflow/flow-emulator/server/backend"
	"github.com/sirupsen/logrus"
)

// WASM-specific emualtor server config
type ConfigPlatform struct {
	JsInstanceName string
}

func configureApis(server *EmulatorServer, conf *Config, logger *logrus.Logger, backend *backend.Backend, store *Storage, livenessTicker *LivenessTicker,) (*GRPCServer, *RestServer, *HTTPServer, error) {
	grpcServer := NewGRPCServer(logger, backend, conf.JsInstanceName, conf.GRPCDebug)
	restServer, err := NewRestServer(backend, conf.JsInstanceName, conf.RESTDebug)
	if err != nil {
		logger.WithError(err).Error("❗  Failed to startup REST API")
		return nil, nil, nil, err
	}

	grpc := grpcServer
  rest := restServer
	admin := NewAdminServer(server, backend, store, grpcServer, livenessTicker, conf.JsInstanceName, conf.HTTPHeaders)

	js.Global().Get(conf.JsInstanceName).Set("stopHandler", js.FuncOf(func(this js.Value, args []js.Value) any {
		server.group.Stop()
		return js.Undefined()
	}))

	return grpc, rest, admin, nil
}

func (s *EmulatorServer) startApis() {
	s.logger.
		WithField("instance", s.config.JsInstanceName).
		Infof("🌱  Starting gRPC server on JS instance '%s'", s.config.JsInstanceName)
	s.group.Add(s.grpc)

	s.logger.
		WithField("instance", s.config.JsInstanceName).
		Infof("🌱  Starting REST API on JS instance '%s'", s.config.JsInstanceName)
	s.group.Add(s.rest)

	s.logger.
		WithField("instance", s.config.JsInstanceName).
		Infof("🌱  Starting admin server on JS instance '%s'", s.config.JsInstanceName)
	s.group.Add(s.admin)
}

func configureStorage(logger *logrus.Logger, conf *Config) (storage Storage, err error) {
	if conf.Persist {
		return nil, syscall.ENOSYS
	}

	return NewMemoryStorage(), nil
}

func sanitizeConfigPlatform(conf *Config) *Config {
	return conf
}
