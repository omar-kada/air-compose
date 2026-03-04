model BooleanResponse {
  success: boolean;
}

@error
model Error {
  code: ErrorCode;
  message: string;
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
  _unknown: "unknown",
}

model Page<T> {
  @pageItems items: T[];
  pageInfo: PageInfo;
}

model PageInfo {
  hasNextPage: boolean;
  @continuationToken endCursor: string;
}
