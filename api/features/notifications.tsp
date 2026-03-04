import "@typespec/http";
import "./shared.tsp";

using Http;
namespace AirComposeService;

@route("/notifications")
@tag("Notifications")
interface NotificationsAPI {
  @get list(@query limit: int32, @query offset?: string): Page<Event> | Error;
}
