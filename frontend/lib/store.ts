import { create } from 'zustand';

interface AuthState {
  token: string | null;
  email: string | null;
  userId: string | null;
  isLoggedIn: boolean;
  setAuth: (token: string, email: string, userId?: string) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  token: typeof window !== 'undefined' ? localStorage.getItem('token') : null,
  email: typeof window !== 'undefined' ? localStorage.getItem('email') : null,
  userId: typeof window !== 'undefined' ? localStorage.getItem('userId') : null,
  isLoggedIn: typeof window !== 'undefined' ? !!localStorage.getItem('token') : false,

  setAuth: (token: string, email: string, userId?: string) => {
    localStorage.setItem('token', token);
    localStorage.setItem('email', email);
    if (userId) localStorage.setItem('userId', userId);
    set({ token, email, userId: userId || null, isLoggedIn: true });
  },

  clearAuth: () => {
    localStorage.removeItem('token');
    localStorage.removeItem('email');
    localStorage.removeItem('userId');
    set({ token: null, email: null, userId: null, isLoggedIn: false });
  },
}));

export interface TableData {
  headers: string[];
  rows: Record<string, string>[];
  sheet_name: string;
}

interface ConversionState {
  tableData: TableData | null;
  selectedTemplateId: string | null;
  markdown: string | null;
  setTableData: (data: TableData) => void;
  setSelectedTemplate: (templateId: string) => void;
  setMarkdown: (content: string) => void;
  clear: () => void;
}

export const useConversionStore = create<ConversionState>((set) => ({
  tableData: null,
  selectedTemplateId: null,
  markdown: null,

  setTableData: (data: TableData) => set({ tableData: data }),
  setSelectedTemplate: (templateId: string) => set({ selectedTemplateId: templateId }),
  setMarkdown: (content: string) => set({ markdown: content }),
  clear: () => set({ tableData: null, selectedTemplateId: null, markdown: null }),
}));
