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

package start

import "github.com/onflow/flow-emulator/server"

type ConfigPlatform struct {
	Port			int	`default:"3569" flag:"port,p" info:"port to run RPC server"`
	RestPort	int	`default:"8888" flag:"rest-port" info:"port to run the REST API"`
	AdminPort	int	`default:"8080" flag:"admin-port" info:"port to run the admin API"`
}

func addPlatformConfig(config *server.Config) {
	config.GRPCPort = conf.Port
	config.AdminPort = conf.AdminPort
	config.RESTPort = conf.RestPort 
}