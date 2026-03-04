import "@typespec/http";
import "./shared.tsp";

using Http;
namespace AirComposeService;

@route("/settings")
@tag("Settings")
interface SettingsAPI {
  @get get(): Settings | Error;
  @post set(@body settings: Settings): Settings | Error;
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
