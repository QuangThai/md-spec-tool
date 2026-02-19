import { PreviewResponse } from "@/lib/types";

export function buildReviewRequiredColumns(preview: PreviewResponse | null): string[] {
  if (!preview) return [];
  const lowConfidence = preview.mapping_quality?.low_confidence_columns ?? [];
  return Array.from(new Set(lowConfidence));
}

export function countRemainingReviews(
  requiredColumns: string[],
  reviewedColumns: Record<string, boolean>
): number {
  return requiredColumns.filter((column) => !reviewedColumns[column]).length;
}

export function canConfirmReview(
  requiredColumns: string[],
  reviewedColumns: Record<string, boolean>
): boolean {
  if (requiredColumns.length === 0) return true;
  return countRemainingReviews(requiredColumns, reviewedColumns) === 0;
}
