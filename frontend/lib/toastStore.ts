import { create } from 'zustand';

export interface Toast {
  id: string;
  type: 'success' | 'error' | 'info' | 'warning';
  message: string;
  description?: string;
  duration?: number;
}

export interface ToastStore {
  toasts: Toast[];
  addToast: (toast: Omit<Toast, 'id'>) => void;
  removeToast: (id: string) => void;
}

export const useToastStore = create<ToastStore>((set) => ({
  toasts: [],
  addToast: (toast) =>
    set((state) => ({
      toasts: [
        ...state.toasts,
        {
          ...toast,
          id: `toast-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`,
          duration: toast.duration ?? 3000,
        },
      ],
    })),
  removeToast: (id) =>
    set((state) => ({
      toasts: state.toasts.filter((t) => t.id !== id),
    })),
}));

export const toast = {
  success: (message: string, description?: string) =>
    useToastStore.getState().addToast({ type: 'success', message, description }),
  error: (message: string, description?: string) =>
    useToastStore.getState().addToast({ type: 'error', message, description }),
  info: (message: string, description?: string) =>
    useToastStore.getState().addToast({ type: 'info', message, description }),
  warning: (message: string, description?: string) =>
    useToastStore.getState().addToast({ type: 'warning', message, description }),
};
