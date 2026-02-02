"use client";

import { formatShortcut, SHORTCUTS } from "@/lib/useKeyboardShortcuts";
import { useBodyScrollLock } from "@/lib/useBodyScrollLock";
import { Command } from "cmdk";
import { AnimatePresence, motion } from "framer-motion";
import {
  Copy,
  Download,
  Eye,
  FileCode,
  FileText,
  History,
  ShieldCheck,
  Zap,
} from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";

const RECENT_COMMANDS_KEY = "mdflow-recent-commands";
const MAX_RECENT = 5;

type CommandId =
  | "convert"
  | "copy"
  | "export"
  | "toggle-preview"
  | "history"
  | "template-editor"
  | "validation"
  | `template-${string}`;

interface CommandItem {
  id: CommandId;
  label: string;
  icon: React.ReactNode;
  shortcut?: string;
  group: "recent" | "actions" | "templates" | "tools";
  onSelect: () => void;
  disabled?: boolean;
}

export interface CommandPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConvert: () => void;
  onCopy: () => void;
  onExport: () => void;
  onTogglePreview: () => void;
  onShowHistory: () => void;
  onOpenTemplateEditor: () => void;
  onOpenValidation: () => void;
  templates: string[];
  currentTemplate: string;
  onSelectTemplate: (template: string) => void;
  hasOutput: boolean;
}

export function CommandPalette({
  open,
  onOpenChange,
  onConvert,
  onCopy,
  onExport,
  onTogglePreview,
  onShowHistory,
  onOpenTemplateEditor,
  onOpenValidation,
  templates,
  currentTemplate,
  onSelectTemplate,
  hasOutput,
}: CommandPaletteProps) {
  useBodyScrollLock(open);
  const [recentCommandIds, setRecentCommandIds] = useState<CommandId[]>([]);

  useEffect(() => {
    try {
      const stored = localStorage.getItem(RECENT_COMMANDS_KEY);
      if (stored) {
        setRecentCommandIds(JSON.parse(stored));
      }
    } catch {}
  }, []);

  const addToRecent = useCallback((id: CommandId) => {
    setRecentCommandIds((prev) => {
      const filtered = prev.filter((c) => c !== id);
      const updated = [id, ...filtered].slice(0, MAX_RECENT);
      try {
        localStorage.setItem(RECENT_COMMANDS_KEY, JSON.stringify(updated));
      } catch {}
      return updated;
    });
  }, []);

  const handleSelect = useCallback(
    (id: CommandId, action: () => void) => {
      addToRecent(id);
      action();
      onOpenChange(false);
    },
    [addToRecent, onOpenChange]
  );

  const allCommands: CommandItem[] = useMemo(() => {
    const actions: CommandItem[] = [
      {
        id: "convert",
        label: "Convert",
        icon: <Zap className="h-4 w-4" />,
        shortcut: formatShortcut(SHORTCUTS.CONVERT),
        group: "actions",
        onSelect: () => handleSelect("convert", onConvert),
      },
      {
        id: "copy",
        label: "Copy Output",
        icon: <Copy className="h-4 w-4" />,
        shortcut: formatShortcut(SHORTCUTS.COPY),
        group: "actions",
        onSelect: () => handleSelect("copy", onCopy),
        disabled: !hasOutput,
      },
      {
        id: "export",
        label: "Export",
        icon: <Download className="h-4 w-4" />,
        shortcut: formatShortcut(SHORTCUTS.EXPORT),
        group: "actions",
        onSelect: () => handleSelect("export", onExport),
        disabled: !hasOutput,
      },
      {
        id: "toggle-preview",
        label: "Toggle Preview",
        icon: <Eye className="h-4 w-4" />,
        shortcut: formatShortcut(SHORTCUTS.TOGGLE_PREVIEW),
        group: "actions",
        onSelect: () => handleSelect("toggle-preview", onTogglePreview),
      },
    ];

    const templateItems: CommandItem[] = templates.map((t) => ({
      id: `template-${t}` as CommandId,
      label: t === currentTemplate ? `${t} (current)` : t,
      icon: <FileText className="h-4 w-4" />,
      group: "templates",
      onSelect: () =>
        handleSelect(`template-${t}` as CommandId, () => onSelectTemplate(t)),
    }));

    const tools: CommandItem[] = [
      {
        id: "template-editor",
        label: "Template Editor",
        icon: <FileCode className="h-4 w-4" />,
        group: "tools",
        onSelect: () => handleSelect("template-editor", onOpenTemplateEditor),
      },
      {
        id: "validation",
        label: "Validation Rules",
        icon: <ShieldCheck className="h-4 w-4" />,
        group: "tools",
        onSelect: () => handleSelect("validation", onOpenValidation),
      },
      {
        id: "history",
        label: "History",
        icon: <History className="h-4 w-4" />,
        group: "tools",
        onSelect: () => handleSelect("history", onShowHistory),
      },
    ];

    return [...actions, ...templateItems, ...tools];
  }, [
    templates,
    currentTemplate,
    hasOutput,
    handleSelect,
    onConvert,
    onCopy,
    onExport,
    onTogglePreview,
    onShowHistory,
    onOpenTemplateEditor,
    onOpenValidation,
    onSelectTemplate,
  ]);

  const recentCommands = useMemo(() => {
    return recentCommandIds
      .map((id) => allCommands.find((c) => c.id === id))
      .filter((c): c is CommandItem => c !== undefined);
  }, [recentCommandIds, allCommands]);

  const actionCommands = allCommands.filter((c) => c.group === "actions");
  const templateCommands = allCommands.filter((c) => c.group === "templates");
  const toolCommands = allCommands.filter((c) => c.group === "tools");

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.15 }}
          className="fixed inset-0 z-50 flex items-start justify-center pt-[20vh] bg-black/60 backdrop-blur-sm"
          onClick={() => onOpenChange(false)}
        >
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            transition={{ duration: 0.15, ease: [0.16, 1, 0.3, 1] }}
            onClick={(e) => e.stopPropagation()}
          >
            <Command
              className="w-[560px] max-h-[400px] overflow-hidden rounded-xl border border-white/20 bg-black/95 shadow-2xl"
              onKeyDown={(e) => {
                if (e.key === "Escape") {
                  onOpenChange(false);
                }
              }}
            >
              <Command.Input
                placeholder="Type a command or search..."
                className="w-full border-b border-white/10 bg-transparent px-4 py-3 text-sm text-white placeholder:text-white/40 outline-none"
                autoFocus
              />
              <Command.List className="max-h-[320px] overflow-y-auto p-2">
                <Command.Empty className="py-6 text-center text-sm text-white/50">
                  No results found
                </Command.Empty>

                {recentCommands.length > 0 && (
                  <Command.Group
                    heading="Recent"
                    className="**:[[cmdk-group-heading]]:px-2 **:[[cmdk-group-heading]]:py-1.5 **:[[cmdk-group-heading]]:text-xs **:[[cmdk-group-heading]]:font-medium **:[[cmdk-group-heading]]:text-white/50"
                  >
                    {recentCommands.map((cmd) => (
                      <CommandItem key={`recent-${cmd.id}`} command={cmd} />
                    ))}
                  </Command.Group>
                )}

                <Command.Group
                  heading="Actions"
                  className="**:[[cmdk-group-heading]]:px-2 **:[[cmdk-group-heading]]:py-1.5 **:[[cmdk-group-heading]]:text-xs **:[[cmdk-group-heading]]:font-medium **:[[cmdk-group-heading]]:text-white/50"
                >
                  {actionCommands.map((cmd) => (
                    <CommandItem key={cmd.id} command={cmd} />
                  ))}
                </Command.Group>

                <Command.Group
                  heading="Templates"
                  className="**:[[cmdk-group-heading]]:px-2 **:[[cmdk-group-heading]]:py-1.5 **:[[cmdk-group-heading]]:text-xs **:[[cmdk-group-heading]]:font-medium **:[[cmdk-group-heading]]:text-white/50"
                >
                  {templateCommands.map((cmd) => (
                    <CommandItem key={cmd.id} command={cmd} />
                  ))}
                </Command.Group>

                <Command.Group
                  heading="Tools"
                  className="**:[[cmdk-group-heading]]:px-2 **:[[cmdk-group-heading]]:py-1.5 **:[[cmdk-group-heading]]:text-xs **:[[cmdk-group-heading]]:font-medium **:[[cmdk-group-heading]]:text-white/50"
                >
                  {toolCommands.map((cmd) => (
                    <CommandItem key={cmd.id} command={cmd} />
                  ))}
                </Command.Group>
              </Command.List>
            </Command>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}

function CommandItem({ command }: { command: CommandItem }) {
  return (
    <Command.Item
      value={command.label}
      onSelect={command.onSelect}
      disabled={command.disabled}
      className="flex items-center gap-3 px-2 py-2 text-sm text-white/80 rounded-lg cursor-pointer data-[selected=true]:bg-accent-orange/20 hover:bg-white/10 data-[disabled=true]:opacity-40 data-[disabled=true]:pointer-events-none"
    >
      <span className="text-white/60">{command.icon}</span>
      <span className="flex-1">{command.label}</span>
      {command.shortcut && (
        <kbd className="px-1.5 py-0.5 text-[10px] font-medium text-white/40 bg-white/10 rounded">
          {command.shortcut}
        </kbd>
      )}
    </Command.Item>
  );
}
