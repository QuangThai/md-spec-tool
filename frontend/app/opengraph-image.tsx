import { ImageResponse } from "next/og";

export const runtime = "edge";

export const size = {
  width: 1200,
  height: 630,
};

export const contentType = "image/png";

export default async function OpenGraphImage() {
  const fontData = await fetch(
    "https://fonts.googleapis.com/css2?family=Inter:wght@700;800&display=swap",
  ).then((res) => res.arrayBuffer());

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          background: "linear-gradient(135deg, #050505 0%, #0b0b0b 100%)",
          color: "#ffffff",
          fontFamily: "Inter",
          position: "relative",
        }}
      >
        <div
          style={{
            position: "absolute",
            inset: 0,
            background:
              "radial-gradient(circle at 20% 20%, rgba(249,115,22,0.25), transparent 50%), radial-gradient(circle at 80% 30%, rgba(245,158,11,0.2), transparent 45%)",
          }}
        />
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            gap: 16,
            textAlign: "center",
            padding: "0 80px",
            zIndex: 1,
          }}
        >
          <div style={{ fontSize: 64, fontWeight: 800, letterSpacing: "-0.02em" }}>
            MDFlow Studio
          </div>
          <div style={{ fontSize: 28, color: "#d4d4d8" }}>
            Technical Specification Automation
          </div>
        </div>
      </div>
    ),
    {
      width: size.width,
      height: size.height,
      fonts: [
        {
          name: "Inter",
          data: fontData,
          weight: 800,
          style: "normal",
        },
      ],
    },
  );
}
