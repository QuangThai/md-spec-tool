export interface TranscriptWord {
  word: string;
  start: number;
  end: number;
  confidence?: number;
}

export interface TranscriptSegment {
  id: number;
  start: number;
  end: number;
  text: string;
}

export type TranscriptSplitType = "sentence" | "paragraph" | "custom";

export interface TranscriptSplit {
  id: string;
  start: number;
  end: number;
  text: string;
  type: TranscriptSplitType;
}

export interface AudioTranscriptionResponse {
  text: string;
  language?: string;
  duration: number;
  segments: TranscriptSegment[];
  words: TranscriptWord[];
  sentences: TranscriptSplit[];
  paragraphs: TranscriptSplit[];
  warnings?: string[];
}
