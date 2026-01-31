import type { TemplateFunction } from "@/lib/mdflowApi";

export function normalizeVariableName(name: string) {
  if (name.startsWith("$.")) return name;
  if (name.startsWith(".")) return `$.${name.slice(1)}`;
  return name;
}

export function getFunctionInsertSnippet(func: TemplateFunction) {
  const mapped = getKnownFunctionSnippet(func.name);
  if (mapped) {
    return mapped;
  }

  const signature = func.signature || "";
  const openParen = signature.indexOf("(");
  const closeParen = signature.indexOf(")");
  if (openParen === -1 || closeParen === -1 || closeParen <= openParen + 1) {
    return `{{${func.name}}}`;
  }

  const params = expandParamList(signature.slice(openParen + 1, closeParen));

  if (params.length === 0) {
    return `{{${func.name}}}`;
  }

  const args = params.map((param) => getArgForParam(param)).join(" ");
  return `{{${func.name} ${args}}}`;
}

export function getKnownFunctionSnippet(name: string) {
  const rowRef = "(index $.Rows 0)";
  const rowField = (field: string) => `${rowRef}.${field}`;
  const snippets: Record<string, string> = {
    formatSteps: `{{formatSteps ${rowField("Instructions")}}}`,
    formatBullets: `{{formatBullets ${rowField("Expected")}}}`,
    notEmpty: `{{notEmpty ${rowField("Instructions")}}}`,
    displayTitle: `{{displayTitle ${rowField("Feature")} ${rowField("Scenario")}}}`,
    escapeYAML: `{{escapeYAML ${rowField("Scenario")}}}`,
    cellValue: `{{cellValue ${rowRef} "Header"}}`,
    headerCell: `{{headerCell "Header"}}`,
    metadataPairs: `{{metadataPairs ${rowRef}}}`,
    trimPrefix: `{{trimPrefix ${rowField("Scenario")} "prefix"}}`,
    lower: `{{lower ${rowField("Scenario")}}}`,
    upper: `{{upper ${rowField("Scenario")}}}`,
    replace: `{{replace ${rowField("Scenario")} "old" "new"}}`,
  };

  return snippets[name];
}

export function getArgForParam(param: string) {
  const parts = param.split(/\s+/).filter(Boolean);
  const nameToken = parts[0] || "";
  const typeToken = parts.length > 1 ? parts[parts.length - 1] : "";
  const nameLower = nameToken.toLowerCase();
  const normalized = typeToken.replace(/\s+/g, "").replace(/^\*+/, "");
  const normalizedLower = normalized.toLowerCase();

  if (nameLower.includes("rows")) {
    return ".Rows";
  }

  if (normalizedLower.includes("[]")) {
    if (normalizedLower.includes("specrow") || normalizedLower.includes("row")) {
      return ".Rows";
    }
    if (normalizedLower.includes("string")) {
      return "(list)";
    }
    return "nil";
  }

  if (nameLower.includes("row")) {
    return ".";
  }

  if (normalizedLower.includes("specrow") || normalizedLower.includes("row")) {
    return ".";
  }
  if (normalizedLower.includes("string")) {
    return "\"\"";
  }
  if (normalizedLower === "interface{}" || normalizedLower === "any") {
    return "nil";
  }
  if (/^(u?int(8|16|32|64)?|float(32|64)?)$/.test(normalizedLower)) {
    return "0";
  }
  if (normalizedLower.includes("bool")) {
    return "false";
  }

  return "nil";
}

function expandParamList(text: string) {
  const items = text
    .split(",")
    .map((param) => param.trim())
    .filter(Boolean);

  const result: string[] = [];
  let currentType = "";

  for (let i = items.length - 1; i >= 0; i -= 1) {
    const item = items[i];
    const spaceIndex = item.lastIndexOf(" ");
    if (spaceIndex > 0) {
      const name = item.slice(0, spaceIndex).trim();
      const type = item.slice(spaceIndex + 1).trim();
      if (type) {
        currentType = type;
      }
      result.unshift(name ? `${name} ${type}` : item);
      continue;
    }

    if (currentType) {
      result.unshift(`${item} ${currentType}`);
      continue;
    }

    result.unshift(item);
  }

  return result;
}
