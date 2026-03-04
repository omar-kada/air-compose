import "@typespec/http";
import "./shared.tsp";

using Http;
namespace AirComposeService;

@route("/deployment")
@tag("Deployment")
interface DeployementAPI {
  /** List deployments */
  @list list(
    @query @pageSize limit: int32,
    @query @continuationToken offset?: string,
  ): Page<Deployment> | Error;
  /** Read deployment */
  @get read(@path id: string): DeploymentWithDetails | Error;
  /** Sync and deploy if changes */
  @post sync(): DeploymentWithDetails | void | Error;
}

@route("/diff")
@tag("Diff")
interface DiffAPI {
  @get get(): FileDiff[] | Error;
}

model Deployment {
  id: string;
  title: string;
  author: string;
  time: utcDateTime;
  endTime: utcDateTime;
  diff: string;
  status: DeploymentStatus;
}

model DeploymentWithDetails is Deployment {
  files: FileDiff[];
  events: Event[];
}

model FileDiff {
  oldFile: string;
  newFile: string;
  diff: string;
}
