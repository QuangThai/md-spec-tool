package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/md-spec-tool/internal/config"
)

const (
	openAITranscribeURL    = "https://api.openai.com/v1/audio/transcriptions"
	openAITranscribeModel  = "whisper-1"
	maxOpenAIUploadBytes   = 25 << 20
	segmentDurationSeconds = 600.0
	segmentMinSeconds      = 30.0
	silenceThresholdDB     = "-30dB"
	silenceMinDuration     = 0.4
)

type AudioTranscribeHandler struct {
	cfg        *config.Config
	httpClient *http.Client
}

type OpenAIWord struct {
	Word       string  `json:"word"`
	Start      float64 `json:"start"`
	End        float64 `json:"end"`
	Confidence float64 `json:"confidence,omitempty"`
}

type OpenAISegment struct {
	ID    int     `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

type OpenAITranscriptionResponse struct {
	Text     string          `json:"text"`
	Language string          `json:"language"`
	Duration float64         `json:"duration"`
	Segments []OpenAISegment `json:"segments"`
	Words    []OpenAIWord    `json:"words"`
}

type TranscriptSplit struct {
	ID    string  `json:"id"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
	Type  string  `json:"type"`
}

type AudioTranscribeResponse struct {
	Text       string            `json:"text"`
	Language   string            `json:"language,omitempty"`
	Duration   float64           `json:"duration"`
	Segments   []OpenAISegment   `json:"segments"`
	Words      []OpenAIWord      `json:"words"`
	Sentences  []TranscriptSplit `json:"sentences"`
	Paragraphs []TranscriptSplit `json:"paragraphs"`
	Warnings   []string          `json:"warnings,omitempty"`
}

func NewAudioTranscribeHandler(cfg *config.Config) *AudioTranscribeHandler {
	timeout := cfg.AIRequestTimeout
	if timeout < 120*time.Second {
		timeout = 120 * time.Second
	}

	return &AudioTranscribeHandler{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: timeout + 30*time.Second,
		},
	}
}

func (h *AudioTranscribeHandler) Transcribe(c *gin.Context) {
	apiKey := strings.TrimSpace(getUserAPIKey(c))
	if apiKey == "" {
		apiKey = strings.TrimSpace(h.cfg.OpenAIAPIKey)
	}
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "OpenAI API key not configured"})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.MaxAudioUploadBytes+(1<<20))

	file, header, err := c.Request.FormFile("file")
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("file exceeds %s limit", humanSize(h.cfg.MaxAudioUploadBytes))})
			return
		}
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	if header.Size > h.cfg.MaxAudioUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("file exceeds %s limit", humanSize(h.cfg.MaxAudioUploadBytes))})
		return
	}

	if err := validateAudioExtension(header.Filename); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	tempFile, err := os.CreateTemp("", "audio-upload-*"+filepath.Ext(header.Filename))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: fmt.Sprintf("failed to process file: %v", err)})
		return
	}
	tempName := tempFile.Name()
	defer os.Remove(tempName)

	bytesCopied, err := io.Copy(tempFile, io.LimitReader(file, h.cfg.MaxAudioUploadBytes+1))
	if err != nil {
		_ = tempFile.Close()
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to save file"})
		return
	}
	if bytesCopied > h.cfg.MaxAudioUploadBytes {
		_ = tempFile.Close()
		c.JSON(http.StatusRequestEntityTooLarge, ErrorResponse{Error: fmt.Sprintf("file exceeds %s limit", humanSize(h.cfg.MaxAudioUploadBytes))})
		return
	}
	if err := tempFile.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to finalize upload"})
		return
	}

	ctx := c.Request.Context()
	response, warnings, err := h.transcribeWithChunking(ctx, tempName, apiKey)
	if err != nil {
		c.JSON(http.StatusBadGateway, ErrorResponse{Error: err.Error()})
		return
	}
	response.Warnings = warnings

	c.JSON(http.StatusOK, response)
}

func validateAudioExtension(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp3", ".mp4", ".mpeg", ".mpga", ".m4a", ".wav", ".webm", ".ogg", ".flac":
		return nil
	default:
		return fmt.Errorf("unsupported audio format")
	}
}

func (h *AudioTranscribeHandler) transcribeWithChunking(ctx context.Context, filePath, apiKey string) (*AudioTranscribeResponse, []string, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read audio file: %w", err)
	}

	if info.Size() <= maxOpenAIUploadBytes {
		resp, err := h.requestTranscription(ctx, filePath, apiKey)
		if err != nil {
			return nil, nil, err
		}
		return h.buildResponse(resp), nil, nil
	}

	warnings := []string{"Audio exceeds 25MB; server-side chunking enabled."}
	chunkPaths, err := splitAudioWithSilence(filePath)
	if err != nil {
		warnings = append(warnings, "Falling back to fixed-length chunking.")
		chunkPaths, err = splitAudioFixed(filePath)
		if err != nil {
			return nil, warnings, err
		}
	}
	defer cleanupFiles(chunkPaths)

	combined := OpenAITranscriptionResponse{}
	var textParts []string
	offset := 0.0

	for _, chunkPath := range chunkPaths {
		resp, err := h.requestTranscription(ctx, chunkPath, apiKey)
		if err != nil {
			return nil, warnings, err
		}
		textParts = append(textParts, strings.TrimSpace(resp.Text))

		for _, word := range resp.Words {
			combined.Words = append(combined.Words, OpenAIWord{
				Word:       word.Word,
				Start:      word.Start + offset,
				End:        word.End + offset,
				Confidence: word.Confidence,
			})
		}
		for _, segment := range resp.Segments {
			combined.Segments = append(combined.Segments, OpenAISegment{
				ID:    segment.ID + len(combined.Segments),
				Start: segment.Start + offset,
				End:   segment.End + offset,
				Text:  segment.Text,
			})
		}
		if combined.Language == "" {
			combined.Language = resp.Language
		}
		if resp.Duration > 0 {
			offset += resp.Duration
		} else {
			offset += segmentDurationSeconds
		}
	}

	combined.Text = strings.TrimSpace(strings.Join(textParts, " "))
	combined.Duration = offset

	response := h.buildResponse(&combined)
	return response, warnings, nil
}

func splitAudioWithSilence(filePath string) ([]string, error) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not available for chunking; please upload a file under 25MB or install ffmpeg")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return nil, fmt.Errorf("ffprobe not available for chunking; please upload a file under 25MB or install ffmpeg")
	}

	duration, err := probeDuration(filePath)
	if err != nil || duration <= 0 {
		return nil, fmt.Errorf("failed to read audio duration")
	}

	silences, err := detectSilence(filePath)
	if err != nil {
		return nil, err
	}

	segments := buildSegments(duration, silences)
	if len(segments) == 0 {
		return nil, fmt.Errorf("failed to build chunk segments")
	}

	return renderSegments(filePath, segments)
}

func splitAudioFixed(filePath string) ([]string, error) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg not available for chunking; please upload a file under 25MB or install ffmpeg")
	}

	tempDir, err := os.MkdirTemp("", "audio-chunks-*")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare audio chunks")
	}

	outputPattern := filepath.Join(tempDir, "chunk-%03d.wav")
	args := []string{
		"-y",
		"-i", filePath,
		"-f", "segment",
		"-segment_time", fmt.Sprintf("%.0f", segmentDurationSeconds),
		"-reset_timestamps", "1",
		"-ac", "1",
		"-ar", "16000",
		outputPattern,
	}

	cmd := exec.Command("ffmpeg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		_ = os.RemoveAll(tempDir)
		return nil, fmt.Errorf("ffmpeg chunking failed: %s", strings.TrimSpace(string(output)))
	}

	chunkPaths, err := filepath.Glob(filepath.Join(tempDir, "chunk-*.wav"))
	if err != nil || len(chunkPaths) == 0 {
		_ = os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to generate audio chunks")
	}

	sort.Strings(chunkPaths)
	return chunkPaths, nil
}

type silenceInterval struct {
	Start float64
	End   float64
}

type segmentInterval struct {
	Start float64
	End   float64
}

func probeDuration(filePath string) (float64, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=nw=1:nk=1", filePath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	trimmed := strings.TrimSpace(string(output))
	return strconv.ParseFloat(trimmed, 64)
}

func detectSilence(filePath string) ([]silenceInterval, error) {
	args := []string{
		"-i", filePath,
		"-af", fmt.Sprintf("silencedetect=noise=%s:d=%.2f", silenceThresholdDB, silenceMinDuration),
		"-f", "null",
		"-",
	}
	cmd := exec.Command("ffmpeg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("silence detection failed: %s", strings.TrimSpace(string(output)))
	}

	lines := strings.Split(string(output), "\n")
	intervals := []silenceInterval{}
	var current *silenceInterval
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "silence_start") {
			value := parseSilenceValue(line, "silence_start")
			if value >= 0 {
				current = &silenceInterval{Start: value}
			}
		}
		if strings.Contains(line, "silence_end") {
			value := parseSilenceValue(line, "silence_end")
			if value >= 0 {
				if current == nil {
					current = &silenceInterval{Start: value}
				}
				current.End = value
				intervals = append(intervals, *current)
				current = nil
			}
		}
	}

	return intervals, nil
}

func parseSilenceValue(line, key string) float64 {
	idx := strings.Index(line, key)
	if idx == -1 {
		return -1
	}
	fragment := line[idx+len(key):]
	fragment = strings.TrimSpace(strings.TrimPrefix(fragment, ":"))
	parts := strings.Fields(fragment)
	if len(parts) == 0 {
		return -1
	}
	value, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return -1
	}
	return value
}

func buildSegments(duration float64, silences []silenceInterval) []segmentInterval {
	if duration <= 0 {
		return nil
	}
	var segments []segmentInterval
	start := 0.0
	for start < duration {
		idealEnd := start + segmentDurationSeconds
		if idealEnd > duration {
			idealEnd = duration
		}

		cut := idealEnd
		for _, silence := range silences {
			if silence.End <= start+segmentMinSeconds {
				continue
			}
			if silence.End > idealEnd {
				break
			}
			cut = silence.End
		}

		if cut <= start {
			cut = idealEnd
		}

		segments = append(segments, segmentInterval{Start: start, End: cut})
		start = cut
	}

	return segments
}

func renderSegments(filePath string, segments []segmentInterval) ([]string, error) {
	tempDir, err := os.MkdirTemp("", "audio-chunks-*")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare audio chunks")
	}

	var chunkPaths []string
	for index, segment := range segments {
		outputPath := filepath.Join(tempDir, fmt.Sprintf("chunk-%03d.wav", index+1))
		args := []string{
			"-y",
			"-i", filePath,
			"-ss", fmt.Sprintf("%.3f", segment.Start),
			"-to", fmt.Sprintf("%.3f", segment.End),
			"-ac", "1",
			"-ar", "16000",
			outputPath,
		}
		cmd := exec.Command("ffmpeg", args...)
		if output, err := cmd.CombinedOutput(); err != nil {
			_ = os.RemoveAll(tempDir)
			return nil, fmt.Errorf("ffmpeg chunking failed: %s", strings.TrimSpace(string(output)))
		}
		chunkPaths = append(chunkPaths, outputPath)
	}

	return chunkPaths, nil
}

func cleanupFiles(files []string) {
	visited := map[string]struct{}{}
	for _, file := range files {
		if file == "" {
			continue
		}
		if _, ok := visited[file]; ok {
			continue
		}
		visited[file] = struct{}{}
		_ = os.Remove(file)
		_ = os.RemoveAll(filepath.Dir(file))
	}
}

func (h *AudioTranscribeHandler) requestTranscription(ctx context.Context, filePath, apiKey string) (*OpenAITranscriptionResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio: %w", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to prepare upload: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to read audio: %w", err)
	}

	if err := writer.WriteField("model", openAITranscribeModel); err != nil {
		return nil, fmt.Errorf("failed to write model field: %w", err)
	}
	if err := writer.WriteField("response_format", "verbose_json"); err != nil {
		return nil, fmt.Errorf("failed to write response_format field: %w", err)
	}
	if err := writer.WriteField("timestamp_granularities[]", "word"); err != nil {
		return nil, fmt.Errorf("failed to write timestamp_granularities field: %w", err)
	}
	if err := writer.WriteField("timestamp_granularities[]", "segment"); err != nil {
		return nil, fmt.Errorf("failed to write timestamp_granularities field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to prepare request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAITranscribeURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := h.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("transcription request timeout: %w", err)
		}
		return nil, fmt.Errorf("transcription request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		errMsg := strings.TrimSpace(string(payload))
		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			return nil, fmt.Errorf("openai authentication failed: %s", errMsg)
		}
		if resp.StatusCode == 429 {
			return nil, fmt.Errorf("openai rate limited: %s", errMsg)
		}
		return nil, fmt.Errorf("transcription failed (status %d): %s", resp.StatusCode, errMsg)
	}

	var response OpenAITranscriptionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse transcription: %w", err)
	}

	return &response, nil
}

func (h *AudioTranscribeHandler) buildResponse(resp *OpenAITranscriptionResponse) *AudioTranscribeResponse {
	sentences := buildSentenceSplits(resp.Words, resp.Segments)
	paragraphs := buildParagraphSplits(sentences)
	return &AudioTranscribeResponse{
		Text:       resp.Text,
		Language:   resp.Language,
		Duration:   resp.Duration,
		Segments:   resp.Segments,
		Words:      resp.Words,
		Sentences:  sentences,
		Paragraphs: paragraphs,
	}
}

func buildSentenceSplits(words []OpenAIWord, segments []OpenAISegment) []TranscriptSplit {
	const sentenceGap = 0.6
	const minSentenceDuration = 0.4

	cleaned := filterAndSortWords(words)
	if len(cleaned) == 0 {
		return segmentsToSplits(segments)
	}

	var results []TranscriptSplit
	var builder strings.Builder
	var start float64
	var end float64
	sentenceIndex := 1

	appendWord := func(word string) {
		trimmed := strings.TrimSpace(word)
		if trimmed == "" {
			return
		}
		if builder.Len() > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(trimmed)
	}

	flush := func() {
		text := strings.TrimSpace(builder.String())
		if text == "" {
			builder.Reset()
			return
		}
		results = append(results, TranscriptSplit{
			ID:    fmt.Sprintf("S%d", sentenceIndex),
			Start: start,
			End:   end,
			Text:  text,
			Type:  "sentence",
		})
		sentenceIndex++
		builder.Reset()
	}

	for i, word := range cleaned {
		if builder.Len() == 0 {
			start = word.Start
		}
		appendWord(word.Word)
		end = word.End

		gap := 0.0
		if i < len(cleaned)-1 {
			gap = cleaned[i+1].Start - word.End
		}
		if isSentenceTerminal(word.Word) || gap >= sentenceGap {
			flush()
		}
	}

	flush()
	return mergeShortSentences(results, minSentenceDuration)
}

func filterAndSortWords(words []OpenAIWord) []OpenAIWord {
	filtered := make([]OpenAIWord, 0, len(words))
	for _, word := range words {
		if strings.TrimSpace(word.Word) == "" {
			continue
		}
		if word.End <= word.Start {
			continue
		}
		filtered = append(filtered, word)
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Start < filtered[j].Start
	})
	return filtered
}

func segmentsToSplits(segments []OpenAISegment) []TranscriptSplit {
	if len(segments) == 0 {
		return nil
	}
	results := make([]TranscriptSplit, 0, len(segments))
	for i, segment := range segments {
		text := strings.TrimSpace(segment.Text)
		if text == "" {
			continue
		}
		results = append(results, TranscriptSplit{
			ID:    fmt.Sprintf("S%d", i+1),
			Start: segment.Start,
			End:   segment.End,
			Text:  text,
			Type:  "sentence",
		})
	}
	return results
}

func mergeShortSentences(sentences []TranscriptSplit, minDuration float64) []TranscriptSplit {
	if len(sentences) == 0 {
		return nil
	}
	var merged []TranscriptSplit
	for _, sentence := range sentences {
		if len(merged) == 0 {
			merged = append(merged, sentence)
			continue
		}
		last := &merged[len(merged)-1]
		if (sentence.End - sentence.Start) < minDuration {
			last.End = sentence.End
			last.Text = strings.TrimSpace(last.Text + " " + sentence.Text)
			continue
		}
		merged = append(merged, sentence)
	}
	for i := range merged {
		merged[i].ID = fmt.Sprintf("S%d", i+1)
	}
	return merged
}

func buildParagraphSplits(sentences []TranscriptSplit) []TranscriptSplit {
	const paragraphGap = 1.6
	const maxSentences = 4

	if len(sentences) == 0 {
		return nil
	}

	var results []TranscriptSplit
	paraIndex := 1
	count := 0
	current := TranscriptSplit{Type: "paragraph"}

	for i, sentence := range sentences {
		if count == 0 {
			current = TranscriptSplit{
				ID:    fmt.Sprintf("P%d", paraIndex),
				Start: sentence.Start,
				End:   sentence.End,
				Text:  sentence.Text,
				Type:  "paragraph",
			}
			count = 1
		} else {
			current.End = sentence.End
			current.Text = strings.TrimSpace(current.Text + " " + sentence.Text)
			count++
		}

		gap := 0.0
		if i < len(sentences)-1 {
			gap = sentences[i+1].Start - sentence.End
		}

		if gap >= paragraphGap || count >= maxSentences || i == len(sentences)-1 {
			results = append(results, current)
			paraIndex++
			count = 0
		}
	}

	return results
}

func isSentenceTerminal(word string) bool {
	trimmed := strings.TrimSpace(word)
	if trimmed == "" {
		return false
	}
	last := trimmed[len(trimmed)-1]
	return last == '.' || last == '!' || last == '?'
}
