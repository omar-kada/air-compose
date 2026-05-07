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
  registered(): RegistrationInfo | Error;

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

@route("/oidc")
@tag("OIDC")
interface OIDCAPI {
  @useAuth(NoAuth)
  @route("login")
  @get
  oidcLogin(): void | Error;

  @useAuth(NoAuth)
  @route("callback")
  @get
  oidcCallback(@query code: string, @query state: string): void | Error;
}
model Credentials {
  username: string;
  @secret password: string;
}

model RegistrationInfo {
  registered: boolean;
  oidc: boolean;
}
