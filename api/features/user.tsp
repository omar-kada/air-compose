import "@typespec/http";
import "./shared.tsp";

using Http;
namespace AirComposeService;

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

model User {
  username: string;
}
