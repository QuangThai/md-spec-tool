"use client";

import { AnimatePresence, motion } from "framer-motion";
import {
  AlertCircle,
  BookOpen,
  Braces,
  Check,
  ChevronDown,
  ChevronRight,
  Code2,
  Copy,
  Download,
  Eye,
  FileCode,
  FileText,
  Lightbulb,
  Loader2,
  Play,
  Plus,
  Save,
  Sparkles,
  Trash2,
  Upload,
  Variable,
  X,
} from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  getTemplateContent,
  getTemplateInfo,
  previewTemplate,
  TemplateFunction,
  TemplateInfo,
  TemplateVariable,
} from "@/lib/mdflowApi";

interface CustomTemplate {
  id: string;
  name: string;
  content: string;
  createdAt: number;
  updatedAt: number;
}

interface TemplateEditorProps {
  isOpen: boolean;
  onClose: () => void;
  onSaveTemplate?: (template: CustomTemplate) => void;
  currentSampleData?: string;
}

const STORAGE_KEY = "mdflow-custom-templates";

// Built-in template names
const BUILT_IN_TEMPLATES = [
  { id: "default", name: "Default", description: "Standard test case format" },
  { id: "feature-spec", name: "Feature Spec", description: "User story format" },
  { id: "test-plan", name: "Test Plan", description: "QA test plan format" },
  { id: "api-endpoint", name: "API Endpoint", description: "API documentation" },
  { id: "spec-table", name: "Spec Table", description: "UI specification table" },
];

// Default sample data for preview
const DEFAULT_SAMPLE_DATA = `Feature	Scenario	Instructions	Expected	Priority	Type	Notes
User Authentication	Valid Login	1. Enter username
2. Enter password
3. Click login button	Dashboard should display with user name	High	Positive	Core feature
User Authentication	Invalid Password	1. Enter valid username
2. Enter wrong password
3. Click login button	Error message: "Invalid credentials"	High	Negative	Security test
Profile Management	Update Profile	1. Go to settings
2. Change display name
3. Click save	Profile updated successfully message shown	Medium	Positive	`;

// Simple starter template
const STARTER_TEMPLATE = `---
name: "{{.Title}}"
version: "1.0"
generated_at: "{{.GeneratedAt}}"
---

# {{.Title}}

This specification contains {{.TotalCount}} items.

{{range .FeatureGroups}}
## {{.Feature}}
{{range .Rows}}
### {{if .ID}}{{.ID}}: {{end}}{{.Scenario}}
{{- if .Priority}}

**Priority:** {{.Priority}}
{{- end}}
{{- if notEmpty .Instructions}}

**Steps:**
{{formatSteps .Instructions}}
{{- end}}
{{- if notEmpty .Expected}}

**Expected:**
{{.Expected}}
{{- end}}

---
{{end}}
{{end}}
`;

export function TemplateEditor({
  isOpen,
  onClose,
  onSaveTemplate,
  currentSampleData,
}: TemplateEditorProps) {
  // Editor state
  const [templateContent, setTemplateContent] = useState(STARTER_TEMPLATE);
  const [templateName, setTemplateName] = useState("My Custom Template");
  const [sampleData, setSampleData] = useState(currentSampleData || DEFAULT_SAMPLE_DATA);
  
  // Preview state
  const [previewOutput, setPreviewOutput] = useState("");
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  
  // Template info (variables/functions)
  const [templateInfo, setTemplateInfo] = useState<TemplateInfo | null>(null);
  const [showVariables, setShowVariables] = useState(true);
  const [showFunctions, setShowFunctions] = useState(true);
  
  // Custom templates
  const [customTemplates, setCustomTemplates] = useState<CustomTemplate[]>([]);
  const [selectedTemplateId, setSelectedTemplateId] = useState<string | null>(null);
  const [selectedBuiltInId, setSelectedBuiltInId] = useState<string | null>(null);
  const [isUnsavedImport, setIsUnsavedImport] = useState(false);
  
  // UI state
  const [activeTab, setActiveTab] = useState<"editor" | "preview" | "sample">("editor");
  const [showSaveDialog, setShowSaveDialog] = useState(false);
  const [copied, setCopied] = useState(false);
  
  // Track if initial preview has run
  const initialPreviewRan = useRef(false);

  // Track if this is the initial open
  const isInitialOpen = useRef(true);

  // Load template info and custom templates on mount
  useEffect(() => {
    if (isOpen) {
      loadTemplateInfo();
      loadCustomTemplates();
      // Reset initial preview flag when modal opens
      initialPreviewRan.current = false;
      // Auto-select default built-in template on first open
      if (isInitialOpen.current) {
        isInitialOpen.current = false;
        // Load the default built-in template
        getTemplateContent("default").then(({ data }) => {
          if (data) {
            setTemplateContent(data.content);
            setTemplateName("default (copy)");
            setSelectedBuiltInId("default");
          }
        });
      }
    }
  }, [isOpen]);

  // Update sample data when prop changes
  useEffect(() => {
    if (currentSampleData) {
      setSampleData(currentSampleData);
    }
  }, [currentSampleData]);

  const loadTemplateInfo = async () => {
    const { data } = await getTemplateInfo();
    if (data) {
      setTemplateInfo(data);
    }
  };

  const loadCustomTemplates = () => {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored) {
        setCustomTemplates(JSON.parse(stored));
      }
    } catch (e) {
      console.error("Failed to load custom templates:", e);
    }
  };

  const saveCustomTemplates = (templates: CustomTemplate[]) => {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(templates));
      setCustomTemplates(templates);
    } catch (e) {
      console.error("Failed to save custom templates:", e);
    }
  };

  // Preview handler
  const runPreview = useCallback(async () => {
    setPreviewLoading(true);
    setPreviewError(null);
    
    const { data, error } = await previewTemplate(templateContent, sampleData);
    
    if (error) {
      setPreviewError(error);
      setPreviewOutput("");
    } else if (data) {
      if (data.error) {
        setPreviewError(data.error);
        setPreviewOutput("");
      } else {
        setPreviewOutput(data.output);
      }
    }
    
    setPreviewLoading(false);
  }, [templateContent, sampleData]);

  // Run preview immediately when modal opens
  useEffect(() => {
    if (isOpen && !initialPreviewRan.current) {
      initialPreviewRan.current = true;
      runPreview();
    }
  }, [isOpen, runPreview]);

  // Auto-preview on template/sample change (debounced)
  useEffect(() => {
    if (!isOpen || !initialPreviewRan.current) return;
    
    const timer = setTimeout(() => {
      runPreview();
    }, 600);
    
    return () => clearTimeout(timer);
  }, [templateContent, sampleData, isOpen, runPreview]);

  // Load built-in template
  const loadBuiltInTemplate = async (id: string) => {
    const { data } = await getTemplateContent(id);
    if (data) {
      setTemplateContent(data.content);
      setTemplateName(`${id} (copy)`);
      setSelectedTemplateId(null);
      setSelectedBuiltInId(id);
      setIsUnsavedImport(false);
    }
  };

  // Load custom template
  const loadCustomTemplate = (template: CustomTemplate) => {
    setTemplateContent(template.content);
    setTemplateName(template.name);
    setSelectedTemplateId(template.id);
    setSelectedBuiltInId(null);
    setIsUnsavedImport(false);
  };

  // Save current template
  const saveTemplate = () => {
    const now = Date.now();
    const existingIndex = customTemplates.findIndex(t => t.id === selectedTemplateId);
    
    if (existingIndex >= 0) {
      // Update existing
      const updated = [...customTemplates];
      updated[existingIndex] = {
        ...updated[existingIndex],
        name: templateName,
        content: templateContent,
        updatedAt: now,
      };
      saveCustomTemplates(updated);
    } else {
      // Create new
      const newTemplate: CustomTemplate = {
        id: `custom-${now}`,
        name: templateName,
        content: templateContent,
        createdAt: now,
        updatedAt: now,
      };
      saveCustomTemplates([...customTemplates, newTemplate]);
      setSelectedTemplateId(newTemplate.id);
      setSelectedBuiltInId(null);
      onSaveTemplate?.(newTemplate);
    }
    
    setIsUnsavedImport(false);
    setShowSaveDialog(false);
  };

  // Delete custom template
  const deleteTemplate = (id: string) => {
    const updated = customTemplates.filter(t => t.id !== id);
    saveCustomTemplates(updated);
    if (selectedTemplateId === id) {
      setSelectedTemplateId(null);
      setTemplateContent(STARTER_TEMPLATE);
      setTemplateName("My Custom Template");
    }
  };

  // Export template as JSON
  const exportTemplate = () => {
    const data = {
      name: templateName,
      content: templateContent,
      exportedAt: new Date().toISOString(),
    };
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${templateName.replace(/\s+/g, "-").toLowerCase()}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  // Import template from JSON
  const importTemplate = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    
    const reader = new FileReader();
    reader.onload = (event) => {
      try {
        const data = JSON.parse(event.target?.result as string);
        if (data.content) {
          setTemplateContent(data.content);
          setTemplateName(data.name || "Imported Template");
          setSelectedTemplateId(null);
          setSelectedBuiltInId(null);
          setIsUnsavedImport(true);
        }
      } catch (err) {
        console.error("Failed to import template:", err);
      }
    };
    reader.readAsText(file);
    e.target.value = "";
  };

  // Copy output to clipboard
  const copyOutput = async () => {
    await navigator.clipboard.writeText(previewOutput);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  // Insert variable/function at cursor
  const insertAtCursor = (text: string, textarea: HTMLTextAreaElement | null) => {
    if (!textarea) return;
    
    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const newContent = 
      templateContent.substring(0, start) + 
      text + 
      templateContent.substring(end);
    
    setTemplateContent(newContent);
    
    // Restore cursor position after insertion
    setTimeout(() => {
      textarea.focus();
      textarea.selectionStart = textarea.selectionEnd = start + text.length;
    }, 0);
  };

  if (!isOpen) return null;

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 bg-black/80 backdrop-blur-sm z-50 flex items-center justify-center p-4"
        onClick={(e) => e.target === e.currentTarget && onClose()}
      >
        <motion.div
          initial={{ scale: 0.95, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.95, opacity: 0 }}
          className="bg-linear-to-br from-surface via-surface to-bg-base rounded-2xl border border-white/10 w-full max-w-7xl h-[90vh] flex flex-col overflow-hidden shadow-2xl"
        >
          {/* Header */}
          <div className="flex items-center justify-between px-6 py-4 border-b border-white/10 bg-black/20">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-xl bg-linear-to-br from-accent-orange/20 to-accent-gold/20 flex items-center justify-center">
                <FileCode className="w-5 h-5 text-accent-orange" />
              </div>
              <div>
                <h2 className="text-lg font-bold text-white">Template Editor</h2>
                <p className="text-xs text-white/50">Create custom MDFlow templates with Go template syntax</p>
              </div>
            </div>
            
            <div className="flex items-center gap-2">
              {/* Import/Export */}
              <label className="p-2 rounded-lg bg-white/5 hover:bg-white/10 text-white/60 hover:text-white transition-all cursor-pointer">
                <Upload className="w-4 h-4" />
                <input type="file" accept=".json" onChange={importTemplate} className="hidden" />
              </label>
              <button
                onClick={exportTemplate}
                className="p-2 rounded-lg bg-white/5 hover:bg-white/10 text-white/60 hover:text-white transition-all cursor-pointer"
                title="Export template"
              >
                <Download className="w-4 h-4" />
              </button>
              
              {/* Save */}
              <button
                onClick={() => setShowSaveDialog(true)}
                className="flex items-center gap-2 px-3 py-2 rounded-lg bg-accent-orange/20 hover:bg-accent-orange/30 text-accent-orange transition-all cursor-pointer"
              >
                <Save className="w-4 h-4" />
                <span className="text-sm font-medium">Save</span>
              </button>
              
              {/* Close */}
              <button
                onClick={onClose}
                className="p-2 rounded-lg hover:bg-white/10 text-white/60 hover:text-white transition-all cursor-pointer"
              >
                <X className="w-5 h-5" />
              </button>
            </div>
          </div>

          {/* Main content */}
          <div className="flex-1 flex overflow-hidden">
            {/* Left sidebar - Templates & Reference */}
            <div className="w-72 border-r border-white/10 flex flex-col bg-black/30">
              {/* Templates section */}
              <div className="p-4 border-b border-white/10">
                <h3 className="text-xs font-bold uppercase tracking-wider text-white/40 mb-3">Templates</h3>
                
                {/* Built-in templates */}
                <div className="space-y-1 mb-4">
                  <p className="text-[10px] text-white/30 uppercase tracking-wider mb-2">Built-in</p>
                  {BUILT_IN_TEMPLATES.map((t) => (
                    <button
                      key={t.id}
                      onClick={() => loadBuiltInTemplate(t.id)}
                      className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-all flex items-center gap-2 cursor-pointer ${
                        selectedBuiltInId === t.id
                          ? "bg-accent-orange/20 text-accent-orange border border-accent-orange/30"
                          : "text-white/70 hover:bg-white/5 hover:text-white border border-transparent"
                      }`}
                    >
                      <FileText className={`w-3.5 h-3.5 ${selectedBuiltInId === t.id ? "text-accent-orange" : "text-white/40"}`} />
                      <span className="truncate">{t.name}</span>
                      {selectedBuiltInId === t.id && (
                        <Check className="w-3 h-3 ml-auto text-accent-orange" />
                      )}
                    </button>
                  ))}
                </div>
                
                {/* Custom templates */}
                {customTemplates.length > 0 && (
                  <div className="space-y-1">
                    <p className="text-[10px] text-white/30 uppercase tracking-wider mb-2">Custom</p>
                    {customTemplates.map((t) => (
                      <div
                        key={t.id}
                        className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-all group ${
                          selectedTemplateId === t.id
                            ? "bg-accent-orange/20 text-accent-orange border border-accent-orange/30"
                            : "text-white/70 hover:bg-white/5 hover:text-white border border-transparent"
                        }`}
                        onClick={() => loadCustomTemplate(t)}
                      >
                        <Sparkles className={`w-3.5 h-3.5 ${selectedTemplateId === t.id ? "text-accent-orange" : "text-accent-orange/60"}`} />
                        <span className="truncate flex-1">{t.name}</span>
                        {selectedTemplateId === t.id && (
                          <Check className="w-3 h-3 text-accent-orange" />
                        )}
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            deleteTemplate(t.id);
                          }}
                          className="opacity-0 group-hover:opacity-100 p-1 hover:bg-red-500/20 rounded text-red-400 transition-all"
                        >
                          <Trash2 className="w-3 h-3" />
                        </button>
                      </div>
                    ))}
                  </div>
                )}
                
                {/* Unsaved import indicator */}
                {isUnsavedImport && (
                  <div className="mt-3 p-3 rounded-lg bg-accent-orange/10 border border-accent-orange/30">
                    <div className="flex items-center gap-2 text-accent-orange text-xs font-medium">
                      <Upload className="w-3.5 h-3.5" />
                      Imported (unsaved)
                    </div>
                    <p className="text-[10px] text-white/50 mt-1">{templateName}</p>
                    <button
                      onClick={() => setShowSaveDialog(true)}
                      className="mt-2 w-full py-1.5 rounded bg-accent-orange/20 hover:bg-accent-orange/30 text-accent-orange text-xs font-medium transition-all"
                    >
                      Save to Custom
                    </button>
                  </div>
                )}
                
                {/* New template button */}
                <button
                  onClick={() => {
                    setTemplateContent(STARTER_TEMPLATE);
                    setTemplateName("My Custom Template");
                    setSelectedTemplateId(null);
                    setSelectedBuiltInId(null);
                    setIsUnsavedImport(false);
                  }}
                  className="w-full mt-3 flex items-center justify-center gap-2 px-3 py-2 rounded-lg border border-dashed border-white/20 text-white/50 hover:border-accent-orange/50 hover:text-accent-orange transition-all"
                >
                  <Plus className="w-4 h-4" />
                  <span className="text-sm">New Template</span>
                </button>
              </div>
              
              {/* Reference section */}
              <div className="flex-1 overflow-auto p-4 custom-scrollbar">
                <h3 className="text-xs font-bold uppercase tracking-wider text-white/40 mb-3 flex items-center gap-2">
                  <BookOpen className="w-3.5 h-3.5" />
                  Reference
                </h3>
                
                {templateInfo && (
                  <div className="space-y-4">
                    {/* Variables */}
                    <div>
                      <button
                        onClick={() => setShowVariables(!showVariables)}
                        className="flex items-center gap-2 text-xs text-white/60 hover:text-white mb-2 w-full"
                      >
                        {showVariables ? <ChevronDown className="w-3 h-3" /> : <ChevronRight className="w-3 h-3" />}
                        <Variable className="w-3 h-3" />
                        <span className="font-medium">Variables</span>
                      </button>
                      {showVariables && (
                        <div className="space-y-1 ml-5">
                          {templateInfo.variables.map((v) => (
                            <VariableItem key={v.name} variable={v} onInsert={insertAtCursor} />
                          ))}
                        </div>
                      )}
                    </div>
                    
                    {/* Functions */}
                    <div>
                      <button
                        onClick={() => setShowFunctions(!showFunctions)}
                        className="flex items-center gap-2 text-xs text-white/60 hover:text-white mb-2 w-full"
                      >
                        {showFunctions ? <ChevronDown className="w-3 h-3" /> : <ChevronRight className="w-3 h-3" />}
                        <Braces className="w-3 h-3" />
                        <span className="font-medium">Functions</span>
                      </button>
                      {showFunctions && (
                        <div className="space-y-1 ml-5">
                          {templateInfo.functions.map((f) => (
                            <FunctionItem key={f.name} func={f} onInsert={insertAtCursor} />
                          ))}
                        </div>
                      )}
                    </div>
                  </div>
                )}
                
                {/* Quick tips */}
                <div className="mt-6 p-3 rounded-lg bg-accent-orange/5 border border-accent-orange/20">
                  <div className="flex items-center gap-2 text-accent-orange/80 text-xs font-medium mb-2">
                    <Lightbulb className="w-3.5 h-3.5" />
                    Quick Tips
                  </div>
                  <ul className="text-[10px] text-white/50 space-y-1">
                    <li>• Use <code className="text-accent-orange">{"{{range}}"}</code> to loop</li>
                    <li>• Use <code className="text-accent-orange">{"{{if}}"}</code> for conditions</li>
                    <li>• Use <code className="text-accent-orange">{"{{- -}}"}</code> to trim whitespace</li>
                    <li>• Access row fields with <code className="text-accent-orange">.Field</code></li>
                  </ul>
                </div>
              </div>
            </div>
            
            {/* Main editor area */}
            <div className="flex-1 flex flex-col">
              {/* Tabs */}
              <div className="flex items-center gap-1 px-4 py-2 border-b border-white/10 bg-black/20">
                <TabButton
                  active={activeTab === "editor"}
                  onClick={() => setActiveTab("editor")}
                  icon={<Code2 className="w-3.5 h-3.5" />}
                  label="Template"
                />
                <TabButton
                  active={activeTab === "preview"}
                  onClick={() => setActiveTab("preview")}
                  icon={<Eye className="w-3.5 h-3.5" />}
                  label="Preview"
                  badge={previewError ? "!" : undefined}
                />
                <TabButton
                  active={activeTab === "sample"}
                  onClick={() => setActiveTab("sample")}
                  icon={<FileText className="w-3.5 h-3.5" />}
                  label="Sample Data"
                />
                
                <div className="flex-1" />
                
                {/* Run preview button */}
                <button
                  onClick={() => {
                    runPreview();
                    setActiveTab("preview");
                  }}
                  disabled={previewLoading}
                  className="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-accent-green/20 hover:bg-accent-green/30 text-accent-green text-sm font-medium transition-all disabled:opacity-50 cursor-pointer"
                >
                  <Play className="w-3.5 h-3.5" />
                  Run Preview
                </button>
              </div>
              
              {/* Editor/Preview content */}
              <div className="flex-1 overflow-hidden">
                {activeTab === "editor" && (
                  <div className="h-full flex flex-col">
                    {/* Template name input */}
                    <div className="px-4 py-2 border-b border-white/5 bg-black/10">
                      <input
                        type="text"
                        value={templateName}
                        onChange={(e) => setTemplateName(e.target.value)}
                        className="bg-transparent text-white/80 text-sm font-medium w-full focus:outline-none placeholder:text-white/30"
                        placeholder="Template name..."
                      />
                    </div>
                    
                    {/* Template editor */}
                    <textarea
                      id="template-editor"
                      value={templateContent}
                      onChange={(e) => setTemplateContent(e.target.value)}
                      className="flex-1 w-full bg-black/40 text-white/90 font-mono text-sm p-4 resize-none focus:outline-none placeholder:text-white/30 leading-relaxed"
                      placeholder="Enter your Go template here..."
                      spellCheck={false}
                    />
                  </div>
                )}
                
                {activeTab === "preview" && (
                  <div className="h-full flex flex-col relative">
                    {/* Loading overlay */}
                    {previewLoading && (
                      <div className="absolute inset-0 bg-black/50 backdrop-blur-sm z-10 flex items-center justify-center">
                        <div className="flex flex-col items-center gap-3">
                          <Loader2 className="w-8 h-8 text-accent-orange animate-spin" />
                          <span className="text-sm text-white/60">Generating preview...</span>
                        </div>
                      </div>
                    )}
                    {previewError ? (
                      <div className="flex-1 flex items-center justify-center p-8">
                        <div className="text-center max-w-md">
                          <div className="w-12 h-12 rounded-full bg-accent-red/20 flex items-center justify-center mx-auto mb-4">
                            <AlertCircle className="w-6 h-6 text-accent-red" />
                          </div>
                          <h3 className="text-white font-medium mb-2">Template Error</h3>
                          <p className="text-accent-red/80 text-sm font-mono bg-accent-red/10 p-3 rounded-lg border border-accent-red/20">
                            {previewError}
                          </p>
                        </div>
                      </div>
                    ) : (
                      <>
                        <div className="px-4 py-2 border-b border-white/5 flex items-center justify-between bg-black/10">
                          <span className="text-xs text-white/40">Preview Output</span>
                          <button
                            onClick={copyOutput}
                            className="flex items-center gap-1.5 px-2 py-1 rounded text-xs text-white/50 hover:text-white hover:bg-white/10 transition-all"
                          >
                            {copied ? <Check className="w-3 h-3 text-accent-green" /> : <Copy className="w-3 h-3" />}
                            {copied ? "Copied!" : "Copy"}
                          </button>
                        </div>
                        <pre className="flex-1 overflow-auto p-4 text-white/80 font-mono text-sm whitespace-pre-wrap bg-black/40 custom-scrollbar">
                          {previewOutput || "No output yet. Edit template to see preview."}
                        </pre>
                      </>
                    )}
                  </div>
                )}
                
                {activeTab === "sample" && (
                  <div className="h-full flex flex-col">
                    <div className="px-4 py-2 border-b border-white/5 flex items-center justify-between bg-black/10">
                      <span className="text-xs text-white/40">Sample Data (TSV format)</span>
                      <button
                        onClick={() => setSampleData(DEFAULT_SAMPLE_DATA)}
                        className="text-xs text-white/50 hover:text-accent-orange transition-all"
                      >
                        Reset to default
                      </button>
                    </div>
                    <textarea
                      value={sampleData}
                      onChange={(e) => setSampleData(e.target.value)}
                      className="flex-1 w-full bg-black/40 text-white/90 font-mono text-sm p-4 resize-none focus:outline-none placeholder:text-white/30 leading-relaxed"
                      placeholder="Paste your sample TSV data here..."
                      spellCheck={false}
                    />
                  </div>
                )}
              </div>
            </div>
          </div>
        </motion.div>
        
        {/* Save dialog */}
        <AnimatePresence>
          {showSaveDialog && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 bg-black/60 z-60 flex items-center justify-center"
              onClick={() => setShowSaveDialog(false)}
            >
              <motion.div
                initial={{ scale: 0.95 }}
                animate={{ scale: 1 }}
                exit={{ scale: 0.95 }}
                onClick={(e) => e.stopPropagation()}
                className="bg-surface rounded-xl border border-white/10 p-6 w-full max-w-md shadow-2xl"
              >
                <h3 className="text-lg font-bold text-white mb-4">Save Template</h3>
                <input
                  type="text"
                  value={templateName}
                  onChange={(e) => setTemplateName(e.target.value)}
                  className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-3 text-white focus:outline-none focus:border-accent-orange/50 mb-4"
                  placeholder="Template name..."
                />
                <div className="flex gap-3 justify-end">
                  <button
                    onClick={() => setShowSaveDialog(false)}
                    className="px-4 py-2 rounded-lg text-white/60 hover:text-white hover:bg-white/10 transition-all"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={saveTemplate}
                    className="px-4 py-2 rounded-lg bg-accent-orange hover:bg-accent-orange/90 text-white font-medium transition-all"
                  >
                    Save Template
                  </button>
                </div>
              </motion.div>
            </motion.div>
          )}
        </AnimatePresence>
      </motion.div>
    </AnimatePresence>
  );
}

// Tab button component
function TabButton({
  active,
  onClick,
  icon,
  label,
  badge,
}: {
  active: boolean;
  onClick: () => void;
  icon: React.ReactNode;
  label: string;
  badge?: string;
}) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm font-medium transition-all ${
        active
          ? "bg-accent-orange/20 text-accent-orange border border-accent-orange/30"
          : "text-white/50 hover:text-white hover:bg-white/5 border border-transparent"
      }`}
    >
      {icon}
      {label}
      {badge && (
        <span className="w-4 h-4 rounded-full bg-accent-red text-white text-[10px] flex items-center justify-center">
          {badge}
        </span>
      )}
    </button>
  );
}

// Variable item component
function VariableItem({
  variable,
  onInsert,
}: {
  variable: TemplateVariable;
  onInsert: (text: string, textarea: HTMLTextAreaElement | null) => void;
}) {
  return (
    <button
      onClick={() => {
        const textarea = document.getElementById("template-editor") as HTMLTextAreaElement;
        onInsert(`{{${variable.name}}}`, textarea);
      }}
      className="w-full text-left p-2 rounded hover:bg-white/5 transition-all group"
    >
      <div className="flex items-center gap-2">
        <code className="text-[10px] text-accent-orange bg-accent-orange/10 px-1.5 py-0.5 rounded">
          {variable.name}
        </code>
        <span className="text-[9px] text-white/30">{variable.type}</span>
      </div>
      <p className="text-[10px] text-white/40 mt-1 group-hover:text-white/60">
        {variable.description}
      </p>
    </button>
  );
}

// Function item component
function FunctionItem({
  func,
  onInsert,
}: {
  func: TemplateFunction;
  onInsert: (text: string, textarea: HTMLTextAreaElement | null) => void;
}) {
  return (
    <button
      onClick={() => {
        const textarea = document.getElementById("template-editor") as HTMLTextAreaElement;
        onInsert(`{{${func.name} }}`, textarea);
      }}
      className="w-full text-left p-2 rounded hover:bg-white/5 transition-all group"
    >
      <div className="flex items-center gap-2">
        <code className="text-[10px] text-accent-green bg-accent-green/10 px-1.5 py-0.5 rounded">
          {func.name}
        </code>
      </div>
      <p className="text-[9px] text-white/30 font-mono mt-1">{func.signature}</p>
      <p className="text-[10px] text-white/40 mt-1 group-hover:text-white/60">
        {func.description}
      </p>
    </button>
  );
}
