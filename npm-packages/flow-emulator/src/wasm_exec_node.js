// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

"use strict";

import nodeFetch from "node-fetch";
import wasmExec from "./wasm_exec";
import os from "os";
import crypto from "crypto";
import fs from "fs";
import { TextEncoder, TextDecoder } from "util";

export let i = 0;

export default async (...argv) => {
  globalThis.require = require;
  globalThis.fs = fs;
  globalThis.TextEncoder = TextEncoder;

  globalThis.TextDecoder = TextDecoder;
  globalThis.Uint8Array = Uint8Array;
  globalThis.Response = nodeFetch.Response;
  globalThis.Object = Object;

  globalThis.performance = {
    now() {
      const [sec, nsec] = process.hrtime();
      return sec * 1000 + nsec / 1000000;
    },
  };

  globalThis.crypto = {
    getRandomValues(b) {
      crypto.randomFillSync(b);
    },
  };

  await wasmExec();

  const go = new Go();
  go.argv = argv;
  go.env = Object.assign({ TMPDIR: os.tmpdir() }, process.env);

  WebAssembly.instantiate(fs.readFileSync(argv[0]), go.importObject)
    .then((result) => {
      process.on("exit", (code) => {
        // Node.js exits if no event handler is pending
        if (code === 0 && !go.exited) {
          // deadlock, make Go print error and stack traces
          go._pendingEvent = { id: 0 };
          go._resume();
        }
      });
      return go.run(result.instance);
    })
    .catch((err) => {
      console.error(err);
      process.exit(1);
    });
};
