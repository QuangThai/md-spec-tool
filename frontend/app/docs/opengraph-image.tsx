import { ImageResponse } from "next/og";

export const runtime = "edge";

export const size = {
  width: 1200,
  height: 630,
};

export const contentType = "image/png";

export default function OpenGraphImage() {
  return new ImageResponse(
    <div
      style={{
        width: "100%",
        height: "100%",
        display: "flex",
        alignItems: "stretch",
        justifyContent: "space-between",
        background: "linear-gradient(135deg, #020202 0%, #111111 100%)",
        color: "#ffffff",
        fontFamily: "sans-serif",
        position: "relative",
        overflow: "hidden",
      }}
    >
      {/* Abstract background shapes */}
      <div
        style={{
          position: "absolute",
          top: "-10%",
          right: "-10%",
          width: "700px",
          height: "700px",
          background:
            "radial-gradient(circle, rgba(251, 191, 36, 0.08) 0%, transparent 70%)",
          filter: "blur(60px)",
        }}
      />

      <div
        style={{
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          padding: "0 80px",
          flex: 1,
          zIndex: 10,
        }}
      >
        {/* Small Brand Label */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 14,
            marginBottom: 24,
          }}
        >
          <svg width="36" height="36" viewBox="0 0 32 32" fill="none">
            <defs>
              <linearGradient id="docs-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stopColor="#F97316" />
                <stop offset="50%" stopColor="#F59E0B" />
                <stop offset="100%" stopColor="#EA580C" />
              </linearGradient>
              <linearGradient id="docs-gradient-subtle" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#F97316" stopOpacity="0.85" />
                <stop offset="100%" stopColor="#FB923C" stopOpacity="0.6" />
              </linearGradient>
            </defs>
            <path
              d="M4 24V10C4 9.2 4.6 8.6 5.4 8.6H7L13 17L19 8.6H21C21.8 8.6 22.4 9.2 22.4 10V24"
              stroke="url(#docs-gradient)"
              strokeWidth="3.2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <path
              d="M25 9L28.5 16L25 23"
              stroke="url(#docs-gradient-subtle)"
              strokeWidth="2.8"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          <div
            style={{
              display: "flex",
              fontSize: 24,
              fontWeight: 700,
              textTransform: "uppercase",
              letterSpacing: "0.1em",
            }}
          >
            <span style={{ color: "rgba(255,255,255,0.6)" }}>MD</span>
            <span style={{ color: "#F97316" }}>Flow</span>
          </div>
        </div>

        <div
          style={{
            fontSize: 72,
            fontWeight: 900,
            lineHeight: 1.1,
            letterSpacing: "-0.02em",
            color: "white",
            marginBottom: 16,
          }}
        >
          Documentation <br />
          <span style={{ color: "#F97316" }}>& Guides</span>
        </div>
        <div
          style={{
            fontSize: 32,
            color: "#a1a1aa",
            maxWidth: "600px",
            lineHeight: 1.4,
          }}
        >
          Mastering the spec engine: architecture, parsing logic, and template
          injection.
        </div>
      </div>

      {/* Right decoration */}
      <div
        style={{
          flex: 0.6,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          position: "relative",
          borderLeft: "1px solid rgba(255,255,255,0.05)",
          background: "rgba(255,255,255,0.01)",
        }}
      >
        <div
          style={{
            width: "80%",
            height: "60%",
            borderRadius: "24px",
            border: "1px solid rgba(255,255,255,0.1)",
            background:
              "linear-gradient(180deg, rgba(255,255,255,0.05), rgba(0,0,0,0))",
            display: "flex",
            flexDirection: "column",
            padding: "32px",
            boxShadow: "0 24px 48px rgba(0,0,0,0.5)",
          }}
        >
          <div style={{ display: "flex", gap: 8, marginBottom: 20 }}>
            <div
              style={{
                width: 12,
                height: 12,
                borderRadius: "50%",
                background: "#F97316",
                opacity: 0.6,
              }}
            ></div>
            <div
              style={{
                width: 12,
                height: 12,
                borderRadius: "50%",
                background: "#F59E0B",
                opacity: 0.6,
              }}
            ></div>
          </div>
          <div
            style={{
              height: 10,
              width: "60%",
              background: "rgba(255,255,255,0.1)",
              borderRadius: 4,
              marginBottom: 12,
            }}
          ></div>
          <div
            style={{
              height: 10,
              width: "90%",
              background: "rgba(255,255,255,0.1)",
              borderRadius: 4,
              marginBottom: 12,
            }}
          ></div>
          <div
            style={{
              height: 10,
              width: "75%",
              background: "rgba(255,255,255,0.1)",
              borderRadius: 4,
              marginBottom: 12,
            }}
          ></div>
          <div
            style={{
              height: 10,
              width: "80%",
              background: "rgba(255,255,255,0.1)",
              borderRadius: 4,
              marginBottom: 12,
            }}
          ></div>
        </div>
      </div>
    </div>,
    {
      width: size.width,
      height: size.height,
    },
  );
}
