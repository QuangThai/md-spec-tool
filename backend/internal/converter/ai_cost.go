package converter

import "strings"

type aiTokenRates struct {
	inputPer1MUSD  float64
	outputPer1MUSD float64
}

var aiRatesByModel = map[string]aiTokenRates{
	"gpt-4o-mini":  {inputPer1MUSD: 0.15, outputPer1MUSD: 0.60},
	"gpt-4.1-mini": {inputPer1MUSD: 0.40, outputPer1MUSD: 1.60},
	"gpt-4.1-nano": {inputPer1MUSD: 0.10, outputPer1MUSD: 0.40},
}

func roughTokenEstimate(text string) int {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return 0
	}
	// Rough heuristic: ~4 chars/token for mixed English/table payloads.
	return (len(trimmed) + 3) / 4
}

func estimateAIMappingCostUSD(model string, inputTokens, outputTokens int) float64 {
	if inputTokens <= 0 && outputTokens <= 0 {
		return 0
	}
	rates, ok := aiRatesByModel[strings.ToLower(strings.TrimSpace(model))]
	if !ok {
		// Conservative fallback for unknown models.
		rates = aiTokenRates{inputPer1MUSD: 0.50, outputPer1MUSD: 2.00}
	}
	inputCost := (float64(inputTokens) / 1_000_000.0) * rates.inputPer1MUSD
	outputCost := (float64(outputTokens) / 1_000_000.0) * rates.outputPer1MUSD
	return inputCost + outputCost
}
