import "@typespec/http";
import "@typespec/versioning";

using Http;
using Versioning;

@service(#{ title: "AirCompose API" })
@versioned(Versions)
@route("/api")
@useAuth(BearerAuth /*| OpenIdConnectAuth<"">*/)
namespace DeploymentService;
enum Versions {
  v1: "1.0",
}

model Event {
  ID: uint64;
  time: utcDateTime;
  msg: string;
  type: EventType;
  objectId: uint64;
  objectName: string;
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

model FileDiff {
  oldFile: string;
  newFile: string;
  diff: string;
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

model Stats {
  error: int32;
  success: int32;
  nextDeploy: utcDateTime;
  lastDeploy: utcDateTime;
  author: string;
  status: DeploymentStatus;
  health: ContainerHealth;
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

enum ErrorCode {
  InvalidToken: "INVALID_TOKEN",
  InvalidCredentials: "INVALID_CREDENTIALS",
  InvalidRequest: "INVALID_REQUEST",
  NotAllowed: "NOT_ALLOWED",
  Disabled: "DISABLED",
  NotFound: "NOT_FOUND",
  ServerError: "SERVER_ERROR",
}

model Settings {
  repo: string;
  branch?: string;
  cron?: string;
  username?: string;
  token?: string;
  notificationURL?: string;
  notificationTypes: Array<EventType>;
}

model Config {
  globalVariables: Record<string>;
  services: Record<Record<string>>;
}

model Features {
  displayConfig: boolean;
  editConfig: boolean;
  displayLogs: boolean;
  editSettings: boolean;
}

model Page<T> {
  items: T[];
  pageInfo: PageInfo;
}

model PageInfo {
  hasNextPage: boolean;
  endCursor: string;
}

@error
model Error {
  code: ErrorCode;
  message: string;
}

model User {
  username: string;
}

model Credentials {
  username: string;
  password: string;
}

model BooleanResponse {
  success: boolean;
}

@route("/auth")
@tag("Auth")
interface AuthAPI {
  @useAuth(NoAuth)
  @route("register")
  @get
  registered(): {
    registered: boolean;
  } | Error;

  @useAuth(NoAuth)
  @route("register")
  @post
  register(@body user: Credentials): BooleanResponse | Error;

  @useAuth(NoAuth)
  @route("login")
  @post
  login(@body credentials: Credentials): BooleanResponse | Error;

  @route("refresh")
  @post
  refresh(): BooleanResponse | Error;

  @route("logout")
  @post
  logout(): BooleanResponse | Error;
}

@route("/user")
@tag("user")
interface UserAPI {
  @get get(): User | Error;
  @post @route("change-password") changePassword(
    oldPass: string,
    newPass: string,
  ): BooleanResponse | Error;
  @delete delete(): BooleanResponse | Error;
}

@route("/deployment")
@tag("Deployment")
interface DeployementAPI {
  /** List deployments */
  @get list(
    @query limit: int32,
    @query offset?: string,
  ): Page<Deployment> | Error;
  /** Read deployment */
  @get read(@path id: string): DeploymentWithDetails | Error;
  /** Sync and deploy if changes */
  @post sync(): DeploymentWithDetails | void | Error;
}

@route("/settings")
@tag("Settings")
interface SettingsAPI {
  @get get(): Settings | Error;
  @post set(@body settings: Settings): Settings | Error;
}

@route("/config")
@tag("Config")
interface ConfigAPI {
  @get get(): Config | Error;
  @post set(@body config: Config): Config | Error;
}

@route("/features")
@tag("Features")
interface FeaturesAPI {
  @get get(): Features | Error;
}

@route("/notifications")
@tag("Notifications")
interface NotificationsAPI {
  @get list(@query limit: int32, @query offset?: string): Page<Event> | Error;
}

@route("/diff")
@tag("Diff")
interface DiffAPI {
  @get get(): FileDiff[] | Error;
}

@route("/stats")
@tag("Stats")
interface StatsAPI {
  @get get(@path days: int32): Stats | Error;
}

@route("/status")
@tag("Status")
interface StatusAPI {
  @get get(): StackStatus[] | Error;
}
