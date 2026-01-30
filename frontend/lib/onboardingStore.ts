import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface OnboardingStep {
  id: string;
  target: string; // CSS selector or data-tour attribute
  title: string;
  description: string;
  position: 'top' | 'bottom' | 'left' | 'right';
}

interface OnboardingState {
  hasSeenTour: boolean;
  isActive: boolean;
  currentStep: number;
  
  // Actions
  startTour: () => void;
  nextStep: () => void;
  prevStep: () => void;
  skipTour: () => void;
  completeTour: () => void;
  resetTour: () => void;
}

export const ONBOARDING_STEPS: OnboardingStep[] = [
  {
    id: 'welcome',
    target: '[data-tour="welcome"]',
    title: 'Welcome to MDFlow Studio',
    description: 'Transform spreadsheet data into structured Markdown specifications. Let me show you how it works.',
    position: 'bottom',
  },
  {
    id: 'input-mode',
    target: '[data-tour="input-mode"]',
    title: 'Choose Input Method',
    description: 'Paste text directly, upload Excel files (.xlsx), or upload TSV files. You can also paste Google Sheets URLs.',
    position: 'bottom',
  },
  {
    id: 'paste-area',
    target: '[data-tour="paste-area"]',
    title: 'Paste Your Data',
    description: 'Paste tab-separated or comma-separated data here. The tool will auto-detect the format and show you a preview.',
    position: 'right',
  },
  {
    id: 'preview-table',
    target: '[data-tour="preview-table"]',
    title: 'Preview & Map Columns',
    description: 'See how your data is parsed. You can manually change column mappings using the dropdowns if auto-detection isn\'t perfect.',
    position: 'top',
  },
  {
    id: 'template-selector',
    target: '[data-tour="template-selector"]',
    title: 'Select a Template',
    description: 'Choose from different output formats: Default, Feature Spec, Test Plan, API Endpoint, or Spec Table.',
    position: 'top',
  },
  {
    id: 'run-button',
    target: '[data-tour="run-button"]',
    title: 'Generate Output',
    description: 'Click "Run" or press ⌘+Enter to convert your data. The output will appear on the right panel.',
    position: 'left',
  },
  {
    id: 'output-panel',
    target: '[data-tour="output-panel"]',
    title: 'Your MDFlow Output',
    description: 'Copy, export (MD/JSON), or save for comparison. Your conversion history is automatically saved.',
    position: 'left',
  },
];

export const useOnboardingStore = create<OnboardingState>()(
  persist(
    (set, get) => ({
      hasSeenTour: false,
      isActive: false,
      currentStep: 0,

      startTour: () => set({ isActive: true, currentStep: 0 }),
      
      nextStep: () => {
        const { currentStep } = get();
        if (currentStep < ONBOARDING_STEPS.length - 1) {
          set({ currentStep: currentStep + 1 });
        } else {
          // Last step - complete the tour
          get().completeTour();
        }
      },
      
      prevStep: () => {
        const { currentStep } = get();
        if (currentStep > 0) {
          set({ currentStep: currentStep - 1 });
        }
      },
      
      skipTour: () => set({ isActive: false, hasSeenTour: true, currentStep: 0 }),
      
      completeTour: () => set({ isActive: false, hasSeenTour: true, currentStep: 0 }),
      
      resetTour: () => set({ hasSeenTour: false, isActive: false, currentStep: 0 }),
    }),
    {
      name: 'mdflow-onboarding',
    }
  )
);
