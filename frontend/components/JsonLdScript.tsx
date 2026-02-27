"use client";

import { useEffect, useRef } from "react";

const ORGANIZATION_JSON_LD = {
  "@context": "https://schema.org",
  "@type": "Organization",
  name: "MDFlow Studio",
  url: "https://md-spec-tool.vercel.app",
};

export function JsonLdScript() {
  const injected = useRef(false);

  useEffect(() => {
    if (injected.current) return;
    injected.current = true;
    const script = document.createElement("script");
    script.type = "application/ld+json";
    script.textContent = JSON.stringify(ORGANIZATION_JSON_LD);
    document.head.appendChild(script);
    return () => {
      if (script.parentNode) {
        document.head.removeChild(script);
      }
      injected.current = false;
    };
  }, []);

  return null;
}
