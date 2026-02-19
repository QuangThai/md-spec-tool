'use client';

import { useState, useCallback } from 'react';
import { AnimatePresence, motion } from 'framer-motion';
import {
  AlertTriangle,
  CheckCircle,
  XCircle,
  ArrowRight,
  RefreshCw,
  Edit2,
  Save,
  X,
} from 'lucide-react';
import { ColumnMapping, MDFlowWarning } from '@/lib/types';
import { Button } from '@/components/ui/Button';
import { Badge } from '@/components/ui/Badge';
import { useBodyScrollLock } from '@/lib/useBodyScrollLock';

interface ColumnMappingReviewProps {
  mappings: ColumnMapping[];
  sourceHeaders: string[];
  warnings: MDFlowWarning[];
  avgConfidence: number;
  onConfirm: (updatedMappings: ColumnMapping[]) => void;
  onRetry: (columnOverrides: Record<string, string>) => void;
  isOpen: boolean;
  onClose: () => void;
}

interface EditingMapping extends ColumnMapping {
  isEditing: boolean;
  newCanonicalName?: string;
}

export function ColumnMappingReview({
  mappings,
  sourceHeaders,
  warnings,
  avgConfidence,
  onConfirm,
  onRetry,
  isOpen,
  onClose,
}: ColumnMappingReviewProps) {
  useBodyScrollLock(isOpen);
  const [editingMappings, setEditingMappings] = useState<EditingMapping[]>(
    mappings.map((m) => ({ ...m, isEditing: false }))
  );
  const [selectedMode, setSelectedMode] = useState<'confirm' | 'retry'>(
    'confirm'
  );

  const lowConfidenceCount = editingMappings.filter(
    (m) => m.confidence < 0.65
  ).length;

  // Get warnings related to low confidence
  const confidenceWarnings = warnings.filter(
    (w) =>
      w.category === 'mapping' &&
      (w.code === 'LOW_MAPPING_CONFIDENCE' || w.code === 'UNMAPPED_COLUMNS')
  );

  const handleToggleEdit = (columnIndex: number) => {
    setEditingMappings((prev) =>
      prev.map((m) =>
        m.column_index === columnIndex
          ? { ...m, isEditing: !m.isEditing }
          : m
      )
    );
  };

  const handleEditMapping = (columnIndex: number, newName: string) => {
    setEditingMappings((prev) =>
      prev.map((m) =>
        m.column_index === columnIndex
          ? { ...m, canonical_name: newName, isEditing: false }
          : m
      )
    );
  };

  const handleConfirm = () => {
    const confirmedMappings = editingMappings.map((m) => {
      const { isEditing, newCanonicalName, ...rest } = m;
      return rest as ColumnMapping;
    });
    onConfirm(confirmedMappings);
    onClose();
  };

  const handleRetry = () => {
    // Build column overrides from edited mappings
    const overrides: Record<string, string> = {};
    editingMappings.forEach((m) => {
      if (m.canonical_name) {
        overrides[m.source_header] = m.canonical_name;
      }
    });
    onRetry(overrides);
    onClose();
  };

  const getConfidenceBadgeColor = (confidence: number) => {
    if (confidence >= 0.80) return 'bg-green-100 text-green-800';
    if (confidence >= 0.65) return 'bg-yellow-100 text-yellow-800';
    return 'bg-red-100 text-red-800';
  };

  const getConfidenceIcon = (confidence: number) => {
    if (confidence >= 0.80) {
      return <CheckCircle className="h-4 w-4 text-green-600" />;
    }
    if (confidence >= 0.65) {
      return <AlertTriangle className="h-4 w-4 text-yellow-600" />;
    }
    return <XCircle className="h-4 w-4 text-red-600" />;
  };

  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          onClick={onClose}
          className="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
        >
          <motion.div
            initial={{ scale: 0.95, opacity: 0, y: 20 }}
            animate={{ scale: 1, opacity: 1, y: 0 }}
            exit={{ scale: 0.95, opacity: 0, y: 20 }}
            onClick={(e) => e.stopPropagation()}
            className="bg-white border border-slate-200 rounded-lg shadow-xl max-w-2xl w-full max-h-[80vh] flex flex-col overflow-hidden"
          >
            {/* Header */}
            <div className="flex items-center justify-between gap-4 px-6 py-4 border-b border-slate-200 shrink-0">
              <div className="flex items-center gap-2">
                <AlertTriangle className="h-5 w-5 text-yellow-600" />
                <div>
                  <h2 className="text-lg font-semibold text-slate-900">
                    Review Column Mappings
                  </h2>
                  <p className="text-sm text-slate-600 mt-0.5">
                    Mapping confidence: {Math.round(avgConfidence * 100)}% •{' '}
                    {lowConfidenceCount} low-confidence{' '}
                    {lowConfidenceCount === 1 ? 'column' : 'columns'}
                  </p>
                </div>
              </div>
              <Button
                variant="secondary"
                size="sm"
                onClick={onClose}
                className="h-8 w-8 p-0"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>

            {/* Content */}
            <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
              {/* Warnings Summary */}
              {confidenceWarnings.length > 0 && (
                <div className="space-y-2 p-3 bg-yellow-50 rounded-md border border-yellow-200">
                  {confidenceWarnings.map((w, idx) => (
                    <div key={idx} className="flex gap-2 text-sm">
                      <AlertTriangle className="h-4 w-4 text-yellow-600 shrink-0 mt-0.5" />
                      <div className="flex-1">
                        <p className="font-medium text-yellow-900">
                          {w.message}
                        </p>
                        {w.hint && (
                          <p className="text-xs text-yellow-800 mt-1">
                            {w.hint}
                          </p>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}

              {/* Confidence Distribution */}
              <div className="grid grid-cols-3 gap-3 p-3 bg-slate-50 rounded-md border border-slate-200">
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">
                    {editingMappings.filter((m) => m.confidence >= 0.80)
                      .length}
                  </div>
                  <p className="text-xs text-slate-600">High ≥80%</p>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-yellow-600">
                    {
                      editingMappings.filter(
                        (m) => m.confidence >= 0.65 && m.confidence < 0.80
                      ).length
                    }
                  </div>
                  <p className="text-xs text-slate-600">Medium 65–79%</p>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-red-600">
                    {editingMappings.filter((m) => m.confidence < 0.65).length}
                  </div>
                  <p className="text-xs text-slate-600">Low &lt;65%</p>
                </div>
              </div>

              {/* Mapping Table */}
              <div className="space-y-2">
                {editingMappings.map((m) => (
                  <div
                    key={m.column_index}
                    className={`flex items-center gap-3 p-3 rounded-md border transition-colors ${
                      m.confidence < 0.65
                        ? 'bg-red-50 border-red-200'
                        : m.confidence < 0.80
                          ? 'bg-yellow-50 border-yellow-200'
                          : 'bg-green-50 border-green-200'
                    }`}
                  >
                    {/* Icon */}
                    <div className="shrink-0">
                      {getConfidenceIcon(m.confidence)}
                    </div>

                    {/* Source Header */}
                    <div className="flex-1 min-w-[120px]">
                      <p className="text-sm font-medium text-slate-900">
                        {m.source_header}
                      </p>
                      {m.reasoning && (
                        <p className="text-xs text-slate-600 mt-1">
                          {m.reasoning}
                        </p>
                      )}
                    </div>

                    {/* Arrow */}
                    <ArrowRight className="h-4 w-4 text-slate-400" />

                    {/* Canonical Name (Editable or Display) */}
                    {m.isEditing ? (
                      <div className="flex items-center gap-1">
                        <input
                          type="text"
                          value={m.canonical_name}
                          onChange={(e) =>
                            handleEditMapping(
                              m.column_index,
                              e.target.value
                            )
                          }
                          className="px-2 py-1 border border-slate-300 rounded text-sm font-mono w-[150px]"
                          placeholder="Enter field name"
                        />
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={() => handleToggleEdit(m.column_index)}
                        >
                          <X className="h-4 w-4" />
                        </Button>
                      </div>
                    ) : (
                      <div className="flex items-center gap-2">
                        <code className="bg-slate-100 px-2 py-1 rounded text-xs font-mono">
                          {m.canonical_name}
                        </code>
                        {m.confidence < 0.80 && (
                          <Button
                            variant="secondary"
                            size="sm"
                            onClick={() => handleToggleEdit(m.column_index)}
                          >
                            <Edit2 className="h-3 w-3" />
                          </Button>
                        )}
                      </div>
                    )}

                    {/* Confidence Badge */}
                    <Badge className={getConfidenceBadgeColor(m.confidence)}>
                      {Math.round(m.confidence * 100)}%
                    </Badge>
                  </div>
                ))}
              </div>

              {/* Help Text */}
              <p className="text-xs text-slate-600 p-2 bg-blue-50 rounded-md border border-blue-200">
                💡 <strong>Tip:</strong> Click the edit icon to reassign column
                mappings. When done, choose an action below.
              </p>
            </div>

            {/* Footer */}
            <div className="px-6 py-4 border-t border-slate-200 bg-slate-50 flex items-center justify-between gap-3 shrink-0">
              <div className="flex gap-2">
                <Button
                  variant={selectedMode === 'confirm' ? 'primary' : 'secondary'}
                  size="sm"
                  onClick={() => setSelectedMode('confirm')}
                >
                  <CheckCircle className="h-4 w-4 mr-1" />
                  Accept
                </Button>
                <Button
                  variant={selectedMode === 'retry' ? 'primary' : 'secondary'}
                  size="sm"
                  onClick={() => setSelectedMode('retry')}
                >
                  <RefreshCw className="h-4 w-4 mr-1" />
                  Re-run
                </Button>
              </div>
              <div className="flex gap-2 ml-auto">
                <Button variant="secondary" onClick={onClose}>
                  Cancel
                </Button>
                {selectedMode === 'confirm' ? (
                  <Button onClick={handleConfirm} className="gap-2">
                    <CheckCircle className="h-4 w-4" />
                    Accept Mappings
                  </Button>
                ) : (
                  <Button onClick={handleRetry} className="gap-2">
                    <RefreshCw className="h-4 w-4" />
                    Re-run Conversion
                  </Button>
                )}
              </div>
            </div>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
