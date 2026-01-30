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
            gap: 20,
            marginBottom: 12,
          }}
        >
          {/* Fluid M Icon */}
          <svg width="84" height="84" viewBox="0 0 48 48" fill="none">
            <defs>
              <linearGradient
                id="og-flow-gradient"
                x1="0%"
                y1="0%"
                x2="100%"
                y2="0%"
              >
                <stop offset="0%" stopColor="#F59E0B" />
                <stop offset="100%" stopColor="#F97316" />
              </linearGradient>
            </defs>
            <path
              d="M10 38V18C10 14.6863 12.6863 12 16 12C19.3137 12 22 14.6863 22 18V28C22 29.1046 22.8954 30 24 30C25.1046 30 26 29.1046 26 28V18C26 14.6863 28.6863 12 32 12C35.3137 12 38 14.6863 38 18V38"
              stroke="url(#og-flow-gradient)"
              strokeWidth="5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <circle cx="24" cy="39" r="3" fill="#F97316" />
          </svg>
          <div
            style={{
              fontSize: 72,
              fontWeight: 800,
              letterSpacing: "-0.03em",
              color: "white",
            }}
          >
            MDFlow
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
