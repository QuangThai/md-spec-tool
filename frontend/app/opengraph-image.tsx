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
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        background: "linear-gradient(135deg, #020202 0%, #0F0F0F 100%)",
        color: "#ffffff",
        fontFamily: "sans-serif",
        position: "relative",
      }}
    >
      {/* Glow Effects */}
      <div
        style={{
          position: "absolute",
          top: "-20%",
          left: "20%",
          width: "800px",
          height: "800px",
          background:
            "radial-gradient(circle, rgba(249,115,22,0.15) 0%, transparent 60%)",
          filter: "blur(40px)",
        }}
      />
      <div
        style={{
          position: "absolute",
          bottom: "-20%",
          right: "10%",
          width: "600px",
          height: "600px",
          background:
            "radial-gradient(circle, rgba(245,158,11,0.1) 0%, transparent 60%)",
          filter: "blur(60px)",
        }}
      />

      {/* Content Container */}
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          gap: 24,
          zIndex: 10,
          padding: "60px 80px",
          border: "1px solid rgba(255,255,255,0.1)",
          borderRadius: "32px",
          background: "rgba(255,255,255,0.03)",
          boxShadow: "0 24px 48px rgba(0,0,0,0.5)",
        }}
      >
        {/* Logo & Brand */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 24,
            marginBottom: 12,
          }}
        >
          {/* Clean geometric M icon with flow accent */}
          <svg width="80" height="80" viewBox="0 0 32 32" fill="none">
            <defs>
              <linearGradient
                id="og-gradient"
                x1="0%"
                y1="0%"
                x2="100%"
                y2="100%"
              >
                <stop offset="0%" stopColor="#F97316" />
                <stop offset="50%" stopColor="#F59E0B" />
                <stop offset="100%" stopColor="#EA580C" />
              </linearGradient>
              <linearGradient
                id="og-gradient-subtle"
                x1="0%"
                y1="0%"
                x2="100%"
                y2="0%"
              >
                <stop offset="0%" stopColor="#F97316" stopOpacity="0.85" />
                <stop offset="100%" stopColor="#FB923C" stopOpacity="0.6" />
              </linearGradient>
            </defs>
            {/* Geometric M */}
            <path
              d="M4 24V10C4 9.2 4.6 8.6 5.4 8.6H7L13 17L19 8.6H21C21.8 8.6 22.4 9.2 22.4 10V24"
              stroke="url(#og-gradient)"
              strokeWidth="3.2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            {/* Flow chevron */}
            <path
              d="M25 9L28.5 16L25 23"
              stroke="url(#og-gradient-subtle)"
              strokeWidth="2.8"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
          <div
            style={{
              display: "flex",
              fontSize: 72,
              fontWeight: 800,
              letterSpacing: "-0.03em",
            }}
          >
            <span style={{ color: "white" }}>MD</span>
            <span
              style={{
                background: "linear-gradient(135deg, #F97316, #F59E0B)",
                backgroundClip: "text",
                color: "transparent",
              }}
            >
              Flow
            </span>
          </div>
        </div>

        <div
          style={{
            width: "120px",
            height: "4px",
            background: "linear-gradient(90deg, #F59E0B, #F97316)",
            borderRadius: "4px",
            opacity: 0.8,
          }}
        />

        <div
          style={{
            fontSize: 32,
            fontWeight: 500,
            color: "#a1a1aa",
            letterSpacing: "-0.01em",
            marginTop: 8,
          }}
        >
          Technical Specification Automation
        </div>

        {/* Badge */}
        <div
          style={{
            marginTop: 24,
            padding: "8px 20px",
            background: "rgba(249, 115, 22, 0.1)",
            border: "1px solid rgba(249, 115, 22, 0.2)",
            borderRadius: "100px",
            fontSize: 18,
            fontWeight: 700,
            color: "#F97316",
            textTransform: "uppercase",
            letterSpacing: "0.1em",
          }}
        >
          Studio v1.2
        </div>
      </div>
    </div>,
    {
      width: size.width,
      height: size.height,
    },
  );
}
