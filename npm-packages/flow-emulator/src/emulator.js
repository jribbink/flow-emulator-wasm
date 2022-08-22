import exec from "./wasm_exec_node";
import path from "path";

export class Emulator {
  constructor() {
    this.globalKey = "__FLOW_EMULATOR__";
    this.restHandler = null;
    this.httpHandler = null;
    this.grpcHandler = null;
    this.stopHandler = null;
  }

  async start() {
    globalThis[this.globalKey] = this;

    var args = [
      `--js-instance-name=${this.globalKey}`,
      `--service-priv-key=68ee617d9bf67a4677af80aaca5a090fcda80ff2f4dbc340e0e36201fa1f1d8c`,
    ];
    const wasm = path.resolve(__dirname, "../bin/emulator.wasm");
    exec(wasm, ...args);

    await this.waitForHandlers();
  }

  async stop() {
    this.stopHandler();
  }

  waitForHandlers() {
    const handlers = ["restHandler"];
    return Promise.all(
      handlers.map(
        (handler) =>
          new Promise((resolve) =>
            Object.defineProperty(this, handler, {
              set(value) {
                if (value) {
                  Object.defineProperty(this, handler, {
                    value,
                  });
                  resolve();
                }
              },
            })
          )
      )
    );
  }
}
