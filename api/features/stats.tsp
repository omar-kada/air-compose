import "@typespec/http";
import "./shared.tsp";

using Http;
namespace AirComposeService;

@route("/stats")
@tag("Stats")
interface StatsAPI {
  @get get(@path days: int32): Stats | Error;
}

model Stats {
  error: int32;
  success: int32;
  nextDeploy: utcDateTime;
  lastDeploy: utcDateTime;
  author: string;
  status: DeploymentStatus;
  health: ContainerHealth;
}
