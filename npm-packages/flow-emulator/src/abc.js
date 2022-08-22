import { Emulator } from "./emulator";
import * as nodeFetch from "node-fetch";
import * as fcl from "@onflow/fcl";

const emulator = new Emulator();

const httpRequest = async ({ hostname, path, body, headers = {}, method }) => {
  const bodyJSON = body ? JSON.stringify(body) : "";
  const res = await emulator.restHandler(path, {
    body: bodyJSON,
    headers,
    method,
  });
  if (res.ok) return res.json();
  throw new Error("request failed");
};

(async () => {
  await emulator.start();
  await fcl.config().put("accessNode.api", "http://foo.com");
  console.log(await fcl.block({}, { httpRequest }));
  setTimeout(async () => {
    await emulator.stop();
    await emulator.start();
    console.log("HELLO");
  }, 1000);
})();
