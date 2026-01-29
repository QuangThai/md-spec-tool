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
        fontSize: 24,
        background: "transparent",
        width: "100%",
        height: "100%",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      <svg
        width="32"
        height="32"
        viewBox="0 0 32 32"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          <linearGradient id="icon-grad" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" stopColor="#F7CE68" />
            <stop offset="100%" stopColor="#f27b2f" />
          </linearGradient>
        </defs>
        <path
          d="M2 24L2 12C2 12 2 8 6 8C10 8 10 12 10 12L10 24"
          stroke="url(#icon-grad)"
          strokeWidth="3"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        <path
          d="M10 24L10 16C10 16 10 12 14 12C18 12 18 16 18 16L18 24"
          stroke="url(#icon-grad)"
          strokeWidth="3"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeOpacity="0.8"
        />
        <path
          d="M18 24L18 20C18 20 18 16 22 16C26 16 26 20 26 20L26 24"
          stroke="url(#icon-grad)"
          strokeWidth="3"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeOpacity="0.6"
        />
        <circle cx="30" cy="10" r="2" fill="#F7CE68" />
      </svg>
    </div>,
    {
      ...size,
    },
  );
}
