"use client";

import * as React from "react";

type CollapsibleContextValue = {
  open: boolean;
  setOpen: (next: boolean) => void;
};

const CollapsibleContext = React.createContext<CollapsibleContextValue | null>(null);

type CollapsibleProps = {
  children: React.ReactNode;
  defaultOpen?: boolean;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
};

function Collapsible({
  children,
  defaultOpen = false,
  open: controlledOpen,
  onOpenChange,
}: CollapsibleProps) {
  const [internalOpen, setInternalOpen] = React.useState(defaultOpen);
  const open = typeof controlledOpen === "boolean" ? controlledOpen : internalOpen;

  const setOpen = React.useCallback(
    (next: boolean) => {
      if (typeof controlledOpen !== "boolean") {
        setInternalOpen(next);
      }
      onOpenChange?.(next);
    },
    [controlledOpen, onOpenChange]
  );

  return (
    <CollapsibleContext.Provider value={{ open, setOpen }}>
      <div>{children}</div>
    </CollapsibleContext.Provider>
  );
}

type TriggerProps = React.ButtonHTMLAttributes<HTMLButtonElement>;

function CollapsibleTrigger({ children, onClick, type, ...props }: TriggerProps) {
  const context = React.useContext(CollapsibleContext);
  if (!context) {
    throw new Error("CollapsibleTrigger must be used within Collapsible");
  }

  return (
    <button
      type={type ?? "button"}
      onClick={(event) => {
        context.setOpen(!context.open);
        onClick?.(event);
      }}
      aria-expanded={context.open}
      {...props}
    >
      {children}
    </button>
  );
}

type ContentProps = React.HTMLAttributes<HTMLDivElement>;

function CollapsibleContent({ children, ...props }: ContentProps) {
  const context = React.useContext(CollapsibleContext);
  if (!context) {
    throw new Error("CollapsibleContent must be used within Collapsible");
  }
  if (!context.open) {
    return null;
  }
  return <div {...props}>{children}</div>;
}

export { Collapsible, CollapsibleTrigger, CollapsibleContent };
