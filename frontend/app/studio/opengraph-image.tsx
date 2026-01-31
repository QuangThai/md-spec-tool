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
        alignItems: "center",
        justifyContent: "center",
        background: "linear-gradient(135deg, #050505 0%, #141414 100%)",
        color: "#ffffff",
        fontFamily: "sans-serif",
        position: "relative",
      }}
    >
      <div
        style={{
          position: "absolute",
          top: "-20%",
          right: "-5%",
          width: "700px",
          height: "700px",
          background:
            "radial-gradient(circle, rgba(249,115,22,0.14) 0%, transparent 60%)",
          filter: "blur(60px)",
        }}
      />
      <div
        style={{
          position: "absolute",
          bottom: "-30%",
          left: "-10%",
          width: "700px",
          height: "700px",
          background:
            "radial-gradient(circle, rgba(245,158,11,0.1) 0%, transparent 60%)",
          filter: "blur(60px)",
        }}
      />

      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          gap: 18,
          padding: "60px 80px",
          border: "1px solid rgba(255,255,255,0.08)",
          borderRadius: "32px",
          background: "rgba(255,255,255,0.03)",
          boxShadow: "0 24px 48px rgba(0,0,0,0.5)",
          zIndex: 10,
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 16,
          }}
        >
          <svg width="64" height="64" viewBox="0 0 32 32" fill="none">
            <defs>
              <linearGradient id="studio-gradient" x1="0%" y1="0%" x2="100%" y2="100%">
                <stop offset="0%" stopColor="#F97316" />
                <stop offset="50%" stopColor="#F59E0B" />
                <stop offset="100%" stopColor="#EA580C" />
              </linearGradient>
              <linearGradient id="studio-gradient-subtle" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#F97316" stopOpacity="0.85" />
                <stop offset="100%" stopColor="#FB923C" stopOpacity="0.6" />
              </linearGradient>
            </defs>
            <path
              d="M4 24V10C4 9.2 4.6 8.6 5.4 8.6H7L13 17L19 8.6H21C21.8 8.6 22.4 9.2 22.4 10V24"
              stroke="url(#studio-gradient)"
              strokeWidth="3.2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <path
              d="M25 9L28.5 16L25 23"
              stroke="url(#studio-gradient-subtle)"
              strokeWidth="2.8"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          <div
            style={{
              display: "flex",
              fontSize: 28,
              fontWeight: 800,
              textTransform: "uppercase",
              letterSpacing: "0.18em",
              color: "rgba(255,255,255,0.7)",
            }}
          >
            MDFlow
          </div>
        </div>

        <div
          style={{
            fontSize: 64,
            fontWeight: 900,
            letterSpacing: "-0.02em",
            color: "white",
            textAlign: "center",
          }}
        >
          Studio Workspace
        </div>
        <div
          style={{
            fontSize: 28,
            color: "#a1a1aa",
            maxWidth: "680px",
            textAlign: "center",
            lineHeight: 1.4,
          }}
        >
          Paste or upload data, map columns, and generate specs in minutes.
        </div>

        <div
          style={{
            display: "flex",
            gap: 12,
            marginTop: 10,
          }}
        >
          {[
            "Parsing",
            "Templates",
            "Validation",
          ].map((label) => (
            <div
              key={label}
              style={{
                padding: "8px 16px",
                borderRadius: "999px",
                background: "rgba(249,115,22,0.12)",
                border: "1px solid rgba(249,115,22,0.2)",
                color: "#F97316",
                fontSize: 16,
                fontWeight: 700,
                textTransform: "uppercase",
                letterSpacing: "0.12em",
              }}
            >
              {label}
            </div>
          ))}
        </div>
      </div>
    </div>,
    {
      width: size.width,
      height: size.height,
    },
  );
}
