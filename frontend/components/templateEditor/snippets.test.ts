import { describe, expect, it } from "vitest";
import {
  getArgForParam,
  getFunctionInsertSnippet,
  getKnownFunctionSnippet,
  normalizeVariableName,
} from "./snippets";

describe("template editor snippets", () => {
  it("normalizes variable names to root context", () => {
    expect(normalizeVariableName(".FeatureGroups")).toBe("$.FeatureGroups");
    expect(normalizeVariableName("$.Rows")).toBe("$.Rows");
    expect(normalizeVariableName("Rows")).toBe("Rows");
  });

  it("returns known function snippets", () => {
    expect(getKnownFunctionSnippet("formatSteps")).toBe(
      "{{formatSteps (index $.Rows 0).Instructions}}"
    );
    expect(getKnownFunctionSnippet("displayTitle")).toBe(
      "{{displayTitle (index $.Rows 0).Feature (index $.Rows 0).Scenario}}"
    );
    expect(getKnownFunctionSnippet("unknown")).toBeUndefined();
  });

  it("builds function snippet for unknown functions", () => {
    expect(
      getFunctionInsertSnippet({
        name: "custom",
        signature: "custom() string",
        description: "",
      })
    ).toBe("{{custom}}");

    expect(
      getFunctionInsertSnippet({
        name: "custom",
        signature: "custom(s string) string",
        description: "",
      })
    ).toBe("{{custom \"\"}}");

    expect(
      getFunctionInsertSnippet({
        name: "custom",
        signature: "custom(feature, scenario string) string",
        description: "",
      })
    ).toBe("{{custom \"\" \"\"}}");
  });

  it("infers arguments by name and type", () => {
    expect(getArgForParam("row SpecRow")).toBe(".");
    expect(getArgForParam("rows []SpecRow")).toBe(".Rows");
    expect(getArgForParam("text string")).toBe('""');
    expect(getArgForParam("count int")).toBe("0");
    expect(getArgForParam("enabled bool")).toBe("false");
    expect(getArgForParam("value interface{}")).toBe("nil");
  });
});
