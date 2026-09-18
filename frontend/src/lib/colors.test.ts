import { EventType } from "@/api/api";
import type { ContainerHealth } from "@/api/api";
import {
  colorForStatus,
  borderForStatus,
  textColorForStatus,
  logColor,
} from "./colors";

describe("colorForStatus", () => {
  it("maps healthy/success to green", () => {
    expect(colorForStatus("healthy")).toBe("bg-green-400");
    expect(colorForStatus("success")).toBe("bg-green-400");
  });

  it("maps unhealthy/error to red", () => {
    expect(colorForStatus("unhealthy")).toBe("bg-red-400");
    expect(colorForStatus("error")).toBe("bg-red-400");
  });

  it("maps starting/planned to slate", () => {
    expect(colorForStatus("starting")).toBe("bg-slate-400");
    expect(colorForStatus("planned")).toBe("bg-slate-400");
  });

  it("maps running to blue", () => {
    expect(colorForStatus("running")).toBe("bg-blue-400");
  });

  it("returns empty string for unknown statuses", () => {
    expect(colorForStatus("unknown" as ContainerHealth)).toBe("");
  });
});

describe("borderForStatus", () => {
  it("maps running/healthy to green", () => {
    expect(borderForStatus("running")).toBe("border-green-400");
    expect(borderForStatus("healthy")).toBe("border-green-400");
  });

  it("maps dead/removing/unhealthy to red", () => {
    expect(borderForStatus("dead")).toBe("border-red-400");
    expect(borderForStatus("removing")).toBe("border-red-400");
    expect(borderForStatus("unhealthy")).toBe("border-red-400");
  });

  it("maps exited/paused/none to slate", () => {
    expect(borderForStatus("exited")).toBe("border-slate-400");
    expect(borderForStatus("paused")).toBe("border-slate-400");
    expect(borderForStatus("none")).toBe("border-slate-400");
  });

  it("maps created/restarting/starting to blue", () => {
    expect(borderForStatus("created")).toBe("border-blue-400");
    expect(borderForStatus("restarting")).toBe("border-blue-400");
    expect(borderForStatus("starting")).toBe("border-blue-400");
  });

  it("returns empty string for undefined", () => {
    expect(borderForStatus(undefined)).toBe("");
  });
});

describe("textColorForStatus", () => {
  it("maps running/healthy to green", () => {
    expect(textColorForStatus("running")).toBe("text-green-400");
    expect(textColorForStatus("healthy")).toBe("text-green-400");
  });

  it("maps dead/removing/unhealthy to red", () => {
    expect(textColorForStatus("dead")).toBe("text-red-400");
    expect(textColorForStatus("removing")).toBe("text-red-400");
    expect(textColorForStatus("unhealthy")).toBe("text-red-400");
  });

  it("maps exited/paused/none to slate", () => {
    expect(textColorForStatus("exited")).toBe("text-slate-400");
    expect(textColorForStatus("paused")).toBe("text-slate-400");
    expect(textColorForStatus("none")).toBe("text-slate-400");
  });

  it("maps created/restarting/starting to blue", () => {
    expect(textColorForStatus("created")).toBe("text-blue-400");
    expect(textColorForStatus("restarting")).toBe("text-blue-400");
    expect(textColorForStatus("starting")).toBe("text-blue-400");
  });

  it("returns empty string for undefined", () => {
    expect(textColorForStatus(undefined)).toBe("");
  });
});

describe("logColor", () => {
  it("maps ERROR and DEPLOYMENT_ERROR to red", () => {
    expect(logColor(EventType.ERROR)).toBe("text-red-700 dark:text-red-300 ");
    expect(logColor(EventType.DEPLOYMENT_ERROR)).toBe(
      "text-red-700 dark:text-red-300 ",
    );
  });

  it("maps MISC to gray", () => {
    expect(logColor(EventType.MISC)).toBe("text-gray-700 dark:text-gray-300");
  });

  it("returns empty string for unhandled event types", () => {
    expect(logColor(EventType.DEPLOYMENT_SUCCESS)).toBe("");
    expect(logColor(EventType.DEPLOYMENT_STARTED)).toBe("");
    expect(logColor(EventType.HEALTH_CHANGE)).toBe("");
    expect(logColor(EventType.CONFIGURATION_UPDATED)).toBe("");
    expect(logColor(EventType.PASSWORD_UPDATED)).toBe("");
    expect(logColor(EventType.SESSION_REUSED)).toBe("");
  });
});
