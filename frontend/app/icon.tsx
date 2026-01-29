import { ImageResponse } from "next/og";

// Route segment config
export const runtime = "edge";

// Image metadata
export const size = {
  width: 32,
  height: 32,
};
export const contentType = "image/png";

// Image generation
export default function Icon() {
  return new ImageResponse(
    <div
      style={{
        background: "linear-gradient(145deg, #050505, #111)",
        width: "100%",
        height: "100%",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        borderRadius: "20%",
        border: "1px solid rgba(255,255,255,0.1)",
      }}
    >
      <svg
        width="24"
        height="24"
        viewBox="0 0 44 40"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          <linearGradient
            id="icon-prism-light"
            x1="0%"
            y1="0%"
            x2="100%"
            y2="100%"
          >
            <stop offset="0%" stopColor="#FFF7ED" />
            <stop offset="100%" stopColor="#FCD34D" />
          </linearGradient>
          <linearGradient id="icon-prism-mid" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="#F59E0B" />
            <stop offset="100%" stopColor="#D97706" />
          </linearGradient>
          <linearGradient
            id="icon-prism-dark"
            x1="100%"
            y1="0%"
            x2="0%"
            y2="100%"
          >
            <stop offset="0%" stopColor="#B45309" />
            <stop offset="100%" stopColor="#78350F" />
          </linearGradient>
        </defs>
        <g transform="translate(0, 2)">
          {/* Left Leg */}
          <path d="M6 32V12L16 6L16 26L6 32Z" fill="url(#icon-prism-dark)" />
          <path
            d="M6 12L16 6L22 10L12 16L6 12Z"
            fill="url(#icon-prism-light)"
          />
          <path d="M16 26L22 22V10L16 6V26Z" fill="url(#icon-prism-mid)" />

          {/* Connector */}
          <path d="M22 10L28 14L22 22Z" fill="#78350F" />

          {/* Right Leg */}
          <path d="M38 32V12L28 6L28 26L38 32Z" fill="url(#icon-prism-dark)" />
          <path
            d="M38 12L28 6L22 10L32 16L38 12Z"
            fill="url(#icon-prism-light)"
          />
          <path d="M28 26L22 22V10L28 6V26Z" fill="url(#icon-prism-mid)" />
        </g>
      </svg>
    </div>,
    {
      ...size,
    },
  );
}
