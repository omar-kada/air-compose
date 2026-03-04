import "@typespec/http";
import "./shared.tsp";

using Http;
namespace AirComposeService;

@route("/status")
@tag("Status")
interface StatusAPI {
  @get get(): StackStatus[] | Error;
}

model StackStatus {
  stackId: string;
  name: string;
  services: Array<ContainerStatus>;
}

model ContainerStatus {
  containerId: string;
  state:
    | "created"
    | "running"
    | "paused"
    | "restarting"
    | "removing"
    | "exited"
    | "dead";
  name: string;
  health: ContainerHealth;
  startedAt: utcDateTime;
}
