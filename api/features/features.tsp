import "@typespec/http";
import "./shared.tsp";

using Http;
namespace AirComposeService;

@route("/features")
@tag("Features")
interface FeaturesAPI {
  @get get(): Features | Error;
}

model Features {
  displayConfig: boolean;
  editConfig: boolean;
  displayLogs: boolean;
  editSettings: boolean;
}
