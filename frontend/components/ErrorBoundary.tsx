"use client";

import React, { ReactNode } from "react";

interface Props {
  children: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export default class ErrorBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error("Error caught by boundary:", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen bg-bg-base flex items-center justify-center px-4">
          <div className="surface max-w-md w-full text-center">
            <div className="mb-6 flex justify-center">
              <div className="h-16 w-16 rounded-2xl bg-accent-red/10 flex items-center justify-center border border-accent-red/20 shadow-premium">
                <span className="text-2xl">⚠️</span>
              </div>
            </div>
            <h1 className="text-2xl font-bold text-white mb-2 uppercase tracking-tighter">
              Engine Halted
            </h1>
            <p className="text-muted text-sm mb-8 leading-relaxed uppercase tracking-widest font-medium">
              {this.state.error?.message ||
                "A critical kernel panic occurred during operation."}
            </p>
            <button
              onClick={() => {
                this.setState({ hasError: false, error: null });
                window.location.href = "/";
              }}
              className="btn-primary w-full"
            >
              Reboot Engine
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
