import { backendClient } from "@/lib/httpClient";
import { ApiResult } from "@/lib/types";
import { AudioTranscriptionResponse } from "@/lib/audioTypes";

export async function transcribeAudio(
  file: File
): Promise<ApiResult<AudioTranscriptionResponse>> {
  const formData = new FormData();
  formData.append("file", file);
  return backendClient.safePost<AudioTranscriptionResponse>(
    "/api/audio/transcribe",
    formData
  );
}
