import "@typespec/http";
import "./shared.tsp";

using Http;
namespace AirComposeService;

@route("/config")
@tag("Config")
interface ConfigAPI {
  @get get(): Config | Error;
  @post set(@body config: Config): Config | Error;
}

model Config {
  globalVariables: Record<string>;
  services: Record<Record<string>>;
}
