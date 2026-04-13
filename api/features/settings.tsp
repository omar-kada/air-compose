import "@typespec/http";
import "./shared.tsp";

using Http;
namespace AirComposeService;

@route("/settings")
@tag("Settings")
interface SettingsAPI {
  @get get(): Settings | Error;
  @post set(@body settings: Settings): Settings | Error;
  @post @route("/test-git") testGitConnection(
    @body gitCredentials: GitCredentails,
  ): BooleanResponse;
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

model GitCredentails {
  repo: string;
  branch?: string;
  username?: string;
  token?: string;
}
