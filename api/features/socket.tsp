import "@typespec/events";
import "@typespec/http";
import "@typespec/streams";
using TypeSpec.Http;

namespace AirComposeService;

@route("/ws")
interface WebSocket {
    @get
    connect():
        | {
              @statusCode _: 101;
          }
        | {
              @statusCode _: 401;
          };
}

// ── Client → Server ──────────────────────────────────────

model StartLogsMessage {
    previousLines: int32;
}

model EndLogsMessage {}

@discriminated
union ClientMessage {
    startLogs: StartLogsMessage,
    endLogs: EndLogsMessage,
}

// ── Server → Client ──────────────────────────────────────

model StateMessage {
    nextDeploy: utcDateTime;
    status: DeploymentStatus;
    health: ContainerHealth;
    deployment: Deployment;
}

model LogLine {
    msg: string;
    time: utcDateTime;
    level: string;
}

model LogMessages is Array<LogLine>;

@discriminated
union ServerMessage {
    state: StateMessage,
    log: LogLine,
    previousLogs: LogMessages,
    newDeployment: Deployment,
}
