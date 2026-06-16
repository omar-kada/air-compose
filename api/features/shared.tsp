import "@typespec/http";
using TypeSpec.Http;

model BooleanResponse {
  success: boolean;
}

@error
model Error {
  code: ErrorCode;
  message: string;
}

@error
model NotFoundError is Error {
  @statusCode statusCode: 404;
}

enum ErrorCode {
  InvalidToken: "INVALID_TOKEN",
  InvalidCredentials: "INVALID_CREDENTIALS",
  InvalidRequest: "INVALID_REQUEST",
  NotAllowed: "NOT_ALLOWED",
  Disabled: "DISABLED",
  NotFound: "NOT_FOUND",
  ServerError: "SERVER_ERROR",
}

model Event {
  ID: uint64;
  time: utcDateTime;
  msg: string;
  type: EventType;
  objectId: uint64;
  objectName: string;
}

enum EventType {
  Misc: "MISC",
  Error: "ERROR",
  DeploymentStarted: "DEPLOYMENT_STARTED",
  DeploymentSuccess: "DEPLOYMENT_SUCCESS",
  DeploymentError: "DEPLOYMENT_ERROR",
  StacksUnhealthy: "STACKS_UNHEALTHY",
  StacksHealthy: "STACKS_HEALTHY",
  ConfigurationUpdated: "CONFIGURATION_UPDATED",
  PasswordUpdated: "PASSWORD_UPDATED",
  SessionReused: "SESSION_REUSED",
}

enum DeploymentStatus {
  planned: "planned",
  running: "running",
  success: "success",
  error: "error",
}

enum ContainerHealth {
  healthy: "healthy",
  unhealthy: "unhealthy",
  starting: "starting",
  none: "none",
}

enum ContainerState {
  created: "created",
  running: "running",
  paused: "paused",
  restarting: "restarting",
  removing: "removing",
  exited: "exited",
  dead: "dead",
}

model Page<T> {
  @pageItems items: T[];
  pageInfo: PageInfo;
}

model PageInfo {
  hasNextPage: boolean;
  @continuationToken endCursor: string;
}
