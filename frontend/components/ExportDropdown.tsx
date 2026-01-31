import { useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { Download, FileJson, FileText } from "lucide-react";

interface ExportDropdownProps {
  mdflowOutput: string;
}

type ExportFormat = "md" | "json" | "yaml";

/**
 * ExportDropdown - Exports content in multiple formats
 * Supports MD, JSON (with parsed frontmatter), and YAML
 */
export function ExportDropdown({ mdflowOutput }: ExportDropdownProps) {
  const [open, setOpen] = useState(false);

  const downloadAs = (format: ExportFormat) => {
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
    setOpen(false);
  };

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="flex items-center justify-center h-8 px-3 rounded-lg bg-accent-orange text-[9px] font-bold uppercase tracking-wider text-white shadow-md shadow-accent-orange/25 hover:bg-accent-orange/90 active:scale-95 transition-all cursor-pointer shrink-0"
      >
        <span className="inline-flex items-center gap-1.5 leading-none">
          <Download className="block w-3 h-3 shrink-0" />
          <span className="leading-none">Export</span>
        </span>
      </button>

      <AnimatePresence>
        {open && (
          <>
            <div
              className="fixed inset-0 z-40"
              onClick={() => setOpen(false)}
            />
            <motion.div
              initial={{ opacity: 0, y: -8, scale: 0.95 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, y: -8, scale: 0.95 }}
              className="absolute right-0 top-full mt-2 z-50 w-48 rounded-xl bg-black/90 backdrop-blur-xl border border-white/20 shadow-2xl overflow-hidden"
            >
              <div className="p-1">
                <button
                  onClick={() => downloadAs("md")}
                  className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-white/10 transition-colors cursor-pointer"
                >
                  <FileText className="w-4 h-4 text-accent-orange" />
                  <div className="text-left">
                    <p className="text-[11px] font-bold text-white">Markdown</p>
                    <p className="text-[9px] text-white/50">.mdflow.md</p>
                  </div>
                </button>
                <button
                  onClick={() => downloadAs("json")}
                  className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-white/10 transition-colors cursor-pointer"
                >
                  <FileJson className="w-4 h-4 text-blue-400" />
                  <div className="text-left">
                    <p className="text-[11px] font-bold text-white">JSON</p>
                    <p className="text-[9px] text-white/50">.json</p>
                  </div>
                </button>
                <button
                  onClick={() => downloadAs("yaml")}
                  className="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg hover:bg-white/10 transition-colors cursor-pointer"
                >
                  <FileText className="w-4 h-4 text-green-400" />
                  <div className="text-left">
                    <p className="text-[11px] font-bold text-white">YAML</p>
                    <p className="text-[9px] text-white/50">.yaml</p>
                  </div>
                </button>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div>
  );
}
