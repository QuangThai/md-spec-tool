"use client";

import { Player } from "@remotion/player";
import { LandingHeroVideo } from "./LandingHeroVideo";

export function LandingHeroPlayer() {
  return (
    <Player
      component={LandingHeroVideo}
      durationInFrames={270}
      fps={30}
      compositionWidth={1200}
      compositionHeight={514}
      autoPlay
      loop
      controls={false}
      acknowledgeRemotionLicense
      style={{ width: "100%", height: "100%" }}
    />
  );
}
