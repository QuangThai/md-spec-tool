"use client";

import {
  MDFlowWarning,
  ValidationCrossFieldRule,
  ValidationFormatRules,
  ValidationRules,
} from "@/lib/types";
import { useMDFlowTemplatesQuery, useValidatePasteMutation } from "@/lib/mdflowQueries";
import { useMDFlowActions, useMDFlowStore } from "@/lib/mdflowStore";
import {
  CANONICAL_FIELDS,
  generatePresetsFromTemplates,
  loadPresets,
  savePreset,
  deletePreset,
} from "@/lib/validationPresets";
import { AnimatePresence, motion } from "framer-motion";
import { useBodyScrollLock } from "@/lib/useBodyScrollLock";
import {
  Check,
  Plus,
  Save,
  Trash2,
  X,
  HelpCircle,
  AlertCircle,
  CheckCircle2,
} from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { WarningPanel } from "./WarningPanel";
import { Select } from "./ui/Select";
import { Tooltip } from "./ui/Tooltip";

interface ValidationConfiguratorProps {
  open: boolean;
  onClose: () => void;
  showValidateAction?: boolean;
}

type TabType = "basic" | "advanced" | "presets";

const emptyRules: ValidationRules = {
  required_fields: [],
  format_rules: null,
  cross_field: [],
};

const HELP_TEXT: Record<string, string> = {
  required: "Mark fields that must be present in the data",
  format: "Configure format validation like email, URL, or custom patterns",
  crossField: "Set up conditional rules between fields (e.g., when X is set, Y is required)",
  idPattern: "Regex pattern to validate IDs. Example: ^[A-Z]{2,}-\\d+$ for IDS-001",
  emailFields: "Select fields to validate as email addresses",
  urlFields: "Select fields to validate as URLs",
};

export function ValidationConfigurator({
  open,
  onClose,
  showValidateAction = true,
}: ValidationConfiguratorProps) {
  useBodyScrollLock(open);
  const validationRules = useMDFlowStore((s) => s.validationRules);
  const pasteText = useMDFlowStore((s) => s.pasteText);
  const { setValidationRules } = useMDFlowActions();
  const [localRules, setLocalRules] = useState<ValidationRules>(validationRules);
  const [activeTab, setActiveTab] = useState<TabType>("basic");
  const [presets, setPresets] = useState(loadPresets());
  const [presetName, setPresetName] = useState("");
  const [validating, setValidating] = useState(false);
  const [validationWarnings, setValidationWarnings] = useState<MDFlowWarning[]>([]);
  const validateMutation = useValidatePasteMutation();

  // Fetch templates from API to generate dynamic presets
  const { data: templates = [] } = useMDFlowTemplatesQuery();
  const templatePresets = useMemo(
    () => generatePresetsFromTemplates(templates),
    [templates]
  );

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
    setLocalRules({ ...localRules, format_rules: next });
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
    if (preset) {
      const { id, name, createdAt, ...rules } = preset;
      setLocalRules(rules as ValidationRules);
    }
  };

  const loadDefaultPreset = (name: string) => {
    const preset = templatePresets.find((p) => p.name === name);
    if (preset) setLocalRules(preset.rules);
  };

  const saveCurrentPreset = () => {
    if (!presetName.trim()) return;
    savePreset({ name: presetName.trim(), ...localRules });
    setPresets(loadPresets());
    setPresetName("");
  };

  const handleValidate = async () => {
    if (!pasteText.trim()) return;
    setValidating(true);
    setValidationWarnings([]);
    try {
      const result = await validateMutation.mutateAsync({
        pasteText,
        rules: localRules,
      });
      setValidationWarnings(result.warnings || []);
    } finally {
      setValidating(false);
    }
  };

  const resetToEmpty = () => setLocalRules(emptyRules);

  const hasAnyRules =
    localRules.required_fields.length > 0 ||
    localRules.format_rules ||
    (localRules.cross_field || []).length > 0;

  if (!open) return null;

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 z-50 flex items-center justify-center p-2 sm:p-4 bg-black/60 backdrop-blur-sm"
        onClick={onClose}
      >
        <motion.div
          initial={{ opacity: 0, scale: 0.96, y: 8 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.96, y: 8 }}
          transition={{ duration: 0.2 }}
          className="w-full max-w-2xl max-h-[95vh] sm:max-h-[90vh] overflow-hidden rounded-xl sm:rounded-2xl bg-[#0d0d0d] border border-white/10 shadow-2xl flex flex-col"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header */}
          <div className="flex items-center justify-between gap-3 px-4 sm:px-6 py-4 border-b border-white/10 bg-linear-to-b from-white/5 to-transparent shrink-0">
            <div>
              <h2 className="text-sm sm:text-base font-bold text-white">
                Validation Rules
              </h2>
              <p className="text-xs text-white/50 mt-1">
                {hasAnyRules ? (
                  <span className="flex items-center gap-2">
                    <CheckCircle2 className="w-3 h-3 text-green-400" />
                    {localRules.required_fields.length} required
                    {localRules.format_rules ? ", format rules" : ""}
                    {(localRules.cross_field || []).length > 0 ? ", conditional rules" : ""}
                  </span>
                ) : (
                  "No rules configured"
                )}
              </p>
            </div>
            <button
              type="button"
              onClick={onClose}
              className="p-2 rounded-lg text-white/50 hover:text-white hover:bg-white/10 transition-colors cursor-pointer shrink-0"
              aria-label="Close"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          {/* Simplified Tab Navigation - 3 tabs only */}
          <div className="flex gap-1 px-4 sm:px-6 py-2 border-b border-white/10 shrink-0 bg-white/2">
            {(["basic", "advanced", "presets"] as TabType[]).map((tab) => {
              const isActive = activeTab === tab;
              const tabConfig = {
                basic: { label: "Basic Rules", icon: "•" },
                advanced: { label: "Conditional", icon: "→" },
                presets: { label: "Presets", icon: "★" },
              };

              return (
                <button
                  key={tab}
                  type="button"
                  onClick={() => setActiveTab(tab)}
                  className={`px-4 py-2.5 rounded-lg text-sm font-medium transition-all cursor-pointer ${
                    isActive
                      ? "bg-accent-orange/20 text-accent-orange border border-accent-orange/30"
                      : "bg-transparent text-white/60 hover:text-white hover:bg-white/5 border border-transparent"
                  }`}
                >
                  <span className="mr-2">{tabConfig[tab].icon}</span>
                  {tabConfig[tab].label}
                </button>
              );
            })}
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto custom-scrollbar">
            {/* Basic Rules Tab - Required + Format combined */}
            {activeTab === "basic" && (
              <div className="px-4 sm:px-6 py-5 space-y-5">
                {/* Required Fields Section */}
                <section>
                  <div className="flex items-center gap-2 mb-3">
                    <h3 className="text-sm font-semibold text-white">Required Fields</h3>
                    <Tooltip content={HELP_TEXT.required}>
                      <HelpCircle className="w-4 h-4 text-white/40 hover:text-white/60 cursor-help transition-colors" />
                    </Tooltip>
                  </div>
                  <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
                    {CANONICAL_FIELDS.map(({ value, label }) => {
                      const checked = localRules.required_fields.includes(value);
                      return (
                        <label
                          key={value}
                          className={`flex items-center gap-2.5 px-3 py-2.5 rounded-lg border cursor-pointer transition-all ${
                            checked
                              ? "bg-accent-orange/15 border-accent-orange/40 text-accent-orange"
                              : "bg-white/5 border-white/10 text-white/70 hover:bg-white/8"
                          }`}
                        >
                          <input
                            type="checkbox"
                            checked={checked}
                            onChange={() => toggleRequired(value)}
                            className="sr-only"
                          />
                          {checked ? (
                            <Check className="w-4 h-4 shrink-0 font-bold" />
                          ) : (
                            <span className="w-4 h-4 rounded border border-white/30 shrink-0" />
                          )}
                          <span className="text-xs font-medium">{label}</span>
                        </label>
                      );
                    })}
                  </div>
                </section>

                {/* Format Validation Section */}
                <section className="pt-3 border-t border-white/10">
                  <div className="flex items-center gap-2 mb-4">
                    <h3 className="text-sm font-semibold text-white">Format Validation</h3>
                    <Tooltip content={HELP_TEXT.format}>
                      <HelpCircle className="w-4 h-4 text-white/40 hover:text-white/60 cursor-help transition-colors" />
                    </Tooltip>
                  </div>

                  <div className="space-y-4">
                    {/* ID Pattern */}
                    <div>
                      <label className="block text-xs font-semibold text-white/70 mb-2">
                        ID Pattern
                      </label>
                      <input
                        type="text"
                        value={localRules.format_rules?.id_pattern || ""}
                        onChange={(e) =>
                          setFormat("id_pattern", e.target.value || undefined)
                        }
                        placeholder="e.g. ^[A-Z]{2,}-\\d+$"
                        className="input w-full h-9 text-xs font-mono rounded-lg bg-white/5 border border-white/10 text-white placeholder-white/30 focus:border-accent-orange/50 focus:bg-white/8 transition-colors"
                      />
                      <p className="text-xs text-white/40 mt-1">
                        {HELP_TEXT.idPattern}
                      </p>
                    </div>

                    {/* Email Fields */}
                    <div>
                      <label className="block text-xs font-semibold text-white/70 mb-2">
                        Email Validation
                      </label>
                      <div className="flex flex-wrap gap-2">
                        {CANONICAL_FIELDS.map(({ value, label }) => {
                          const checked = (
                            localRules.format_rules?.email_fields || []
                          ).includes(value);
                          return (
                            <label
                              key={value}
                              className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border cursor-pointer transition-all text-xs ${
                                checked
                                  ? "bg-accent-orange/15 border-accent-orange/40 text-accent-orange"
                                  : "bg-white/5 border-white/10 text-white/60 hover:bg-white/8"
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
                              {checked ? (
                                <Check className="w-3 h-3" />
                              ) : (
                                <span className="w-3 h-3 rounded border border-white/30" />
                              )}
                              {label}
                            </label>
                          );
                        })}
                      </div>
                    </div>

                    {/* URL Fields */}
                    <div>
                      <label className="block text-xs font-semibold text-white/70 mb-2">
                        URL Validation
                      </label>
                      <div className="flex flex-wrap gap-2">
                        {CANONICAL_FIELDS.map(({ value, label }) => {
                          const checked = (
                            localRules.format_rules?.url_fields || []
                          ).includes(value);
                          return (
                            <label
                              key={value}
                              className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border cursor-pointer transition-all text-xs ${
                                checked
                                  ? "bg-accent-orange/15 border-accent-orange/40 text-accent-orange"
                                  : "bg-white/5 border-white/10 text-white/60 hover:bg-white/8"
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
                              {checked ? (
                                <Check className="w-3 h-3" />
                              ) : (
                                <span className="w-3 h-3 rounded border border-white/30" />
                              )}
                              {label}
                            </label>
                          );
                        })}
                      </div>
                    </div>
                  </div>
                </section>
              </div>
            )}

            {/* Advanced Rules Tab - Conditional only */}
            {activeTab === "advanced" && (
              <div className="px-4 sm:px-6 py-5">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-2">
                    <h3 className="text-sm font-semibold text-white">Conditional Rules</h3>
                    <Tooltip content={HELP_TEXT.crossField}>
                      <HelpCircle className="w-4 h-4 text-white/40 hover:text-white/60 cursor-help transition-colors" />
                    </Tooltip>
                  </div>
                  <button
                    type="button"
                    onClick={addCrossField}
                    className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold text-accent-orange hover:bg-accent-orange/10 border border-accent-orange/30 transition-all cursor-pointer"
                  >
                    <Plus className="w-3.5 h-3.5" />
                    Add
                  </button>
                </div>

                <p className="text-xs text-white/50 mb-4 leading-relaxed">
                  Create rules like: "When <strong>ID</strong> is set, then <strong>Feature</strong> becomes required"
                </p>

                {(localRules.cross_field || []).length === 0 ? (
                  <div className="p-4 rounded-lg bg-white/3 border border-white/10 text-center">
                    <AlertCircle className="w-5 h-5 mx-auto text-white/30 mb-2" />
                    <p className="text-xs text-white/50">
                      No conditional rules yet. Click "Add" to create one.
                    </p>
                  </div>
                ) : (
                  <div className="space-y-3">
                    {(localRules.cross_field || []).map((rule, index) => (
                      <motion.div
                        key={index}
                        initial={{ opacity: 0, y: -8 }}
                        animate={{ opacity: 1, y: 0 }}
                        exit={{ opacity: 0, y: -8 }}
                        className="flex flex-col sm:flex-row items-start sm:items-center gap-3 p-3 rounded-lg bg-white/5 border border-white/10 hover:bg-white/8 group transition-colors"
                      >
                        <div className="flex flex-wrap items-center gap-2 flex-1">
                          <span className="text-xs font-semibold text-white/40 uppercase">When</span>
                          <Select
                            value={rule.if_field}
                            onValueChange={(value) =>
                              updateCrossField(index, { if_field: value })
                            }
                            options={CANONICAL_FIELDS}
                            size="compact"
                            className="h-8 text-xs"
                            aria-label="If field"
                          />
                          <span className="text-white/40 text-xs font-bold">→</span>
                          <span className="text-xs font-semibold text-white/40 uppercase">Then require</span>
                          <Select
                            value={rule.then_field}
                            onValueChange={(value) =>
                              updateCrossField(index, { then_field: value })
                            }
                            options={CANONICAL_FIELDS}
                            size="compact"
                            className="h-8 text-xs"
                            aria-label="Then field"
                          />
                        </div>
                        <button
                          type="button"
                          onClick={() => removeCrossField(index)}
                          className="p-2 rounded text-white/30 hover:text-red-400 hover:bg-red-400/10 transition-all cursor-pointer sm:group-hover:opacity-100"
                          aria-label="Remove rule"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </motion.div>
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* Presets Tab */}
            {activeTab === "presets" && (
              <div className="px-4 sm:px-6 py-5 space-y-5">
                {/* Template Presets */}
                {templatePresets.length > 0 && (
                  <section>
                    <h3 className="text-sm font-semibold text-white mb-3">
                      Quick Templates
                    </h3>
                    <div className="grid grid-cols-2 sm:grid-cols-2 gap-2">
                      {templatePresets.map((p) => (
                        <motion.button
                          key={p.name}
                          type="button"
                          whileHover={{ scale: 1.02 }}
                          whileTap={{ scale: 0.98 }}
                          onClick={() => loadDefaultPreset(p.name)}
                          className="p-3 rounded-lg text-center text-xs font-semibold bg-white/5 hover:bg-accent-orange/15 border border-white/10 hover:border-accent-orange/30 text-white/80 hover:text-accent-orange transition-all cursor-pointer"
                        >
                          {p.name}
                        </motion.button>
                      ))}
                    </div>
                  </section>
                )}

                {/* Saved Presets */}
                <section className={templatePresets.length > 0 ? "border-t border-white/10 pt-5" : ""}>
                  <h3 className="text-sm font-semibold text-white mb-3">
                    Saved Presets
                  </h3>
                  {presets.length === 0 ? (
                    <div className="p-4 rounded-lg bg-white/3 border border-white/10 text-center">
                      <AlertCircle className="w-5 h-5 mx-auto text-white/30 mb-2" />
                      <p className="text-xs text-white/50">
                        Save your first preset to get started
                      </p>
                    </div>
                  ) : (
                    <div className="grid grid-cols-1 gap-2 mb-4">
                      {presets.map((p) => (
                        <div
                          key={p.id}
                          className="flex items-center justify-between gap-3 p-3 rounded-lg bg-white/5 hover:bg-white/8 border border-white/10 group transition-all"
                        >
                          <button
                            type="button"
                            onClick={() => loadPresetRules(p.id)}
                            className="flex-1 text-left text-xs font-medium text-white/80 hover:text-accent-orange transition-colors cursor-pointer"
                          >
                            {p.name}
                          </button>
                          <button
                            type="button"
                            onClick={() => {
                              deletePreset(p.id);
                              setPresets(loadPresets());
                            }}
                            className="p-2 rounded text-white/30 hover:text-red-400 hover:bg-red-400/10 transition-all cursor-pointer"
                            aria-label="Delete preset"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                </section>

                {/* Save Current as Preset */}
                <section className="border-t border-white/10 pt-5">
                  <label className="text-sm font-semibold text-white block mb-3">
                    Save Current Rules
                  </label>
                  <div className="flex flex-col sm:flex-row gap-2">
                    <input
                      type="text"
                      value={presetName}
                      onChange={(e) => setPresetName(e.target.value)}
                      onKeyDown={(e) => {
                        if (e.key === "Enter" && presetName.trim() && hasAnyRules) {
                          saveCurrentPreset();
                        }
                      }}
                      placeholder="e.g., Email & ID validation"
                      className="input flex-1 h-9 text-xs rounded-lg bg-white/5 border border-white/10 text-white placeholder-white/30 focus:border-accent-orange/50 focus:bg-white/8 transition-colors"
                    />
                    <button
                      type="button"
                      onClick={saveCurrentPreset}
                      disabled={!presetName.trim() || !hasAnyRules}
                      className="flex items-center justify-center gap-2 px-4 py-2 rounded-lg text-xs font-semibold bg-accent-orange/20 hover:bg-accent-orange/30 border border-accent-orange/30 text-accent-orange disabled:opacity-50 disabled:cursor-not-allowed transition-all shrink-0"
                    >
                      <Save className="w-3.5 h-3.5" />
                      Save
                    </button>
                  </div>
                  {!hasAnyRules && (
                    <p className="text-xs text-white/40 mt-2">
                      ℹ️ Configure some rules first to save
                    </p>
                  )}
                </section>
              </div>
            )}

            {/* Validation Results */}
            {validationWarnings.length > 0 && (
              <div className="border-t border-white/10 px-4 sm:px-6 py-4 bg-red-950/20">
                <h3 className="text-sm font-semibold text-white mb-3">
                  Validation Results ({validationWarnings.length} issue
                  {validationWarnings.length !== 1 ? "s" : ""})
                </h3>
                <WarningPanel warnings={validationWarnings} />
              </div>
            )}
          </div>

          {/* Footer - Simplified Actions */}
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between px-4 sm:px-6 py-4 border-t border-white/10 bg-linear-to-t from-white/3 to-transparent shrink-0">
            <button
              type="button"
              onClick={resetToEmpty}
              className="px-4 py-2 text-xs font-semibold text-white/50 hover:text-white/70 transition-colors cursor-pointer order-2 sm:order-1"
            >
              Reset
            </button>
            <div className="order-1 sm:order-2 flex flex-col xs:flex-row gap-2 w-full sm:w-74">
              {showValidateAction && (
                <motion.button
                  type="button"
                  onClick={handleValidate}
                  disabled={!pasteText.trim() || validating}
                  whileHover={{ scale: 1.02 }}
                  whileTap={{ scale: 0.98 }}
                  className="flex-1 px-4 py-2 rounded-lg text-xs font-semibold bg-white/10 hover:bg-white/15 border border-white/20 text-white/80 disabled:opacity-50 disabled:cursor-not-allowed transition-all cursor-pointer"
                >
                  {validating ? "Testing…" : "Test Rules"}
                </motion.button>
              )}
              <motion.button
                type="button"
                onClick={applyToStore}
                whileHover={{ scale: 1.02 }}
                whileTap={{ scale: 0.98 }}
                className="flex-1 px-4 py-2 rounded-lg text-xs font-semibold bg-accent-orange/20 hover:bg-accent-orange/30 border border-accent-orange/30 text-accent-orange transition-all cursor-pointer"
              >
                Apply & Close
              </motion.button>
            </div>
          </div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
}
