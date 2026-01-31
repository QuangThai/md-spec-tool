import { ImageResponse } from "next/og";

// Route segment config
export const runtime = "edge";

// Image metadata
export const size = {
  width: 32,
  height: 32,
};
export const contentType = "image/png";

// Image generation - Clean geometric MDFlow icon
export default function Icon() {
  return new ImageResponse(
    <div
      style={{
        background: "linear-gradient(145deg, #0a0a0a, #111111)",
        width: "100%",
        height: "100%",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        borderRadius: "22%",
        border: "1px solid rgba(249, 115, 22, 0.15)",
        boxShadow: "0 2px 8px rgba(0,0,0,0.4), inset 0 1px 0 rgba(255,255,255,0.03)",
      }}
    >
      <svg
        width="24"
        height="24"
        viewBox="0 0 32 32"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          <linearGradient
            id="icon-gradient"
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
            id="icon-gradient-subtle"
            x1="0%"
            y1="0%"
            x2="100%"
            y2="0%"
          >
            <stop offset="0%" stopColor="#F97316" stopOpacity="0.85" />
            <stop offset="100%" stopColor="#FB923C" stopOpacity="0.6" />
          </linearGradient>
        </defs>
        {/* Geometric M shape */}
        <path
          d="M4 24V10C4 9.2 4.6 8.6 5.4 8.6H7L13 17L19 8.6H21C21.8 8.6 22.4 9.2 22.4 10V24"
          stroke="url(#icon-gradient)"
          strokeWidth="3.2"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        {/* Flow accent chevron */}
        <path
          d="M25 9L28.5 16L25 23"
          stroke="url(#icon-gradient-subtle)"
          strokeWidth="2.8"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </svg>
    </div>,
    {
      ...size,
    },
  );
}
