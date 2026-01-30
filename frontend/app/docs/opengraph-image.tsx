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
            gap: 12,
            marginBottom: 24,
          }}
        >
          <svg width="32" height="32" viewBox="0 0 48 48" fill="none">
            <defs>
              <linearGradient id="tiny-flow" x1="0%" y1="0%" x2="100%" y2="0%">
                <stop offset="0%" stopColor="#F59E0B" />
                <stop offset="100%" stopColor="#F97316" />
              </linearGradient>
            </defs>
            <path
              d="M10 38V18C10 14.6863 12.6863 12 16 12C19.3137 12 22 14.6863 22 18V28C22 29.1046 22.8954 30 24 30C25.1046 30 26 29.1046 26 28V18C26 14.6863 28.6863 12 32 12C35.3137 12 38 14.6863 38 18V38"
              stroke="url(#tiny-flow)"
              strokeWidth="5"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            <circle cx="24" cy="39" r="3" fill="#F97316" />
          </svg>
          <div
            style={{
              fontSize: 24,
              fontWeight: 700,
              color: "rgba(255,255,255,0.6)",
              textTransform: "uppercase",
              letterSpacing: "0.1em",
            }}
          >
            MDFlow
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
