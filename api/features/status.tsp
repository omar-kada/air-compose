import "@typespec/http";
import "./shared.tsp";

using Http;
namespace AirComposeService;

@route("/status")
@tag("Status")
interface StatusAPI {
  @get get(): StackStatus | Error;
}

model StackStatus is Record<Record<ContainerStatus>>;

model ContainerStatus {
  containerId: string;
  state: ContainerState;
  name: string;
  health: ContainerHealth;
  startedAt: utcDateTime;
}
