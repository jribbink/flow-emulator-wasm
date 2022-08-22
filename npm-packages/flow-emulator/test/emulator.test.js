import { Emulator } from "../src/emulator";
import * as nodeFetch from "node-fetch";
import * as fcl from "@onflow/fcl";
import { i } from "../src/wasm_exec_node";

it("works", async () => {
  const emulator = new Emulator();

  const httpRequest = async ({
    hostname,
    path,
    body,
    headers = {},
    method,
  }) => {
    const bodyJSON = body ? JSON.stringify(body) : "";
    const res = await emulator.restHandler(path, {
      body: bodyJSON,
      headers,
      method,
    });
    if (res.ok) return res.json();
    throw new Error("request failed");
  };

  console.time("t");
  await emulator.start();
  console.log(i, "i");
  console.timeEnd("t");
  await fcl.config().put("accessNode.api", "http://foo.com");
  await emulator.stop();
  console.log("HELLO");
});
