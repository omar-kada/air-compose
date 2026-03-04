import "@typespec/http";
import "./shared.tsp";

using Http;
namespace AirComposeService;

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

model Credentials {
  username: string;
  @secret password: string;
}
