import React from "react";
import {
  AbsoluteFill,
  Easing,
  interpolate,
  spring,
  useCurrentFrame,
  useVideoConfig,
} from "remotion";

const accent = "#f27b2f";
const glow = "rgba(242, 123, 47, 0.35)";
const ink = "#f5f5f5";
const soft = "rgba(255,255,255,0.06)";

const useFloat = (speed: number, amount: number) => {
  const frame = useCurrentFrame();
  return Math.sin(frame / speed) * amount;
};

const easeInOut = (frame: number, start: number, end: number) =>
  interpolate(frame, [start, end], [0, 1], {
    easing: Easing.inOut(Easing.cubic),
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

const ambientDots = [
  { x: 12, y: 18, size: 6, speed: 0.6, opacity: 0.4 },
  { x: 22, y: 62, size: 4, speed: 0.8, opacity: 0.3 },
  { x: 78, y: 22, size: 5, speed: 0.7, opacity: 0.35 },
  { x: 86, y: 58, size: 7, speed: 0.5, opacity: 0.4 },
  { x: 64, y: 78, size: 3, speed: 1.0, opacity: 0.25 },
  { x: 38, y: 32, size: 4, speed: 0.9, opacity: 0.25 },
];

const AmbientParticles: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  return (
    <AbsoluteFill style={{ pointerEvents: "none" }}>
      {ambientDots.map((dot, index) => {
        const float = Math.sin((frame / fps) * dot.speed + index) * 8;
        const pulse = 0.6 + 0.4 * Math.sin((frame / fps) * 0.7 + index * 1.4);
        return (
          <div
            key={index}
            style={{
              position: "absolute",
              left: `${dot.x}%`,
              top: `${dot.y}%`,
              width: dot.size,
              height: dot.size,
              borderRadius: 999,
              background: "rgba(242,123,47,0.4)",
              boxShadow: `0 0 12px rgba(242,123,47,0.5)`,
              opacity: dot.opacity * pulse,
              transform: `translateY(${float}px)`,
            }}
          />
        );
      })}
    </AbsoluteFill>
  );
};

const SceneFrame: React.FC<
  React.PropsWithChildren<{ start: number; end: number; style?: React.CSSProperties }>
> = ({ start, end, style, children }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const enter = spring({
    frame: frame - start,
    fps,
    config: { damping: 200 },
  });
  const exit = spring({
    frame: frame - (end - fps * 0.4),
    fps,
    config: { damping: 200 },
  });
  const opacity = interpolate(
    frame,
    [start, start + fps * 0.35, end - fps * 0.35, end],
    [0, 1, 1, 0],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
  );
  const translateY = interpolate(
    frame,
    [start, start + fps * 0.6],
    [10, 0],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
  );
  const scale = 0.98 + enter * 0.02 - exit * 0.02;
  return (
    <div
      style={{
        position: "absolute",
        inset: 0,
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        opacity: opacity * (1 - exit),
        transform: `translateY(${translateY}px) scale(${scale})`,
        ...style,
      }}
    >
      {children}
    </div>
  );
};

const LabelPill: React.FC<{ text: string }> = ({ text }) => (
  <div
    style={{
      display: "inline-flex",
      alignItems: "center",
      gap: 10,
      padding: "8px 14px",
      borderRadius: 999,
      border: "1px solid rgba(255,255,255,0.12)",
      background: "rgba(255,255,255,0.04)",
      fontSize: 10,
      letterSpacing: "0.32em",
      textTransform: "uppercase",
      fontWeight: 700,
      width: "fit-content",
    }}
  >
    <span
      style={{
        width: 8,
        height: 8,
        borderRadius: 999,
        background: accent,
        boxShadow: `0 0 12px ${glow}`,
      }}
    />
    {text}
  </div>
);

const TitleBlock: React.FC<{ title: string; subtitle: string }> = ({
  title,
  subtitle,
}) => (
  <div style={{ display: "grid", gap: 12 }}>
    <div
      style={{
        fontSize: 38,
        fontWeight: 800,
        letterSpacing: "-0.03em",
        color: ink,
        lineHeight: 1.05,
        textShadow: "0 18px 45px rgba(0,0,0,0.45)",
      }}
    >
      {title}
    </div>
    <div style={{ color: "rgba(255,255,255,0.6)", fontSize: 18, maxWidth: 320 }}>
      {subtitle}
    </div>
  </div>
);

const WindowShell: React.FC<React.PropsWithChildren> = ({ children }) => {
  return (
    <div
      style={{
        width: "100%",
        height: "100%",
        borderRadius: 28,
        border: "1px solid rgba(255,255,255,0.12)",
        background: "rgba(10,10,10,0.75)",
        boxShadow: "0 30px 80px rgba(0,0,0,0.55)",
        overflow: "hidden",
        position: "relative",
      }}
    >
      <div
        style={{
          position: "absolute",
          inset: 0,
          background:
            "linear-gradient(120deg, rgba(255,255,255,0.08), transparent 45%, rgba(242,123,47,0.15))",
          opacity: 0.3,
          pointerEvents: "none",
        }}
      />
      <div
        style={{
          height: 44,
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          padding: "0 20px",
          borderBottom: "1px solid rgba(255,255,255,0.08)",
          background: "linear-gradient(90deg, rgba(255,255,255,0.03), rgba(255,255,255,0))",
        }}
      >
        <div style={{ display: "flex", gap: 8 }}>
          <div style={{ width: 10, height: 10, borderRadius: 999, background: "#5c2b2b" }} />
          <div style={{ width: 10, height: 10, borderRadius: 999, background: "#6a4a1b" }} />
          <div style={{ width: 10, height: 10, borderRadius: 999, background: "#5a3a1d" }} />
        </div>
        <div style={{ display: "flex", gap: 16, opacity: 0.6 }}>
          <div style={{ width: 90, height: 6, borderRadius: 999, background: "rgba(255,255,255,0.08)" }} />
          <div style={{ width: 50, height: 6, borderRadius: 999, background: "rgba(242,123,47,0.2)" }} />
        </div>
      </div>
      <div style={{ padding: 32, height: "calc(100% - 44px)" }}>{children}</div>
    </div>
  );
};

const RightPanelBase: React.FC<React.PropsWithChildren> = ({ children }) => (
  <div
    style={{
      width: "100%",
      borderRadius: 22,
      border: "1px solid rgba(255,255,255,0.08)",
      background: "rgba(14,14,14,0.7)",
      padding: 24,
      height: "100%",
      boxShadow: "0 22px 55px rgba(0,0,0,0.55)",
      position: "relative",
      overflow: "hidden",
    }}
  >
    <div
      style={{
        position: "absolute",
        inset: -40,
        background:
          "radial-gradient(circle at 20% 10%, rgba(242,123,47,0.22), transparent 55%), radial-gradient(circle at 90% 60%, rgba(255,255,255,0.05), transparent 50%)",
        opacity: 0.4,
        pointerEvents: "none",
      }}
    />
    <div
      style={{
        position: "absolute",
        inset: 0,
        borderRadius: 22,
        border: "1px solid rgba(255,255,255,0.04)",
        boxShadow: "inset 0 1px 0 rgba(255,255,255,0.04)",
        pointerEvents: "none",
      }}
    />
    {children}
  </div>
);

const IngestPanel: React.FC<{ start: number; end: number }> = ({ start, end }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const pulse = spring({
    frame: frame - start,
    fps,
    config: { damping: 18, stiffness: 120 },
  });
  const burst = spring({
    frame: frame - start - 8,
    fps,
    config: { damping: 20, stiffness: 140 },
  });
  const progress = interpolate(
    frame,
    [start + 10, end - 10],
    [0.15, 1],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
  );

  return (
    <RightPanelBase>
      <div style={{ display: "grid", gap: 18 }}>
        <div style={{ height: 12, width: 120, borderRadius: 999, background: soft }} />
        <div style={{ height: 10, width: 220, borderRadius: 999, background: "rgba(255,255,255,0.1)" }} />
        <div
          style={{
            height: 70,
            borderRadius: 18,
            border: "1px solid rgba(255,255,255,0.08)",
            background: "rgba(255,255,255,0.04)",
            position: "relative",
            overflow: "hidden",
          }}
        >
          <div
            style={{
              position: "absolute",
              inset: 0,
              background:
                "linear-gradient(90deg, transparent, rgba(242,123,47,0.45), transparent)",
              opacity: 0.35,
              transform: `translateX(${interpolate(
                frame,
                [start, end],
                [-60, 120],
                { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
              )}%)`,
            }}
          />
          <div
            style={{
              position: "absolute",
              inset: 0,
              width: `${progress * 100}%`,
              background: `linear-gradient(90deg, rgba(242,123,47,0.2), rgba(242,123,47,0.6))`,
              opacity: 0.5,
            }}
          />
          <div
            style={{
              position: "absolute",
              inset: 0,
              background:
                "linear-gradient(180deg, rgba(255,255,255,0.04), transparent 60%)",
              opacity: 0.6,
            }}
          />
          <div
            style={{
              position: "absolute",
              inset: "18px 24px",
              display: "grid",
              gap: 8,
            }}
          >
            <span
              style={{
                position: "absolute",
                top: 0,
                left: 0,
                fontSize: 8,
                fontWeight: 600,
                letterSpacing: "0.15em",
                textTransform: "uppercase",
                color: "rgba(255,255,255,0.35)",
              }}
            >
              Paste stream
            </span>
            {[0, 1, 2].map((i) => (
              <div
                key={i}
                style={{
                  height: 8,
                  borderRadius: 999,
                  background: "rgba(255,255,255,0.1)",
                }}
              />
            ))}
          </div>
          <div
            style={{
              position: "absolute",
              right: 20,
              top: 14,
              width: 36,
              height: 36,
              borderRadius: 12,
              background: accent,
              boxShadow: `0 0 24px ${glow}`,
              transform: `scale(${0.9 + pulse * 0.1})`,
            }}
          />
          <div
            style={{
              position: "absolute",
              right: 8,
              bottom: -12,
              width: 90,
              height: 90,
              borderRadius: "50%",
              border: "1px solid rgba(242,123,47,0.35)",
              opacity: 0.4,
              transform: `scale(${0.6 + burst * 0.6})`,
              filter: "blur(0.5px)",
            }}
          />
        </div>
        <div style={{ display: "flex", gap: 12 }}>
          {["TSV", "CSV", "XLSX"].map((label, i) => (
            <div
              key={i}
              style={{
                flex: 1,
                height: 46,
                borderRadius: 16,
                border: "1px solid rgba(255,255,255,0.08)",
                background: "rgba(255,255,255,0.03)",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              <span
                style={{
                  fontSize: 10,
                  fontWeight: 600,
                  letterSpacing: "0.12em",
                  textTransform: "uppercase",
                  color: "rgba(255,255,255,0.5)",
                }}
              >
                {label}
              </span>
            </div>
          ))}
        </div>
      </div>
    </RightPanelBase>
  );
};

const MapPanel: React.FC<{ start: number; end: number }> = ({ start, end }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const nodeLabels = ["Col A", "Col B", "Node", "Schema", "Valid"];
  const sweep = interpolate(frame, [start + 10, end - 10], [-20, 100], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  return (
    <RightPanelBase>
      <div style={{ display: "grid", gap: 18 }}>
        <div
          style={{
            height: 14,
            width: 140,
            borderRadius: 999,
            background: "rgba(255,255,255,0.06)",
            display: "flex",
            alignItems: "center",
            paddingLeft: 12,
          }}
        >
          <span
            style={{
              fontSize: 8,
              fontWeight: 600,
              letterSpacing: "0.18em",
              textTransform: "uppercase",
              color: "rgba(255,255,255,0.35)",
            }}
          >
            Columns
          </span>
        </div>
        <div style={{ display: "grid", gridTemplateColumns: "repeat(5, 1fr)", gap: 14 }}>
          {nodeLabels.map((label, index) => {
            const scale = spring({
              frame: frame - start - index * 6,
              fps,
              config: { damping: 18, stiffness: 140 },
            });
            return (
              <div
                key={label}
                style={{
                  height: 60,
                  borderRadius: 16,
                  border: "1px solid rgba(255,255,255,0.08)",
                  background: "rgba(255,255,255,0.04)",
                  transform: `scale(${0.9 + scale * 0.1})`,
                  position: "relative",
                  overflow: "hidden",
                  display: "flex",
                  alignItems: "center",
                  justifyContent: "center",
                }}
              >
                <div
                  style={{
                    position: "absolute",
                    inset: 8,
                    borderRadius: 12,
                    background: `linear-gradient(120deg, rgba(255,255,255,0.02), rgba(242,123,47,0.2))`,
                  }}
                />
                <div
                  style={{
                    position: "absolute",
                    inset: 0,
                    background:
                      "linear-gradient(90deg, transparent, rgba(242,123,47,0.25), transparent)",
                    transform: `translateX(${sweep}%)`,
                    opacity: 0.4,
                  }}
                />
                <span
                  style={{
                    position: "relative",
                    zIndex: 1,
                    fontSize: 9,
                    fontWeight: 600,
                    letterSpacing: "0.08em",
                    textTransform: "uppercase",
                    color: "rgba(255,255,255,0.55)",
                  }}
                >
                  {label}
                </span>
              </div>
            );
          })}
        </div>
        <div
          style={{
            height: 2,
            borderRadius: 999,
            background: "rgba(255,255,255,0.1)",
            position: "relative",
          }}
        >
          <div
            style={{
              position: "absolute",
              inset: 0,
              width: `${interpolate(
                frame,
                [start + 10, end - 10],
                [0, 1],
                { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
              ) * 100}%`,
              background: accent,
              boxShadow: `0 0 18px ${glow}`,
            }}
          />
        </div>

        {/* Mapping preview - fills empty space */}
        <div
          style={{
            marginTop: 20,
            padding: "14px 16px",
            borderRadius: 14,
            border: "1px solid rgba(255,255,255,0.06)",
            background: "rgba(255,255,255,0.02)",
            display: "grid",
            gap: 12,
          }}
        >
          <div
            style={{
              fontSize: 8,
              fontWeight: 700,
              letterSpacing: "0.2em",
              textTransform: "uppercase",
              color: "rgba(255,255,255,0.35)",
            }}
          >
            Mapping preview
          </div>
          <div style={{ display: "grid", gap: 8 }}>
            {[
              { from: "id", to: "node_id" },
              { from: "name", to: "title" },
              { from: "ts", to: "created_at" },
            ].map((row, i) => (
              <div
                key={i}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 10,
                  fontSize: 10,
                  color: "rgba(255,255,255,0.5)",
                  fontFamily: "monospace",
                }}
              >
                <span style={{ color: "rgba(255,255,255,0.4)", minWidth: 48 }}>{row.from}</span>
                <span style={{ color: "rgba(242,123,47,0.6)", fontSize: 9 }}>→</span>
                <span style={{ color: "rgba(255,255,255,0.55)" }}>{row.to}</span>
              </div>
            ))}
          </div>
          <div
            style={{
              height: 1,
              background: "rgba(255,255,255,0.06)",
              borderRadius: 999,
            }}
          />
          <div
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
              fontSize: 9,
              color: accent,
              fontWeight: 600,
              letterSpacing: "0.1em",
            }}
          >
            <span>5 nodes</span>
            <span style={{ color: "rgba(255,255,255,0.3)" }}>·</span>
            <span>3 dependencies</span>
          </div>
        </div>
      </div>
    </RightPanelBase>
  );
};

const OutputPanel: React.FC<{ start: number; end: number }> = ({ start, end }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const rows = [0, 1, 2, 3, 4];
  const tilt = useFloat(55, 2);
  const glowPulse = interpolate(
    frame,
    [start, start + 30, end - 20, end],
    [0.2, 0.6, 0.6, 0.2],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
  );
  const orbit = useFloat(70, 6);
  const hudReveal = spring({
    frame: frame - start - fps * 0.2,
    fps,
    config: { damping: 200 },
  });
  return (
    <RightPanelBase>
      <div style={{ display: "grid", gap: 16, position: "relative" }}>
        <div
          style={{
            position: "absolute",
            inset: 0,
            background:
              "radial-gradient(circle at 30% 80%, rgba(255,255,255,0.06), transparent 45%)",
            opacity: 0.7,
          }}
        />
        <div
          style={{
            position: "absolute",
            inset: 0,
            background:
              "radial-gradient(circle at 80% 0%, rgba(242,123,47,0.12), transparent 45%)",
            opacity: 0.8,
          }}
        />
        <div
          style={{
            position: "absolute",
            top: 12,
            right: 12,
            display: "grid",
            gap: 10,
            opacity: hudReveal,
            transform: `translateY(${(1 - hudReveal) * 10}px)`,
          }}
        >
          <div
            style={{
              padding: "8px 12px",
              borderRadius: 14,
              border: "1px solid rgba(255,255,255,0.08)",
              background: "rgba(10,10,10,0.75)",
              boxShadow: "0 12px 30px rgba(0,0,0,0.4)",
              display: "grid",
              gap: 6,
            }}
          >
            <div style={{ fontSize: 9, letterSpacing: "0.22em", textTransform: "uppercase", color: "rgba(255,255,255,0.45)" }}>
              Output Health
            </div>
            <div style={{ display: "flex", gap: 6 }}>
              {[0, 1, 2].map((i) => (
                <div
                  key={i}
                  style={{
                    width: 10,
                    height: 10,
                    borderRadius: 999,
                    background: i === 2 ? accent : "rgba(255,255,255,0.15)",
                    boxShadow: i === 2 ? `0 0 12px ${glow}` : "none",
                  }}
                />
              ))}
            </div>
          </div>
          <div
            style={{
              padding: "10px 12px",
              borderRadius: 14,
              border: "1px solid rgba(242,123,47,0.25)",
              background: "rgba(0,0,0,0.6)",
              display: "grid",
              gap: 6,
            }}
          >
            <div style={{ fontSize: 9, letterSpacing: "0.22em", textTransform: "uppercase", color: accent }}>
              Throughput
            </div>
            <div style={{ fontSize: 16, fontWeight: 700, color: ink }}>2.4k/s</div>
            <div style={{ height: 4, borderRadius: 999, background: "rgba(255,255,255,0.1)", overflow: "hidden" }}>
              <div
                style={{
                  height: "100%",
                  width: `${60 + hudReveal * 30}%`,
                  background: `linear-gradient(90deg, ${accent}, rgba(255,255,255,0.6))`,
                }}
              />
            </div>
          </div>
        </div>
        <div
          style={{
            position: "absolute",
            left: 18,
            bottom: 24,
            display: "grid",
            gap: 10,
            opacity: 0.85,
          }}
        >
          <div
            style={{
              fontSize: 9,
              letterSpacing: "0.22em",
              textTransform: "uppercase",
              color: "rgba(255,255,255,0.4)",
            }}
          >
            Compliance
          </div>
          <div style={{ display: "grid", gap: 6 }}>
            {[0, 1, 2].map((i) => (
              <div
                key={i}
                style={{
                  width: 90,
                  height: 6,
                  borderRadius: 999,
                  background: "rgba(255,255,255,0.08)",
                  position: "relative",
                  overflow: "hidden",
                }}
              >
                <div
                  style={{
                    position: "absolute",
                    inset: 0,
                    width: `${70 + i * 8}%`,
                    background: `linear-gradient(90deg, rgba(255,255,255,0.2), ${accent})`,
                    opacity: 0.4,
                  }}
                />
              </div>
            ))}
          </div>
        </div>
        {rows.map((row) => (
          <div
            key={row}
            style={{
              height: 10,
              width: `${70 + row * 6}%`,
              borderRadius: 999,
              background: "rgba(255,255,255,0.08)",
              position: "relative",
              overflow: "hidden",
            }}
          >
            <div
              style={{
                position: "absolute",
                inset: 0,
                transform: `translateX(${interpolate(
                  frame,
                  [start + row * 6, end - 10],
                  [-50, 0],
                  { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
                )}%)`,
                background: `linear-gradient(90deg, transparent, ${accent}, transparent)`,
                opacity: 0.5,
              }}
            />
          </div>
        ))}
        <div
          style={{
            position: "absolute",
            left: 16,
            bottom: -100,
            padding: "16px 22px",
            borderRadius: 18,
            background: "rgba(0,0,0,0.65)",
            border: "1px solid rgba(242,123,47,0.35)",
            boxShadow: `0 18px 40px rgba(0,0,0,0.45), 0 0 ${24 + glowPulse * 20}px rgba(242,123,47,0.25)`,
            transform: `rotate(${tilt}deg) translateY(${orbit}px)`,
          }}
        >
          <div
            style={{
              fontSize: 10,
              letterSpacing: "0.25em",
              textTransform: "uppercase",
              color: accent,
              fontWeight: 700,
              marginBottom: 6,
            }}
          >
            Efficiency Output
          </div>
          <div style={{ fontSize: 30, fontWeight: 800, color: ink }}>+84%</div>
        </div>
      </div>
    </RightPanelBase>
  );
};

export const LandingHeroVideo: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps, durationInFrames } = useVideoConfig();
  const driftX = useFloat(90, 6);
  const driftY = useFloat(110, 4);
  const tilt = useFloat(140, 1.2);
  const shimmerX = interpolate(frame % (8 * fps), [0, 8 * fps], [-40, 140]);
  const gridShift = (frame * 0.6) % 48;
  const windowReveal = spring({
    frame,
    fps,
    config: { damping: 200 },
  });
  const spotlight = easeInOut(frame, 0, durationInFrames * 0.5);
  const sweep = interpolate(frame % (6 * fps), [0, 6 * fps], [-60, 120]);
  return (
    <AbsoluteFill
      style={{
        fontFamily: "var(--font-inter), system-ui, sans-serif",
        color: ink,
        background: `radial-gradient(circle at ${22 + driftX * 0.6}% ${18 + driftY * 0.4}%, rgba(242,123,47,0.2), transparent 45%), radial-gradient(circle at ${78 - driftX * 0.4}% ${8 + driftY * 0.2}%, rgba(195,125,13,0.22), transparent 50%), #0a0a0a`,
      }}
    >
      <AmbientParticles />
      <AbsoluteFill
        style={{
          backgroundImage:
            "linear-gradient(rgba(255,255,255,0.04) 1px, transparent 1px), linear-gradient(90deg, rgba(255,255,255,0.04) 1px, transparent 1px)",
          backgroundSize: "48px 48px",
          backgroundPosition: `0 ${gridShift}px`,
          opacity: 0.22,
        }}
      />
      <AbsoluteFill
        style={{
          background:
            "radial-gradient(circle at 50% 60%, transparent 35%, rgba(0,0,0,0.7) 100%)",
          mixBlendMode: "multiply",
        }}
      />
      <AbsoluteFill
        style={{
          background:
            "linear-gradient(120deg, transparent 0%, rgba(242,123,47,0.18) 45%, transparent 70%)",
          opacity: 0.5,
          transform: `translateX(${shimmerX}%)`,
          mixBlendMode: "screen",
        }}
      />
      <AbsoluteFill
        style={{
          background:
            "linear-gradient(90deg, transparent, rgba(255,255,255,0.08), transparent)",
          opacity: 0.35,
          transform: `translateX(${sweep}%)`,
          mixBlendMode: "screen",
        }}
      />
      <AbsoluteFill style={{ position: "relative", padding: "26px 38px" }}>
        <div
          style={{
            position: "absolute",
            top: 26,
            left: 38,
            right: 38,
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            fontSize: 11,
            textTransform: "uppercase",
            letterSpacing: "0.22em",
            color: "rgba(255,255,255,0.5)",
          }}
        >
          <div>MDFlow Runtime</div>
          <div style={{ display: "flex", gap: 16 }}>
            <span>CPU 2%</span>
            <span>RAM 42MB</span>
          </div>
        </div>

        <div
          style={{
            position: "absolute",
            left: "50%",
            top: "50%",
            transform: `translate(-50%, -50%) perspective(1200px) rotateX(1.2deg) rotateY(-2deg) scale(${0.96 + windowReveal * 0.04}) rotate(${tilt}deg) translate(${driftX}px, ${driftY}px)`,
            width: "92%",
            height: "70%",
          }}
        >
          <div style={{ position: "relative", width: "100%", height: "100%" }}>
            <div
              style={{
                position: "absolute",
                inset: -30,
                background:
                  "radial-gradient(circle at 30% 20%, rgba(242,123,47,0.2), transparent 55%), radial-gradient(circle at 70% 80%, rgba(255,255,255,0.06), transparent 55%)",
                opacity: 0.7 * spotlight,
                filter: "blur(12px)",
              }}
            />
            <div
              style={{
                position: "absolute",
                inset: 0,
                borderRadius: 28,
                border: "1px solid rgba(242,123,47,0.2)",
                boxShadow: `0 0 ${28 + spotlight * 40}px rgba(242,123,47,0.15)`,
                opacity: 0.6 * spotlight,
              }}
            />
            <div
              style={{
                position: "absolute",
                inset: 0,
                borderRadius: 28,
                background:
                  "linear-gradient(110deg, rgba(255,255,255,0.08), transparent 35%, rgba(255,255,255,0.02))",
                opacity: 0.6,
                mixBlendMode: "screen",
              }}
            />
            <WindowShell>
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "0.9fr 1.1fr",
                  gap: 28,
                  height: "100%",
                }}
              >
                <div style={{ position: "relative", height: "100%" }}>
                  <SceneFrame start={0} end={90} style={{ justifyContent: "flex-start" }}>
                    <div style={{ display: "grid", gap: 20 }}>
                      <LabelPill text="INGEST" />
                      <TitleBlock
                        title="Stream Ingestion"
                        subtitle="Parse TSV, CSV, and XLSX at wire speed."
                      />
                    </div>
                  </SceneFrame>
                  <SceneFrame start={90} end={180} style={{ justifyContent: "flex-start" }}>
                    <div style={{ display: "grid", gap: 20 }}>
                      <LabelPill text="MAP" />
                      <TitleBlock
                        title="Node Mapping"
                        subtitle="Identify dependencies and verify structure."
                      />
                    </div>
                  </SceneFrame>
                  <SceneFrame start={180} end={270} style={{ justifyContent: "flex-start" }}>
                    <div style={{ display: "grid", gap: 20 }}>
                      <LabelPill text="OUTPUT" />
                      <TitleBlock
                        title="Markdown Specs"
                        subtitle="Generate compliant documentation in minutes."
                      />
                    </div>
                  </SceneFrame>
                </div>

                <div style={{ position: "relative", height: "100%" }}>
                  <SceneFrame start={0} end={90}>
                    <IngestPanel start={0} end={90} />
                  </SceneFrame>
                  <SceneFrame start={90} end={180}>
                    <MapPanel start={90} end={180} />
                  </SceneFrame>
                  <SceneFrame start={180} end={270}>
                    <OutputPanel start={180} end={270} />
                  </SceneFrame>
                </div>
              </div>
            </WindowShell>
          </div>
        </div>

        <div
          style={{
            position: "absolute",
            bottom: 22,
            left: 38,
            right: 38,
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            fontSize: 11,
            color: "rgba(255,255,255,0.4)",
          }}
        >
          <div>BUILD 24.3KB</div>
          <div style={{ color: accent }}>Output Spec.MD</div>
        </div>
      </AbsoluteFill>
    </AbsoluteFill>
  );
};
