import "@typespec/http";
import "@typespec/versioning";
import "./features/auth.tsp";
import "./features/config.tsp";
import "./features/deployment.tsp";
import "./features/features.tsp";
import "./features/notifications.tsp";
import "./features/settings.tsp";
import "./features/stats.tsp";
import "./features/status.tsp";
import "./features/user.tsp";

using Http;
using Versioning;

@service(#{ title: "AirCompose API" })
@versioned(Versions)
@route("/api")
@useAuth(BearerAuth /*| OpenIdConnectAuth<"">*/)
namespace AirComposeService;

enum Versions {
  v1: "1.0",
}
