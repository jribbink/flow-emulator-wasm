import exec from "./wasm_exec_node";
import path from "path";

export class Emulator {
  constructor() {
    this.httpHandler = null;
    this.globalKey = "flow-emulator";
  }

  async start() {
    globalThis[this.globalKey] = this;

    return new Promise((resolve) => {
      var args = [`--globalKey=${this.globalKey}`];
      const wasm = path.resolve(__dirname, "../bin/emulator.wasm");
      exec(wasm, ...args);

      const originalSetHttpHandler = this.setHttpHandler;
      this.setHttpHandler = (handler) => {
        if (handler) resolve();
        originalSetHttpHandler(handler);
      };
    });
  }

  setHttpHandler(handler) {
    this.httpHandler = handler;
  }
}
