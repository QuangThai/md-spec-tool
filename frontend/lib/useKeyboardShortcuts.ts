import { useEffect, useCallback } from 'react';

export interface ShortcutDefinition {
  key: string;
  mod?: boolean;
  shift?: boolean;
  label: string;
}

export const SHORTCUTS = {
  COMMAND_PALETTE: { key: 'k', mod: true, label: 'Command Palette' },
  CONVERT: { key: 'Enter', mod: true, label: 'Convert' },
  COPY: { key: 'c', mod: true, shift: true, label: 'Copy Output' },
  EXPORT: { key: 'e', mod: true, shift: true, label: 'Export' },
  TOGGLE_PREVIEW: { key: 'p', mod: true, label: 'Toggle Preview' },
  SHOW_SHORTCUTS: { key: '/', mod: true, label: 'Show Shortcuts' },
} as const;

export interface KeyboardShortcutsConfig {
  commandPalette?: () => void;
  convert?: () => void;
  copy?: () => void;
  export?: () => void;
  togglePreview?: () => void;
  showShortcuts?: () => void;
  escape?: () => void;
}

export function isMac(): boolean {
  if (typeof navigator === 'undefined') return false;
  const uaData = (navigator as Navigator & { userAgentData?: { platform: string } })
    .userAgentData;
  return (
    uaData?.platform === 'macOS' ||
    /Mac|iPod|iPhone|iPad/.test(navigator.userAgent)
  );
}

export function formatShortcut(shortcut: ShortcutDefinition): string {
  const mac = isMac();
  const parts: string[] = [];

  if (shortcut.mod) {
    parts.push(mac ? '⌘' : 'Ctrl');
  }
  if (shortcut.shift) {
    parts.push(mac ? '⇧' : 'Shift');
  }

  const keyDisplay = shortcut.key === 'Enter' ? '↵' : shortcut.key.toUpperCase();
  parts.push(keyDisplay);

  return mac ? parts.join('') : parts.join('+');
}

export function useKeyboardShortcuts(config: KeyboardShortcutsConfig): void {
  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      const mac = isMac();
      const modPressed = mac ? event.metaKey : event.ctrlKey;
      const key = event.key.toLowerCase();

      if (event.key === 'Escape' && config.escape) {
        config.escape();
        return;
      }

      if (modPressed && key === 'k' && !event.shiftKey && config.commandPalette) {
        event.preventDefault();
        config.commandPalette();
        return;
      }

      if (modPressed && event.key === 'Enter' && !event.shiftKey && config.convert) {
        event.preventDefault();
        config.convert();
        return;
      }

      if (modPressed && key === 'c' && event.shiftKey && config.copy) {
        event.preventDefault();
        config.copy();
        return;
      }

      if (modPressed && key === 'e' && event.shiftKey && config.export) {
        event.preventDefault();
        config.export();
        return;
      }

      if (modPressed && key === 'p' && !event.shiftKey && config.togglePreview) {
        event.preventDefault();
        config.togglePreview();
        return;
      }

      if (modPressed && event.key === '/' && !event.shiftKey && config.showShortcuts) {
        event.preventDefault();
        config.showShortcuts();
        return;
      }
    },
    [config]
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
    };
  }, [handleKeyDown]);
}
