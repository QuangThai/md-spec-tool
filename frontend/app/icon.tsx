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
        width="100%"
        height="100%"
        viewBox="0 0 48 48"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          <linearGradient
            id="icon-flow-gradient"
            x1="0%"
            y1="0%"
            x2="100%"
            y2="0%"
          >
            <stop offset="0%" stopColor="#F59E0B" />
            <stop offset="100%" stopColor="#F97316" />
          </linearGradient>
        </defs>
        <g>
          <path
            d="M10 38V18C10 14.6863 12.6863 12 16 12C19.3137 12 22 14.6863 22 18V28C22 29.1046 22.8954 30 24 30C25.1046 30 26 29.1046 26 28V18C26 14.6863 28.6863 12 32 12C35.3137 12 38 14.6863 38 18V38"
            stroke="url(#icon-flow-gradient)"
            strokeWidth="5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
          <circle cx="24" cy="39" r="3" fill="#F97316" />
        </g>
      </svg>
    </div>,
    {
      ...size,
    },
  );
}
