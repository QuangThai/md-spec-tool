import { useCallback } from "react";

interface ExportFunctionalityProps {
  mdflowOutput: string;
}

/**
 * Custom hook for export functionality
 * Handles file downloads in multiple formats (MD, JSON, YAML)
 */
export function useExportFunctionality({ mdflowOutput }: ExportFunctionalityProps) {
  const downloadAs = useCallback(
    (format: "md" | "json" | "yaml") => {
      let content = mdflowOutput;
      let filename = "spec";
      let mimeType = "text/plain";

      if (format === "json") {
        // Extract YAML frontmatter and convert to JSON
        const yamlMatch = mdflowOutput.match(/^---\n([\s\S]*?)\n---/);
        const frontmatter = yamlMatch ? yamlMatch[1] : "";
        const body = mdflowOutput.replace(/^---\n[\s\S]*?\n---\n?/, "");

        content = JSON.stringify(
          {
            frontmatter: frontmatter.split("\n").reduce((acc, line) => {
              const [key, ...vals] = line.split(":");
              if (key && vals.length) acc[key.trim()] = vals.join(":").trim();
              return acc;
            }, {} as Record<string, string>),
            content: body,
          },
          null,
          2
        );
        filename = "spec.json";
        mimeType = "application/json";
      } else if (format === "yaml") {
        filename = "spec.yaml";
        mimeType = "text/yaml";
      } else {
        filename = "spec.mdflow.md";
        mimeType = "text/markdown";
      }

      const blob = new Blob([content], { type: mimeType });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    },
    [mdflowOutput]
  );

  return { downloadAs };
}
