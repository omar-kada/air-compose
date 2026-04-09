import "@typespec/http";
import "./shared.tsp";

using Http;
namespace AirComposeService;

@route("/state")
@tag("State")
interface StateAPI {
  @get get(): State | Error;
}

model State {
  nextDeploy: utcDateTime;
  status: DeploymentStatus;
  health: ContainerHealth;
  initialized: boolean;
}
