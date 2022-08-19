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
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/onflow/flow-go/access"
	legacyaccess "github.com/onflow/flow-go/access/legacy"
	"github.com/onflow/flow-go/model/flow"
	accessproto "github.com/onflow/flow/protobuf/go/flow/access"
	legacyaccessproto "github.com/onflow/flow/protobuf/go/flow/legacy/access"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/onflow/flow-emulator/server/backend"
)

type GRPCServer struct {
	logger     *logrus.Logger
	grpcServer *grpc.Server
}

func NewGRPCServer(logger *logrus.Logger, b *backend.Backend) *GRPCServer {
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpcprometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpcprometheus.UnaryServerInterceptor),
	)

	chain := flow.Emulator.Chain()
	adaptedBackend := backend.NewAdapter(b)

	legacyaccessproto.RegisterAccessAPIServer(grpcServer, legacyaccess.NewHandler(adaptedBackend, chain))
	accessproto.RegisterAccessAPIServer(grpcServer, access.NewHandler(adaptedBackend, chain))

	grpcprometheus.Register(grpcServer)

	return &GRPCServer{
		logger:     logger,
		grpcServer: grpcServer,
	}
}

func (g *GRPCServer) Server() *grpc.Server {
	return g.grpcServer
}