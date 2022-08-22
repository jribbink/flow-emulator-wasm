//go:build !wasm
// +build !wasm

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
	"github.com/sirupsen/logrus"

	"github.com/onflow/flow-emulator/server/backend"
)

const (
	defaultGRPCPort               = 3569
	defaultRESTPort               = 8888
	defaultAdminPort              = 8080
)

// Native-specific emulator server configuration
type ConfigPlatform struct {
	GRPCPort                  int
	AdminPort                 int
	RESTPort                  int
}

func configureApis(server *EmulatorServer, conf *Config, logger *logrus.Logger, backend *backend.Backend, store *Storage, livenessTicker *LivenessTicker,) (*GRPCServer, *RestServer, *HTTPServer, error) {
	grpcServer := NewGRPCServer(logger, backend, conf.GRPCPort, conf.GRPCDebug)
	restServer, err := NewRestServer(backend, conf.RESTPort, conf.RESTDebug)
	if err != nil {
		logger.WithError(err).Error("❗  Failed to startup REST API")
		return nil, nil, nil, err
	}

	grpc := grpcServer
  rest := restServer
	admin := NewAdminServer(server, backend, store, grpcServer, livenessTicker, conf.AdminPort, conf.HTTPHeaders)

	return grpc, rest, admin, nil
}

func (s *EmulatorServer) startApis() {
	s.logger.
		WithField("port", s.config.GRPCPort).
		Infof("🌱  Starting gRPC server on port %d", s.config.GRPCPort)
	s.group.Add(s.grpc)

	s.logger.
		WithField("port", s.config.RESTPort).
		Infof("🌱  Starting REST API on port %d", s.config.RESTPort)
	s.group.Add(s.rest)

	s.logger.
		WithField("port", s.config.AdminPort).
		Infof("🌱  Starting admin server on port %d", s.config.AdminPort)
	s.group.Add(s.admin)
}

func configureStorage(logger *logrus.Logger, conf *Config) (storage Storage, err error) {
	if conf.Persist {
		return NewBadgerStorage(logger, conf.DBPath, conf.DBGCInterval, conf.DBGCDiscardRatio)
	}

	return NewMemoryStorage(), nil
}

func sanitizeConfigPlatform(conf *Config) *Config {
	if conf.GRPCPort == 0 {
		conf.GRPCPort = defaultGRPCPort
	}

	if conf.RESTPort == 0 {
		conf.RESTPort = defaultRESTPort
	}

	if conf.AdminPort == 0 {
		conf.AdminPort = defaultAdminPort
	}

	return conf
}
