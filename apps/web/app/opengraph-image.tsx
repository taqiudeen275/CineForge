import { ImageResponse } from "next/og";

export const alt = "CineForge — story-first AI filmmaking";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function OpenGraphImage() {
  return new ImageResponse(
    <div
      style={{
        width: "100%",
        height: "100%",
        display: "flex",
        flexDirection: "column",
        justifyContent: "space-between",
        padding: "72px 78px",
        color: "white",
        background:
          "radial-gradient(circle at 78% 18%, rgba(123,92,246,.58), transparent 32%), radial-gradient(circle at 90% 90%, rgba(35,194,144,.25), transparent 28%), #07080d",
      }}
    >
      <div style={{ display: "flex", alignItems: "center", gap: 18, fontSize: 28 }}>
        <div
          style={{
            width: 44,
            height: 44,
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            borderRadius: 12,
            background: "#7b5cf6",
            fontWeight: 800,
          }}
        >
          C
        </div>
        <span style={{ fontWeight: 700, letterSpacing: "-0.02em" }}>CineForge</span>
      </div>
      <div style={{ display: "flex", flexDirection: "column", gap: 24 }}>
        <div style={{ fontSize: 70, lineHeight: 1.02, letterSpacing: "-0.055em", maxWidth: 960 }}>
          Develop the world. Direct the shots. Finish the story.
        </div>
        <div style={{ fontSize: 25, color: "#bbb6c5" }}>
          One private, traceable production truth from first idea to final cut.
        </div>
      </div>
      <div style={{ display: "flex", fontSize: 16, letterSpacing: ".16em", color: "#8f899b" }}>
        EARLY ACCESS · STORY-FIRST AI FILMMAKING
      </div>
    </div>,
    size,
  );
}
