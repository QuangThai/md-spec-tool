"use client";

import { transcribeAudio } from "@/lib/audioApi";
import {
  AudioTranscriptionResponse,
  TranscriptSplit,
} from "@/lib/audioTypes";
import {
  Download,
  Minus,
  Pause,
  Play,
  Plus,
  Scissors,
  Upload,
  X,
} from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";

// Style for scrollbar (scoped to component)
const scrollbarStyles = `
  ::-webkit-scrollbar {
    height: 6px;
  }
  ::-webkit-scrollbar-track {
    background: rgba(0, 0, 0, 0.1);
    border-radius: 3px;
  }
  ::-webkit-scrollbar-thumb {
    background: rgba(242, 123, 47, 0.3);
    border-radius: 3px;
  }
  ::-webkit-scrollbar-thumb:hover {
    background: rgba(242, 123, 47, 0.5);
  }
`;

type TranscriptMode = "sentences" | "paragraphs";
type SplitSource = "sentence" | "paragraph" | "custom";
type WaveEvent = "ready" | "audioprocess" | "interaction" | "finish" | "error";

interface WaveController {
  load: (url: string) => void;
  on: (event: WaveEvent, callback: () => void) => void;
  playPause: () => void;
  isPlaying: () => boolean;
  pause: () => void;
  play: (start?: number, end?: number) => void;
  getCurrentTime: () => number;
  getDuration: () => number;
  zoom: (value: number) => void;
  destroy: () => void;
}

function createWaveController(): WaveController {
  const audio = new Audio();
  const listeners: Record<WaveEvent, Array<() => void>> = {
    ready: [],
    audioprocess: [],
    interaction: [],
    finish: [],
    error: [],
  };

  let playing = false;
  let segmentEnd: number | null = null;

  const emit = (event: WaveEvent) => {
    for (const callback of listeners[event]) {
      callback();
    }
  };

  audio.addEventListener("canplay", () => emit("ready"));
  audio.addEventListener("timeupdate", () => {
    emit("audioprocess");
    emit("interaction");
    if (segmentEnd !== null && audio.currentTime >= segmentEnd) {
      audio.pause();
      playing = false;
      segmentEnd = null;
    }
  });
  audio.addEventListener("ended", () => {
    playing = false;
    segmentEnd = null;
    emit("finish");
  });
  audio.addEventListener("error", () => emit("error"));

  return {
    load: (url: string) => {
      audio.src = url;
      audio.load();
    },
    on: (event, callback) => {
      listeners[event].push(callback);
    },
    playPause: () => {
      if (audio.paused) {
        void audio.play();
        playing = true;
      } else {
        audio.pause();
        playing = false;
      }
    },
    isPlaying: () => playing && !audio.paused,
    pause: () => {
      audio.pause();
      playing = false;
      segmentEnd = null;
    },
    play: (start, end) => {
      if (typeof start === "number" && !Number.isNaN(start)) {
        audio.currentTime = Math.max(0, start);
      }
      segmentEnd = typeof end === "number" && !Number.isNaN(end) ? end : null;
      void audio.play();
      playing = true;
    },
    getCurrentTime: () => audio.currentTime || 0,
    getDuration: () => audio.duration || 0,
    zoom: () => {
      // no-op for native audio fallback
    },
    destroy: () => {
      audio.pause();
      audio.src = "";
      segmentEnd = null;
      playing = false;
    },
  };
}

const DEFAULT_ZOOM = 100;
const MIN_SPLIT_GAP = 0.2;

export default function AudioTranscribeStudio() {
  const waveformRef = useRef<HTMLDivElement | null>(null);
  const waveSurferRef = useRef<WaveController | null>(null);
  const playingSegmentEndRef = useRef<number | null>(null);
  const [file, setFile] = useState<File | null>(null);
  const [audioUrl, setAudioUrl] = useState<string | null>(null);
  const [response, setResponse] = useState<AudioTranscriptionResponse | null>(
    null
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const [previewSupported, setPreviewSupported] = useState(true);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [zoom, setZoom] = useState(DEFAULT_ZOOM);
  const zoomRef = useRef(DEFAULT_ZOOM);
  const zoomRafRef = useRef<number | null>(null);
  const [playing, setPlaying] = useState(false);
  const [playingSegmentId, setPlayingSegmentId] = useState<string | null>(null);
  const [transcriptMode, setTranscriptMode] = useState<TranscriptMode>(
    "sentences"
  );
  const [splitSource, setSplitSource] = useState<SplitSource>("sentence");
  const [customSplitTimes, setCustomSplitTimes] = useState<number[]>([]);

  // Inject scrollbar styles on mount
  useEffect(() => {
    const styleId = "audio-transcribe-scrollbar-styles";
    if (!document.getElementById(styleId)) {
      const style = document.createElement("style");
      style.id = styleId;
      // style.textContent = scrollbarStyles; // disabled for now
      document.head.appendChild(style);
    }
    return () => {
      // Keep styles for performance, don't remove
    };
  }, []);

  useEffect(() => {
    if (!audioUrl || !waveformRef.current || !previewSupported) return;

    const waveSurfer = createWaveController();

    waveSurfer.load(audioUrl);
    waveSurfer.on("ready", () => {
      setDuration(waveSurfer.getDuration());
      setCurrentTime(0);
    });
    waveSurfer.on("audioprocess", () => {
      const currentTime = waveSurfer.getCurrentTime();
      setCurrentTime(currentTime);

      // Check if segment playback has finished
      if (playingSegmentEndRef.current !== null && currentTime >= playingSegmentEndRef.current) {
        waveSurfer.pause();
        setPlayingSegmentId(null);
        playingSegmentEndRef.current = null;
      }
    });
    waveSurfer.on("interaction", () => {
      setCurrentTime(waveSurfer.getCurrentTime());
    });
    waveSurfer.on("finish", () => {
      setPlaying(false);
      setPlayingSegmentId(null);
    });
    waveSurfer.on("error", () => {
      setPreviewError(
        "Audio preview is not available in this browser. Transcription still works."
      );
      setPreviewSupported(false);
    });

    waveSurferRef.current = waveSurfer;

    return () => {
      waveSurfer.destroy();
      waveSurferRef.current = null;
      if (zoomRafRef.current) {
        cancelAnimationFrame(zoomRafRef.current);
        zoomRafRef.current = null;
      }
    };
  }, [audioUrl, previewSupported]);

  const applyZoom = useCallback((value: number) => {
    zoomRef.current = value;
    if (!waveSurferRef.current) return;
    if (zoomRafRef.current) {
      cancelAnimationFrame(zoomRafRef.current);
    }
    zoomRafRef.current = requestAnimationFrame(() => {
      waveSurferRef.current?.zoom(zoomRef.current);
    });
  }, []);

  useEffect(() => {
    if (!file) return;
    const supported = isSupportedAudio(file);
    setPreviewSupported(supported);
    setPreviewError(
      supported
        ? null
        : "Audio preview is not available in this browser. Transcription still works."
    );
    const url = URL.createObjectURL(file);
    setAudioUrl(url);
    setResponse(null);
    setError(null);
    setCustomSplitTimes([]);
    setSplitSource("sentence");
    setTranscriptMode("sentences");
    return () => URL.revokeObjectURL(url);
  }, [file]);

  const handleFileChange = useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const nextFile = event.target.files?.[0] || null;
      if (!nextFile) return;
      setFile(nextFile);
    },
    []
  );

  const handleTranscribe = useCallback(async () => {
    if (!file) return;
    setLoading(true);
    setError(null);
    try {
      const result = await transcribeAudio(file);
      if (result.error) {
        setError(result.error);
      } else if (result.data) {
        setResponse(result.data);
        setCustomSplitTimes([]);
        setSplitSource("sentence");
        setTranscriptMode("sentences");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setLoading(false);
    }
  }, [file]);

  const handlePlayToggle = useCallback(() => {
    if (!waveSurferRef.current) return;
    waveSurferRef.current.playPause();
    setPlaying(waveSurferRef.current.isPlaying());
    if (!waveSurferRef.current.isPlaying()) {
      setPlayingSegmentId(null);
      playingSegmentEndRef.current = null;
    }
  }, []);

  const handleSplitHere = useCallback(() => {
    if (!duration || !response) return;
    const time = Math.min(Math.max(currentTime, 0), duration);
    const existing = customSplitTimes.some(
      (value) => Math.abs(value - time) < MIN_SPLIT_GAP
    );
    if (existing) return;
    const updated = [...customSplitTimes, time].sort((a, b) => a - b);
    setCustomSplitTimes(updated);
    setSplitSource("custom");
  }, [currentTime, customSplitTimes, duration, response]);

  const handleClearFile = useCallback(() => {
    setFile(null);
    setAudioUrl(null);
    setResponse(null);
    setError(null);
    setCustomSplitTimes([]);
    setSplitSource("sentence");
    setTranscriptMode("sentences");
    setCurrentTime(0);
    setDuration(0);
    setPlayingSegmentId(null);
    playingSegmentEndRef.current = null;
  }, []);

  const transcriptSplits = useMemo(() => {
    if (!response) return [];
    return transcriptMode === "sentences"
      ? response.sentences
      : response.paragraphs;
  }, [response, transcriptMode]);

  const autoSplitCounts = useMemo(() => {
    return {
      sentence: response?.sentences.length || 0,
      paragraph: response?.paragraphs.length || 0,
    };
  }, [response]);

  const customSplits = useMemo(() => {
    if (!response || !duration || response.words.length === 0) return [];
    const points = [0, ...customSplitTimes, duration]
      .filter((value, index, arr) => index === 0 || value > arr[index - 1] + 0.05)
      .map((value) => Math.min(Math.max(value, 0), duration));

    const splits: TranscriptSplit[] = [];
    for (let i = 0; i < points.length - 1; i += 1) {
      const start = points[i];
      const end = points[i + 1];
      const words = response.words.filter(
        (word) => word.start >= start && word.end <= end
      );
      const text = words.map((word) => word.word.trim()).join(" ").trim();
      splits.push({
        id: `C${i + 1}`,
        start,
        end,
        text: text || "(No speech detected)",
        type: "custom",
      });
    }
    return splits;
  }, [customSplitTimes, duration, response]);

  const activeSplits = useMemo(() => {
    if (!response) return [];
    if (splitSource === "custom") return customSplits;
    return splitSource === "sentence"
      ? response.sentences
      : response.paragraphs;
  }, [customSplits, response, splitSource]);

  const handleSplitSource = useCallback((source: SplitSource) => {
    setSplitSource(source);
  }, []);

  const handlePlaySegment = useCallback((split: TranscriptSplit) => {
    if (!waveSurferRef.current) return;
    if (playingSegmentId === split.id) {
      // Stop playing if clicking the same segment
      waveSurferRef.current.pause();
      setPlaying(false);
      setPlayingSegmentId(null);
      playingSegmentEndRef.current = null;
    } else {
      // Play new segment
      waveSurferRef.current.play(split.start, split.end);
      setPlaying(true);
      setPlayingSegmentId(split.id);
      playingSegmentEndRef.current = split.end;
    }
  }, [playingSegmentId]);

  const handleRemoveSplit = useCallback(
    (index: number) => {
      if (splitSource !== "custom") return;
      const updated = [...customSplitTimes];
      updated.splice(index, 1);
      setCustomSplitTimes(updated);
    },
    [customSplitTimes, splitSource]
  );

  const downloadJSON = useCallback((splits: TranscriptSplit[]) => {
    const blob = new Blob([JSON.stringify(splits, null, 2)], {
      type: "application/json",
    });
    triggerDownload(blob, "transcript.json");
  }, []);

  const downloadSRT = useCallback((splits: TranscriptSplit[]) => {
    const lines = splits.map((split, index) => {
      const wrapped = wrapSubtitleLines(split.text, 42, 2);
      return [
        String(index + 1),
        `${formatSrtTime(split.start)} --> ${formatSrtTime(split.end)}`,
        wrapped,
        "",
      ].join("\n");
    });
    const blob = new Blob([lines.join("\n")], { type: "text/plain" });
    triggerDownload(blob, "transcript.srt");
  }, []);

  const captionRail = useMemo(() => {
    if (!duration || transcriptSplits.length === 0) return null;
    return (
      <div className="flex w-full gap-1.5 overflow-x-auto custom-scrollbar pb-2" style={{
        scrollbarWidth: 'thin',
        scrollbarColor: 'rgba(242, 123, 47, 0.4) rgba(0, 0, 0, 0.2)'
      }}>
        {transcriptSplits.map((split) => {
          const width = ((split.end - split.start) / duration) * 100;
          return (
            <div
              key={split.id}
              className="flex h-7 shrink-0 items-center gap-1.5 rounded-full bg-linear-to-r from-accent-orange/10 via-accent-orange/20 to-accent-orange/30 px-3 shadow-md transition hover:shadow-lg hover:opacity-90 cursor-pointer"
              style={{ minWidth: `max(3rem, ${Math.max(width, 15)}%)` }}
              title={split.text}
            >
              <span className="text-xs font-bold text-white/90">📌</span>
              <span className="truncate text-xs font-medium text-white">
                {split.text}
              </span>
            </div>
          );
        })}
      </div>
    );
  }, [duration, transcriptSplits]);

  return (
    <div className="min-h-screen bg-bg-mesh">
      <div className="mx-auto max-w-7xl">

        <div className="grid gap-6 lg:grid-cols-[1fr_340px]">
          <section className="min-w-0 space-y-5">
            <div className="rounded-2xl border border-white/10 bg-white/2 px-6 py-5">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p className="text-lg font-bold text-white">
                    {file ? file.name : "Drop an audio file"}
                  </p>
                  <p className="text-xs text-white/50">
                    {file
                      ? `Duration: ${formatTime(duration || 0)} · Size: ${formatBytes(file.size)}`
                      : "MP3, WAV, M4A, MP4 (max 30MB default)"}
                  </p>
                </div>
                <div className="flex flex-wrap items-center gap-2">
                  <label className="group relative inline-flex cursor-pointer items-center gap-2 rounded-lg border border-white/20 bg-white/5 px-3.5 py-2 text-xs font-medium text-white/70 transition hover:border-white/30 hover:bg-white/10">
                    <Upload className="h-3.5 w-3.5" />
                    <span>New File</span>
                    <input
                      type="file"
                      accept="audio/*"
                      onChange={handleFileChange}
                      className="absolute inset-0 cursor-pointer opacity-0"
                    />
                  </label>
                  {file && (
                    <button
                      type="button"
                      onClick={handleClearFile}
                      className="inline-flex items-center gap-2 rounded-lg border border-white/20 bg-white/5 px-3.5 py-2 text-xs font-medium text-white/70 transition hover:border-white/30 hover:bg-white/10"
                    >
                      <X className="h-3.5 w-3.5" />
                      Clear
                    </button>
                  )}
                </div>
              </div>
            </div>

            {(error || previewError) && (
              <div className="rounded-xl border border-red-900/50 bg-red-950/30 px-4 py-3 text-sm text-red-400">
                {error || previewError}
              </div>
            )}

            {response?.warnings?.length ? (
              <div className="rounded-xl border border-amber-900/50 bg-amber-950/30 px-4 py-3 text-xs text-amber-400">
                {response.warnings.join(" ")}
              </div>
            ) : null}

            <div className="rounded-2xl border border-white/10 bg-white/2 px-6 py-5">
              <div className="flex items-center gap-4">
                <button
                  type="button"
                  onClick={handlePlayToggle}
                  className="flex h-12 w-12 items-center justify-center rounded-full bg-accent-orange text-white shadow-md transition hover:bg-accent-orange/90"
                  disabled={!audioUrl}
                >
                  {playing ? <Pause className="h-5 w-5" /> : <Play className="h-5 w-5" />}
                </button>
                <p className="font-mono text-sm text-white/70">
                  {formatTime(currentTime)} / {formatTime(duration)}
                </p>
                <div className="ml-auto flex items-center gap-1.5">
                  <button
                    type="button"
                    onClick={() => { const v = Math.max(40, zoom - 20); setZoom(v); applyZoom(v); }}
                    className="flex h-7 w-7 items-center justify-center rounded-md border border-white/20 bg-white/5 text-white/50 transition hover:bg-white/10"
                  >
                    <Minus className="h-3.5 w-3.5" />
                  </button>
                  <span className="w-10 text-center text-xs font-medium text-white/50">{zoom}%</span>
                  <button
                    type="button"
                    onClick={() => { const v = Math.min(200, zoom + 20); setZoom(v); applyZoom(v); }}
                    className="flex h-7 w-7 items-center justify-center rounded-md border border-white/20 bg-white/5 text-white/50 transition hover:bg-white/10"
                  >
                    <Plus className="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>
              <div className="mt-4">
                {previewSupported ? (
                  <div className="max-w-full overflow-x-auto rounded-xl" style={{
                    scrollbarWidth: 'thin',
                    scrollbarColor: 'rgba(242, 123, 47, 0.4) rgba(0, 0, 0, 0.2)'
                  }}>
                    <div
                      ref={waveformRef}
                      className="min-w-full rounded-xl bg-black/50 p-4"
                    />
                  </div>
                ) : (
                  <div className="rounded-xl border border-dashed border-white/10 bg-black/20 px-4 py-6 text-center text-sm text-white/40">
                    Audio preview unavailable. You can still transcribe.
                    {audioUrl ? (
                      <audio
                        controls
                        src={audioUrl}
                        className="mt-3 w-full"
                      />
                    ) : null}
                  </div>
                )}
                {captionRail && (
                  <div className="mt-4 rounded-xl bg-black/40 p-3 overflow-hidden">
                    {captionRail}
                  </div>
                )}
              </div>
              <p className="mt-3 text-xs text-white/40">
                Click to seek · Drag orange line to scrub · Shift+Click to add split point
              </p>
            </div>

            <div className="flex flex-wrap items-center gap-3 border-t border-white/5 pt-4">
              <button
                type="button"
                onClick={handleSplitHere}
                disabled={!response}
                className="inline-flex items-center gap-2 rounded-lg bg-accent-orange/20 px-4 py-2 text-sm font-semibold text-accent-orange transition hover:bg-accent-orange/30 disabled:cursor-not-allowed disabled:opacity-60"
              >
                <Scissors className="h-4 w-4" />
                Split Here
              </button>
              <div className="ml-auto flex flex-wrap items-center gap-2 text-xs text-white/50">
                <span className="font-medium">Auto-split by:</span>
                <button
                  type="button"
                  onClick={() => handleSplitSource("sentence")}
                  className={`inline-flex items-center gap-1.5 rounded-lg border px-3 py-1.5 font-medium transition ${splitSource === "sentence"
                    ? "border-accent-orange/50 bg-accent-orange/20 text-white"
                    : "border-white/10 bg-white/5 text-white/50 hover:bg-white/10"
                    }`}
                >
                  <Scissors className="h-3 w-3" />
                  Sent. ({autoSplitCounts.sentence})
                </button>
                <button
                  type="button"
                  onClick={() => handleSplitSource("paragraph")}
                  className={`inline-flex items-center gap-1.5 rounded-lg border px-3 py-1.5 font-medium transition ${splitSource === "paragraph"
                    ? "border-accent-orange/50 bg-accent-orange/20 text-white"
                    : "border-white/10 bg-white/5 text-white/50 hover:bg-white/10"
                    }`}
                >
                  <Scissors className="h-3 w-3" />
                  Para. ({autoSplitCounts.paragraph})
                </button>
              </div>
            </div>

            <div className="rounded-2xl border border-white/10 bg-white/2 px-6 py-5">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-bold text-white">
                  Split Segments
                </h2>
                <p className="text-sm text-white/50">
                  {activeSplits.length} segments
                </p>
              </div>

              <div className="mt-4 space-y-3">
                {activeSplits.length === 0 ? (
                  <div className="rounded-xl border border-dashed border-white/10 bg-black/20 px-4 py-6 text-center text-sm text-white/40">
                    Upload audio and transcribe to see split segments.
                  </div>
                ) : (
                  activeSplits.map((split, index) => (
                    <div
                      key={`${split.id}-${index}`}
                      className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-accent-orange/20 bg-accent-orange/5 px-4 py-3"
                    >
                      <div>
                        <p className="text-sm font-bold text-white">
                          {splitSource === "paragraph" ? "Paragraph" : "Sentence"} {index + 1}
                        </p>
                        <p className="font-mono text-xs text-white/50">
                          {formatSrtTime(split.start)} - {formatSrtTime(split.end)}{" "}
                          <span className="text-white/30">
                            ({formatTime(split.end - split.start)})
                          </span>
                        </p>
                      </div>
                      <div className="flex items-center gap-2">
                        <button
                          type="button"
                          onClick={() => handlePlaySegment(split)}
                          className={`flex h-8 w-8 items-center justify-center rounded-lg border transition ${playingSegmentId === split.id
                            ? "border-accent-orange/50 bg-accent-orange/20 text-accent-orange"
                            : "border-white/20 bg-white/5 text-white/70 hover:bg-white/10"
                            }`}
                        >
                          {playingSegmentId === split.id ? (
                            <Pause className="h-3.5 w-3.5" />
                          ) : (
                            <Play className="h-3.5 w-3.5" />
                          )}
                        </button>
                        <button
                          type="button"
                          onClick={() => downloadJSON([split])}
                          className="inline-flex items-center gap-1.5 rounded-lg border border-white/20 bg-white/5 px-3 py-1.5 text-xs font-medium text-white/70 transition hover:bg-white/10"
                        >
                          <Download className="h-3.5 w-3.5" />
                          Download
                        </button>
                        {splitSource === "custom" && (
                          <button
                            type="button"
                            onClick={() => handleRemoveSplit(index)}
                            className="flex h-8 w-8 items-center justify-center rounded-lg border border-red-900/50 bg-red-950/20 text-red-400 transition hover:bg-red-950/40"
                          >
                            <X className="h-3.5 w-3.5" />
                          </button>
                        )}
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          </section>

          <aside className="min-w-0 self-start rounded-2xl border border-white/10 bg-white/2 px-6 py-5 lg:sticky lg:top-6">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-bold text-white">Transcript</h2>
              <button
                type="button"
                onClick={handleTranscribe}
                disabled={!file || loading}
                className="inline-flex items-center gap-2 rounded-lg border border-accent-orange/60 bg-accent-orange/80 px-3.5 py-2 text-xs font-semibold text-white transition hover:border-accent-orange hover:bg-accent-orange disabled:opacity-50 disabled:cursor-not-allowed shadow-sm"
              >
                {loading ? (
                  <>
                    <svg className="h-3.5 w-3.5 animate-spin" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Processing...
                  </>
                ) : (
                  <>
                    Transcribe{response && <span className="text-white/70">(Done)</span>}
                  </>
                )}
              </button>
            </div>

            <div className="mt-4 rounded-xl bg-black/30 p-1">
              <div className="flex">
                <button
                  type="button"
                  onClick={() => setTranscriptMode("sentences")}
                  className={`flex-1 rounded-lg px-4 py-2 text-xs font-semibold transition ${transcriptMode === "sentences"
                    ? "bg-accent-orange/20 text-white shadow-sm"
                    : "text-white/40 hover:text-white/60"
                    }`}
                >
                  Sentences
                </button>
                <button
                  type="button"
                  onClick={() => setTranscriptMode("paragraphs")}
                  className={`flex-1 rounded-lg px-4 py-2 text-xs font-semibold transition ${transcriptMode === "paragraphs"
                    ? "bg-accent-orange/20 text-white shadow-sm"
                    : "text-white/40 hover:text-white/60"
                    }`}
                >
                  Paragraphs
                </button>
              </div>
            </div>

            <div className="mt-5 space-y-4">
              {transcriptSplits.length === 0 ? (
                <div className="rounded-xl border border-dashed border-white/10 bg-black/20 px-4 py-8 text-center text-sm text-white/40">
                  Transcript will appear here after processing.
                </div>
              ) : (
                transcriptSplits.map((split, index) => (
                  <div key={split.id} className="flex gap-3 border-b border-white/5 pb-4 last:border-0">
                    <span className="shrink-0 text-xs font-bold text-white/40">
                      {transcriptMode === "sentences" ? `S${index + 1}` : `P${index + 1}`}
                    </span>
                    <p className="text-sm leading-relaxed text-white/70">{split.text}</p>
                  </div>
                ))
              )}
            </div>

            <div className="mt-5 flex gap-3">
              <button
                type="button"
                onClick={() => downloadJSON(transcriptSplits)}
                disabled={transcriptSplits.length === 0}
                className="flex flex-1 items-center justify-center gap-2 rounded-lg border border-white/20 bg-white/5 py-2.5 text-xs font-medium text-white/70 transition hover:bg-white/10 disabled:opacity-50"
              >
                <Download className="h-3.5 w-3.5" />
                JSON
              </button>
              <button
                type="button"
                onClick={() => downloadSRT(transcriptSplits)}
                disabled={transcriptSplits.length === 0}
                className="flex flex-1 items-center justify-center gap-2 rounded-lg border border-white/20 bg-white/5 py-2.5 text-xs font-medium text-white/70 transition hover:bg-white/10 disabled:opacity-50"
              >
                <Download className="h-3.5 w-3.5" />
                SRT
              </button>
            </div>
          </aside>
        </div>
      </div>
    </div>
  );
}

function formatTime(seconds: number) {
  if (!Number.isFinite(seconds)) return "0:00";
  const mins = Math.floor(seconds / 60);
  const secs = Math.floor(seconds % 60);
  return `${mins}:${secs.toString().padStart(2, "0")}`;
}

function formatSrtTime(seconds: number) {
  const total = Math.max(seconds, 0);
  const hours = Math.floor(total / 3600);
  const mins = Math.floor((total % 3600) / 60);
  const secs = Math.floor(total % 60);
  const ms = Math.floor((total - Math.floor(total)) * 1000);
  return `${hours.toString().padStart(2, "0")}:${mins
    .toString()
    .padStart(2, "0")}:${secs.toString().padStart(2, "0")},${ms
      .toString()
      .padStart(3, "0")}`;
}

function formatBytes(bytes: number) {
  if (bytes >= 1 << 20) return `${(bytes / (1 << 20)).toFixed(2)} MB`;
  if (bytes >= 1 << 10) return `${(bytes / (1 << 10)).toFixed(1)} KB`;
  return `${bytes} B`;
}

function triggerDownload(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = filename;
  anchor.click();
  URL.revokeObjectURL(url);
}

function wrapSubtitleLines(text: string, maxChars: number, maxLines: number) {
  const words = text.trim().split(/\s+/).filter(Boolean);
  if (words.length === 0) return "";

  const lines: string[] = [];
  let current = "";

  for (const word of words) {
    const next = current ? `${current} ${word}` : word;
    if (next.length <= maxChars || lines.length >= maxLines - 1) {
      current = next;
    } else {
      lines.push(current);
      current = word;
    }
  }

  if (current) {
    lines.push(current);
  }

  if (lines.length > maxLines) {
    const clipped = lines.slice(0, maxLines);
    clipped[maxLines - 1] = clipped.slice(maxLines - 1).join(" ");
    return clipped.join("\n");
  }

  return lines.join("\n");
}

function isSupportedAudio(file: File) {
  const audio = document.createElement("audio");
  const type = file.type || "";
  if (type && audio.canPlayType(type)) {
    return true;
  }

  const ext = file.name.split(".").pop()?.toLowerCase() || "";
  const mimeMap: Record<string, string> = {
    mp3: "audio/mpeg",
    wav: "audio/wav",
    m4a: "audio/mp4",
    mp4: "audio/mp4",
    webm: "audio/webm",
    ogg: "audio/ogg",
    flac: "audio/flac",
  };
  const inferred = mimeMap[ext];
  if (!inferred) return false;
  return Boolean(audio.canPlayType(inferred));
}
