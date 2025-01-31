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

package badger

import (
	"github.com/dgraph-io/badger/v2"
)

// Config defines the configurable parameters of the Badger storage implementation.
type Config struct {
	// Logger is where Badger system logs will go.
	Logger badger.Logger
	// DBPath the path to the database directory.
	DBPath string
	// Truncate whether to truncate the write log to remove corrupt data.
	Truncate bool
	// Allow git snapshots.
	Snapshot bool
	// Badger options to open DB
	BadgerOptions badger.Options
}

// getBadgerconfig returns configuration object with  options and
// a Badger Options object defining the current  configuration.
// It starts with the defaultConfig, applies any options
// to it, then merges with the Badger default options.

func getBadgerConfig(opts ...Opt) Config {
	conf := defaultConfig
	for _, applyOption := range opts {
		applyOption(&conf)
	}

	conf.BadgerOptions = badger.DefaultOptions(conf.DBPath)
	conf.BadgerOptions.Logger = conf.Logger
	conf.BadgerOptions.Truncate = conf.Truncate

	return conf
}

// noopLogger implements the badger.Logger interface and discards all logs.
type noopLogger struct{}

func (noopLogger) Errorf(string, ...interface{})   {}
func (noopLogger) Warningf(string, ...interface{}) {}
func (noopLogger) Infof(string, ...interface{})    {}
func (noopLogger) Debugf(string, ...interface{})   {}

// The default config to use when instantiating a Badger store.
var defaultConfig = Config{
	Logger:   noopLogger{},
	DBPath:   "./flowdb",
	Snapshot: false,
}

type Opt func(*Config)

func WithPath(path string) Opt {
	return func(c *Config) {
		c.DBPath = path
	}
}
func WithSnapshot(enabled bool) Opt {
	return func(c *Config) {
		c.Snapshot = enabled
	}
}
func WithLogger(logger badger.Logger) Opt {
	return func(c *Config) {
		c.Logger = logger
	}
}

func WithTruncate(trunc bool) Opt {
	return func(c *Config) {
		c.Truncate = trunc
	}
}
