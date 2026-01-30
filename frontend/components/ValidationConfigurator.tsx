"use client";

import {
  ValidationCrossFieldRule,
  ValidationFormatRules,
  ValidationRules,
  validatePaste,
} from "@/lib/mdflowApi";
import { useMDFlowStore } from "@/lib/mdflowStore";
import {
  CANONICAL_FIELDS,
  DEFAULT_PRESETS,
  loadPresets,
  savePreset,
  deletePreset,
} from "@/lib/validationPresets";
import { AnimatePresence, motion } from "framer-motion";
import {
  Check,
  ChevronDown,
  Link2,
  Plus,
  Save,
  Trash2,
  X,
} from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { WarningPanel } from "./WarningPanel";

interface ValidationConfiguratorProps {
  open: boolean;
  onClose: () => void;
  /** When true, show Validate button and run validation on current paste */
  showValidateAction?: boolean;
}

const emptyRules: ValidationRules = {
  required_fields: [],
  format_rules: null,
  cross_field: [],
};

export function ValidationConfigurator({
  open,
  onClose,
  showValidateAction = true,
}: ValidationConfiguratorProps) {
  const { validationRules, setValidationRules, pasteText } = useMDFlowStore();
  const [localRules, setLocalRules] = useState<ValidationRules>(validationRules);
  const [presets, setPresets] = useState(loadPresets());
  const [presetName, setPresetName] = useState("");
  const [validating, setValidating] = useState(false);
  const [validationWarnings, setValidationWarnings] = useState<any[]>([]);
  const [presetDropdownOpen, setPresetDropdownOpen] = useState(false);

  useEffect(() => {
    setLocalRules(validationRules);
  }, [validationRules, open]);

  useEffect(() => {
    if (open) setPresets(loadPresets());
  }, [open]);

  const applyToStore = useCallback(() => {
    setValidationRules(localRules);
    onClose();
  }, [localRules, setValidationRules, onClose]);

  const toggleRequired = (field: string) => {
    const set = new Set(localRules.required_fields);
    if (set.has(field)) set.delete(field);
    else set.add(field);
    setLocalRules({ ...localRules, required_fields: Array.from(set) });
  };

  const setFormatRules = (next: ValidationFormatRules | null) => {
    setLocalRules({ ...localRules, format_rules: next || undefined });
  };

  const setFormat = <K extends keyof ValidationFormatRules>(
    key: K,
    value: ValidationFormatRules[K]
  ) => {
    setFormatRules({
      ...(localRules.format_rules || {}),
      [key]: value,
    });
  };

  const toggleFormatList = (
    key: "email_fields" | "url_fields",
    field: string
  ) => {
    const current = localRules.format_rules?.[key] || [];
    const next = current.includes(field)
      ? current.filter((f) => f !== field)
      : [...current, field];
    setFormat(key, next);
  };

  const addCrossField = () => {
    setLocalRules({
      ...localRules,
      cross_field: [
        ...(localRules.cross_field || []),
        { if_field: "id", then_field: "feature" },
      ],
    });
  };

  const updateCrossField = (
    index: number,
    patch: Partial<ValidationCrossFieldRule>
  ) => {
    const list = [...(localRules.cross_field || [])];
    list[index] = { ...list[index], ...patch };
    setLocalRules({ ...localRules, cross_field: list });
  };

  const removeCrossField = (index: number) => {
    setLocalRules({
      ...localRules,
      cross_field: (localRules.cross_field || []).filter((_, i) => i !== index),
    });
  };

  const loadPresetRules = (presetId: string) => {
    const preset = presets.find((p) => p.id === presetId);
    if (preset) setLocalRules(preset.rules);
    setPresetDropdownOpen(false);
  };

  const loadDefaultPreset = (name: string) => {
    const preset = DEFAULT_PRESETS.find((p) => p.name === name);
    if (preset) setLocalRules(preset.rules);
    setPresetDropdownOpen(false);
  };

  const saveCurrentPreset = () => {
    if (!presetName.trim()) return;
    const saved = savePreset({ name: presetName.trim(), rules: localRules });
    setPresets(loadPresets());
    setPresetName("");
  };

  const handleValidate = async () => {
    if (!pasteText.trim()) return;
    setValidating(true);
    setValidationWarnings([]);
    const result = await validatePaste(pasteText, localRules);
    setValidating(false);
    if (result.data) {
      setValidationWarnings(result.data.warnings || []);
    }
  };

  const resetToEmpty = () => setLocalRules(emptyRules);

  if (!open) return null;

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      >
        <motion.div
          initial={{ opacity: 0, scale: 0.96, y: 8 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.96, y: 8 }}
          transition={{ duration: 0.2 }}
          className="w-full max-w-2xl max-h-[90vh] overflow-hidden rounded-2xl bg-[#0d0d0d] border border-white/10 shadow-2xl flex flex-col"
          onClick={(e) => e.stopPropagation()}
        >
          <div className="flex items-center justify-between gap-3 px-5 py-4 border-b border-white/10 bg-white/5 shrink-0">
            <div className="flex items-center gap-2">
              <Link2 className="w-5 h-5 text-accent-orange" />
              <h2 className="text-sm font-black uppercase tracking-widest text-white">
                Validation Rules
              </h2>
            </div>
            <button
              type="button"
              onClick={onClose}
              className="p-2 rounded-lg text-white/50 hover:text-white hover:bg-white/10 transition-colors cursor-pointer"
              aria-label="Close"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          <div className="flex-1 overflow-y-auto px-5 py-4 space-y-6 custom-scrollbar">
            {/* Presets */}
            <section>
              <label className="label mb-2 block">Presets</label>
              <div className="flex flex-wrap gap-2">
                {DEFAULT_PRESETS.map((p) => (
                  <button
                    key={p.name}
                    type="button"
                    onClick={() => loadDefaultPreset(p.name)}
                    className="px-3 py-1.5 rounded-lg text-[10px] font-bold uppercase tracking-wider bg-white/5 hover:bg-accent-orange/20 border border-white/10 hover:border-accent-orange/30 text-white/80 hover:text-accent-orange transition-all"
                  >
                    {p.name}
                  </button>
                ))}
                <div className="relative">
                  <button
                    type="button"
                    onClick={() => setPresetDropdownOpen((v) => !v)}
                    className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[10px] font-bold uppercase tracking-wider bg-white/5 hover:bg-white/10 border border-white/10 text-white/70"
                  >
                    Saved
                    <ChevronDown
                      className={`w-3 h-3 transition-transform ${presetDropdownOpen ? "rotate-180" : ""}`}
                    />
                  </button>
                  <AnimatePresence>
                    {presetDropdownOpen && (
                      <motion.div
                        initial={{ opacity: 0, y: -4 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -4 }}
                        className="absolute top-full left-0 mt-1 w-64 max-w-[16rem] rounded-lg bg-[#141414] border border-white/10 shadow-xl z-10 overflow-hidden box-border"
                      >
                        <div className="w-full min-w-0 overflow-hidden">
                          {presets.length === 0 ? (
                            <p className="px-3 py-2 text-[10px] text-white/50">
                              No saved presets
                            </p>
                          ) : (
                            presets.map((p) => (
                              <div
                                key={p.id}
                                className="group relative flex items-center justify-between gap-2 px-3 py-2 min-w-0 overflow-hidden rounded"
                              >
                                <span
                                  className="pointer-events-none absolute inset-0 rounded bg-white/5 opacity-0 transition-opacity group-hover:opacity-100"
                                  aria-hidden
                                />
                                <button
                                  type="button"
                                  onClick={() => loadPresetRules(p.id)}
                                  className="relative z-10 text-[11px] font-medium text-white/90 truncate flex-1 min-w-0 text-left"
                                >
                                  {p.name}
                                </button>
                                <button
                                  type="button"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    deletePreset(p.id);
                                    setPresets(loadPresets());
                                  }}
                                  className="relative z-10 p-1 rounded text-white/40 hover:text-accent-red opacity-0 group-hover:opacity-100 transition-opacity shrink-0"
                                  aria-label="Delete preset"
                                >
                                  <Trash2 className="w-3 h-3" />
                                </button>
                              </div>
                            ))
                          )}
                        </div>
                      </motion.div>
                    )}
                  </AnimatePresence>
                </div>
              </div>
              <div className="flex gap-2 mt-2">
                <input
                  type="text"
                  value={presetName}
                  onChange={(e) => setPresetName(e.target.value)}
                  placeholder="Preset name"
                  className="input flex-1 h-9 text-xs"
                />
                <button
                  type="button"
                  onClick={saveCurrentPreset}
                  disabled={!presetName.trim()}
                  className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-[10px] font-bold uppercase tracking-wider bg-accent-orange/20 hover:bg-accent-orange/30 border border-accent-orange/30 text-accent-orange disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <Save className="w-3 h-3" />
                  Save
                </button>
              </div>
            </section>

            {/* Required fields */}
            <section>
              <label className="label mb-2 block">Required fields</label>
              <div className="flex flex-wrap gap-2">
                {CANONICAL_FIELDS.map(({ value, label }) => {
                  const checked = localRules.required_fields.includes(value);
                  return (
                    <label
                      key={value}
                      className={`flex items-center gap-2 px-3 py-2 rounded-lg border cursor-pointer transition-all ${
                        checked
                          ? "bg-accent-orange/10 border-accent-orange/30 text-accent-orange"
                          : "bg-white/5 border-white/10 text-white/70 hover:bg-white/10"
                      }`}
                    >
                      <input
                        type="checkbox"
                        checked={checked}
                        onChange={() => toggleRequired(value)}
                        className="sr-only"
                      />
                      {checked ? (
                        <Check className="w-3.5 h-3.5 shrink-0" />
                      ) : (
                        <span className="w-3.5 h-3.5 rounded border border-current shrink-0" />
                      )}
                      <span className="text-[11px] font-medium">{label}</span>
                    </label>
                  );
                })}
              </div>
            </section>

            {/* Format rules */}
            <section>
              <label className="label mb-2 block">Format validation</label>
              <div className="space-y-3 p-4 rounded-xl bg-white/5 border border-white/10">
                <div>
                  <span className="text-[10px] text-white/50 uppercase font-bold tracking-wider block mb-1">
                    ID pattern (regex)
                  </span>
                  <input
                    type="text"
                    value={localRules.format_rules?.id_pattern || ""}
                    onChange={(e) =>
                      setFormat("id_pattern", e.target.value || undefined)
                    }
                    placeholder="e.g. ^[A-Z]{2,}-\\d+$"
                    className="input w-full h-9 text-xs font-mono"
                  />
                </div>
                <div>
                  <span className="text-[10px] text-white/50 uppercase font-bold tracking-wider block mb-1">
                    Fields to validate as email
                  </span>
                  <div className="flex flex-wrap gap-1.5">
                    {CANONICAL_FIELDS.slice(0, 8).map(({ value, label }) => {
                      const checked = (
                        localRules.format_rules?.email_fields || []
                      ).includes(value);
                      return (
                        <label
                          key={value}
                          className={`inline-flex items-center gap-1.5 px-2 py-1 rounded text-[10px] cursor-pointer ${
                            checked
                              ? "bg-accent-orange/20 text-accent-orange"
                              : "bg-white/5 text-white/60 hover:bg-white/10"
                          }`}
                        >
                          <input
                            type="checkbox"
                            checked={checked}
                            onChange={() =>
                              toggleFormatList("email_fields", value)
                            }
                            className="sr-only"
                          />
                          {label}
                        </label>
                      );
                    })}
                  </div>
                </div>
                <div>
                  <span className="text-[10px] text-white/50 uppercase font-bold tracking-wider block mb-1">
                    Fields to validate as URL
                  </span>
                  <div className="flex flex-wrap gap-1.5">
                    {CANONICAL_FIELDS.filter((f) =>
                      ["endpoint", "notes", "feature"].includes(f.value)
                    ).map(({ value, label }) => {
                      const checked = (
                        localRules.format_rules?.url_fields || []
                      ).includes(value);
                      return (
                        <label
                          key={value}
                          className={`inline-flex items-center gap-1.5 px-2 py-1 rounded text-[10px] cursor-pointer ${
                            checked
                              ? "bg-accent-orange/20 text-accent-orange"
                              : "bg-white/5 text-white/60 hover:bg-white/10"
                          }`}
                        >
                          <input
                            type="checkbox"
                            checked={checked}
                            onChange={() =>
                              toggleFormatList("url_fields", value)
                            }
                            className="sr-only"
                          />
                          {label}
                        </label>
                      );
                    })}
                  </div>
                </div>
              </div>
            </section>

            {/* Cross-field rules */}
            <section>
              <div className="flex items-center justify-between mb-2">
                <label className="label">Cross-field rules</label>
                <button
                  type="button"
                  onClick={addCrossField}
                  className="flex items-center gap-1.5 px-2 py-1 rounded-lg text-[10px] font-bold uppercase tracking-wider text-accent-orange hover:bg-accent-orange/10 border border-accent-orange/20"
                >
                  <Plus className="w-3 h-3" />
                  Add
                </button>
              </div>
              <p className="text-[10px] text-white/50 mb-2">
                When &quot;If field&quot; is set, &quot;Then field&quot; is
                required.
              </p>
              <div className="space-y-2">
                {(localRules.cross_field || []).map((rule, index) => (
                  <div
                    key={index}
                    className="flex flex-wrap items-center gap-2 p-3 rounded-lg bg-white/5 border border-white/10"
                  >
                    <select
                      value={rule.if_field}
                      onChange={(e) =>
                        updateCrossField(index, {
                          if_field: e.target.value,
                        })
                      }
                      className="input h-8 text-[11px] w-28"
                    >
                      {CANONICAL_FIELDS.map((f) => (
                        <option key={f.value} value={f.value}>
                          {f.label}
                        </option>
                      ))}
                    </select>
                    <span className="text-white/50 text-[10px]">→</span>
                    <select
                      value={rule.then_field}
                      onChange={(e) =>
                        updateCrossField(index, {
                          then_field: e.target.value,
                        })
                      }
                      className="input h-8 text-[11px] w-28"
                    >
                      {CANONICAL_FIELDS.map((f) => (
                        <option key={f.value} value={f.value}>
                          {f.label}
                        </option>
                      ))}
                    </select>
                    <button
                      type="button"
                      onClick={() => removeCrossField(index)}
                      className="p-1.5 rounded text-white/40 hover:text-accent-red hover:bg-white/5"
                      aria-label="Remove rule"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                ))}
                {(localRules.cross_field || []).length === 0 && (
                  <p className="text-[10px] text-white/40 italic">
                    No cross-field rules. Click Add to require a field when
                    another is set.
                  </p>
                )}
              </div>
            </section>

            {/* Validation result */}
            {validationWarnings.length > 0 && (
              <section>
                <label className="label mb-2 block">
                  Validation result ({validationWarnings.length} issue
                  {validationWarnings.length !== 1 ? "s" : ""})
                </label>
                <WarningPanel warnings={validationWarnings} />
              </section>
            )}
          </div>

          <div className="flex items-center justify-between gap-3 px-5 py-4 border-t border-white/10 bg-white/5 shrink-0">
            <button
              type="button"
              onClick={resetToEmpty}
              className="text-[10px] font-bold uppercase tracking-wider text-white/50 hover:text-white/70"
            >
              Reset to empty
            </button>
            <div className="flex gap-2">
              {showValidateAction && (
                <button
                  type="button"
                  onClick={handleValidate}
                  disabled={!pasteText.trim() || validating}
                  className="px-4 py-2 rounded-lg text-[10px] font-bold uppercase tracking-wider bg-white/10 hover:bg-white/15 border border-white/10 text-white/80 disabled:opacity-50"
                >
                  {validating ? "Validating…" : "Validate"}
                </button>
              )}
              <button
                type="button"
                onClick={applyToStore}
                className="px-4 py-2 rounded-lg text-[10px] font-bold uppercase tracking-wider bg-accent-orange/20 hover:bg-accent-orange/30 border border-accent-orange/30 text-accent-orange"
              >
                Apply & close
              </button>
            </div>
          </div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
}
