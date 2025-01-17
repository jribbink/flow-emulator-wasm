/*
 * Flow Emulator
 *
 * Copyright 2019 Dapper Labs, Inc.
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
	"encoding/json"
	"net/http"

	fvmerrors "github.com/onflow/flow-go/fvm/errors"

	flowsdk "github.com/onflow/flow-go-sdk"

	"github.com/gorilla/mux"
	"github.com/onflow/flow-emulator/server/backend"
	"github.com/onflow/flow-emulator/storage/badger"
)

type BlockResponse struct {
	Height  int    `json:"height"`
	BlockId string `json:"blockId"`
	Context string `json:"context,omitempty"`
}

type EmulatorAPIServer struct {
	router  *mux.Router
	server  *EmulatorServer
	backend *backend.Backend
	storage *Storage
}

func NewEmulatorAPIServer(server *EmulatorServer, backend *backend.Backend, storage *Storage) *EmulatorAPIServer {
	router := mux.NewRouter().StrictSlash(true)
	r := &EmulatorAPIServer{router: router,
		server:  server,
		backend: backend,
		storage: storage,
	}

	router.HandleFunc("/emulator/newBlock", r.CommitBlock)
	router.HandleFunc("/emulator/snapshot/{name}", r.Snapshot)
	router.HandleFunc("/emulator/storages/{address}", r.Storage)

	return r
}

func (m EmulatorAPIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.router.ServeHTTP(w, r)
}

func (m EmulatorAPIServer) CommitBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	m.backend.CommitBlock()

	header, err := m.backend.GetLatestBlockHeader(r.Context(), true)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	blockResponse := &BlockResponse{
		Height:  int(header.Height),
		BlockId: header.ID().String(),
	}

	err = json.NewEncoder(w).Encode(blockResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (m EmulatorAPIServer) Snapshot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	name := vars["name"]

	switch (*m.storage).Store().(type) {
	case *badger.Store:
		badgerStore := (*m.storage).Store().(*badger.Store)
		err := badgerStore.JumpToContext(name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		blockchain, err := configureBlockchain(m.server.config, badgerStore)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		m.backend.SetEmulator(blockchain)
		block, err := blockchain.GetLatestBlock()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		blockResponse := &BlockResponse{
			Height:  int(block.Header.Height),
			BlockId: block.Header.ID().String(),
			Context: name,
		}

		err = json.NewEncoder(w).Encode(blockResponse)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	default:
		m.server.logger.Error("State management only available with badger storage")
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func (m EmulatorAPIServer) Storage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	address := vars["address"]

	addr := flowsdk.HexToAddress(address)

	storage, err := m.backend.GetAccountStorage(addr)
	if err != nil {
		if fvmerrors.IsAccountNotFoundError(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(storage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
