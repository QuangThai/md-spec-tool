"use client";

import { motion } from "framer-motion";
import { Check, FileText, Globe, Sparkles, Table2, TestTube, Workflow } from "lucide-react";

interface TemplateCardProps {
  templates: Array<string | { id?: string; name?: string }>;
  selected: string;
  onSelect: (template: string) => void;
  compact?: boolean;
}

// Template metadata with preview snippets
const TEMPLATE_META: Record<string, {
  icon: typeof FileText;
  label: string;
  description: string;
  preview: string;
}> = {
  test_spec_v1: {
    icon: FileText,
    label: "Standard Spec",
    description: "Traditional table-based specification",
    preview: `## Feature: Login
### TC-001: Valid Login
**Steps:**
1. Enter credentials
2. Click login
**Expected:** Dashboard shown`,
  },
  bdd_scenarios_cards: {
    icon: Workflow,
    label: "BDD / Gherkin",
    description: "Given/When/Then behavioral specs",
    preview: `## Scenario: User Login
**Given** I am on the login page
**When** I enter valid credentials
**And** I click the login button
**Then** I should be redirected to the dashboard`,
  },
  test_case_cards: {
    icon: TestTube,
    label: "QA Test Cases",
    description: "Quality assurance test procedures",
    preview: `# Test Suite: Authentication
| ID | Title | Priority |
| TC-01 | Login with email | High |
| TC-02 | Password reset | Medium |`,
  },
  api_endpoints_cards: {
    icon: Globe,
    label: "API Documentation",
    description: "REST API endpoint specifications",
    preview: `## POST /api/v1/login
**Method:** POST
**Description:** User authentication endpoint
**Request:** { "user": "..." }
**Response:** 200 OK`,
  },
  ui_specs_cards: {
    icon: Table2,
    label: "UI Design Spec",
    description: "Interface component requirements",
    preview: `## Screen: User Profile
| Component | Type | Required |
| Avatar | Image | No |
| Display Name | Input | Yes |`,
  },
  requirements_cards: {
    icon: Sparkles,
    label: "User Requirements",
    description: "Business and functional requirements",
    preview: `## REQ-001: Mobile Login
**Description:** Users must be able to login via mobile
**Acceptance Criteria:**
- Works on iOS and Android
- Supports biometrics`,
  },
};

export function TemplateCards({ templates, selected, onSelect, compact = false }: TemplateCardProps) {
  if (compact) {
    return (
      <div className="flex flex-nowrap gap-3 overflow-x-auto custom-scrollbar pb-1 -mx-1 px-1">
        {templates.map((template, idx) => {
          const templateName = typeof template === "string" ? template : (template.name || template.id || `template-${idx}`);
          const templateId = typeof template === "string" ? template : (template.id || template.name || `template-${idx}`);
          const meta = TEMPLATE_META[templateName] || {
            icon: FileText,
            label: templateName,
            description: "",
            preview: "",
          };
          const Icon = meta.icon;
          const isSelected = selected === templateName;

          return (
            <motion.button
              key={templateId}
              type="button"
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
              onClick={() => onSelect(templateName)}
              className={`
                flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg border transition-all cursor-pointer shrink-0 whitespace-nowrap
                ${isSelected
                  ? "bg-orange-500/20 border-orange-500/50 text-white"
                  : "bg-white/5 border-white/10 text-white/70 hover:bg-white/10 hover:border-white/20"
                }
              `}
            >
              <Icon className={`w-3 h-3 shrink-0 ${isSelected ? "text-orange-400" : ""}`} />
              <span className="text-[9px] font-bold uppercase tracking-wider">{meta.label}</span>
              {isSelected && <Check className="w-2.5 h-2.5 shrink-0 text-accent-orange" />}
            </motion.button>
          );
        })}
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-3">
      {templates.map((template, idx) => {
        const templateName = typeof template === "string" ? template : (template.name || template.id || `template-${idx}`);
        const templateId = typeof template === "string" ? template : (template.id || template.name || `template-${idx}`);
        const meta = TEMPLATE_META[templateName] || {
          icon: FileText,
          label: templateName,
          description: "",
          preview: "",
        };
        const Icon = meta.icon;
        const isSelected = selected === templateName;

        return (
          <motion.button
            key={templateId}
            type="button"
            whileHover={{ scale: 1.02, y: -2 }}
            whileTap={{ scale: 0.98 }}
            onClick={() => onSelect(templateName)}
            className={`
              relative group text-left p-4 rounded-xl border transition-all cursor-pointer overflow-hidden
              ${isSelected
                ? "bg-accent-orange/10 border-accent-orange/40 shadow-lg shadow-accent-orange/10"
                : "bg-white/5 border-white/10 hover:bg-white/8 hover:border-white/20"
              }
            `}
          >
            {/* Selection indicator */}
            {isSelected && (
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                className="absolute top-2 right-2 w-5 h-5 rounded-full bg-accent-orange flex items-center justify-center"
              >
                <Check className="w-3 h-3 text-white" />
              </motion.div>
            )}

            {/* Header */}
            <div className="flex items-center gap-2 mb-2">
              <div className={`
                w-7 h-7 rounded-lg flex items-center justify-center
                ${isSelected ? "bg-accent-orange/20" : "bg-white/10"}
              `}>
                <Icon className={`w-4 h-4 ${isSelected ? "text-accent-orange" : "text-white/60"}`} />
              </div>
              <div>
                <h3 className={`text-[11px] font-black uppercase tracking-wider ${isSelected ? "text-white" : "text-white/80"}`}>
                  {meta.label}
                </h3>
                <p className="text-[9px] text-white/40">{meta.description}</p>
              </div>
            </div>

            {/* Preview */}
            <div className={`
              mt-3 p-2 rounded-lg bg-black/30 border border-white/5
              font-mono text-[8px] leading-relaxed overflow-hidden
              ${isSelected ? "text-white/70" : "text-white/50"}
            `}>
              <pre className="whitespace-pre-wrap line-clamp-6">
                {meta.preview}
              </pre>
            </div>

            {/* Hover gradient */}
            <div className={`
              absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none
              bg-linear-to-t from-accent-orange/5 to-transparent
            `} />
          </motion.button>
        );
      })}
    </div>
  );
}
