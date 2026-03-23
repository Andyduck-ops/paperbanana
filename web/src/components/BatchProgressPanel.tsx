import { useState, useEffect, useCallback } from 'react';
import { useLanguage } from '../hooks';

export interface BatchCandidate {
  candidateId: number;
  status: 'pending' | 'running' | 'completed' | 'failed';
  artifacts?: Array<{
    id: string;
    kind: string;
    mimeType: string;
  }>;
  error?: string;
}

export interface BatchProgress {
  batchId: string;
  status: 'running' | 'completed' | 'failed';
  candidates: BatchCandidate[];
  successful: number;
  failed: number;
  startedAt: string;
  completedAt?: string;
}

export interface BatchProgressPanelProps {
  batchId: string | null;
  apiBase?: string;
  initialProgress?: BatchProgress;
  onComplete?: (result: BatchProgress) => void;
}

export function BatchProgressPanel({ batchId, apiBase = '', initialProgress, onComplete }: BatchProgressPanelProps) {
  const { t } = useLanguage();
  const [progress, setProgress] = useState<BatchProgress | null>(initialProgress || null);
  const [isDownloading, setIsDownloading] = useState(false);

  useEffect(() => {
    // If initialProgress is provided, use it directly
    if (initialProgress) {
      setProgress(initialProgress);
      return;
    }

    if (!batchId) {
      setProgress(null);
      return;
    }

    // Initialize progress
    setProgress({
      batchId,
      status: 'running',
      candidates: [],
      successful: 0,
      failed: 0,
      startedAt: new Date().toISOString(),
    });

    // SSE connection would be handled by the parent component
    // This component just displays the progress
  }, [batchId, initialProgress]);

  // Call onComplete when batch finishes
  useEffect(() => {
    if (progress?.status === 'completed' && onComplete) {
      onComplete(progress);
    }
  }, [progress?.status, onComplete, progress]);

  const handleDownloadZip = async () => {
    if (!progress?.batchId) return;

    setIsDownloading(true);
    try {
      const response = await fetch(`${apiBase}/api/v1/batch/download`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ batch_id: progress.batchId }),
      });

      if (!response.ok) throw new Error('Download failed');

      // Get the blob and trigger download
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `paperviz_candidates_${Date.now()}.zip`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (error) {
      console.error('Download failed:', error);
    } finally {
      setIsDownloading(false);
    }
  };

  if (!progress) return null;

  const totalCandidates = progress.candidates.length || 0;
  const completedCount = progress.successful + progress.failed;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-medium text-foreground">{t('generate.batchProgress')}</h3>
        <span className="text-sm text-muted-foreground">
          {completedCount} / {totalCandidates}
        </span>
      </div>

      {/* Overall progress bar */}
      <div className="w-full h-2 bg-muted rounded-full overflow-hidden">
        <div
          className="h-full bg-primary transition-all duration-300"
          style={{
            width: totalCandidates > 0
              ? `${(completedCount / totalCandidates) * 100}%`
              : '0%'
          }}
        />
      </div>

      {/* Candidate cards */}
      <div className="grid gap-2">
        {progress.candidates.map((candidate) => (
          <div
            key={candidate.candidateId}
            className={`p-3 rounded-lg border ${
              candidate.status === 'completed'
                ? 'border-green-500/50 bg-green-500/10'
                : candidate.status === 'failed'
                ? 'border-red-500/50 bg-red-500/10'
                : candidate.status === 'running'
                ? 'border-primary/50 bg-primary/10'
                : 'border-border bg-background'
            }`}
          >
            <div className="flex items-center justify-between">
              <span className="font-medium text-foreground">
                {t('generate.candidate', { number: candidate.candidateId + 1 })}
              </span>
              <span
                className={`text-sm ${
                  candidate.status === 'completed'
                    ? 'text-green-600'
                    : candidate.status === 'failed'
                    ? 'text-red-600'
                    : 'text-muted-foreground'
                }`}
              >
                {candidate.status === 'completed' && `✓ ${t('generate.successful')}`}
                {candidate.status === 'failed' && `✗ ${t('generate.failed')}`}
                {candidate.status === 'running' && '◆'}
                {candidate.status === 'pending' && '○'}
              </span>
            </div>
            {candidate.error && (
              <p className="mt-1 text-sm text-red-600">{candidate.error}</p>
            )}
          </div>
        ))}
      </div>

      {/* Summary when complete */}
      {progress.status === 'completed' && (
        <div className="p-4 rounded-lg bg-muted">
          <h4 className="font-medium text-foreground">{t('generate.batchComplete')}</h4>
          <p className="text-sm text-muted-foreground mt-1">
            {progress.successful} {t('generate.successful')}, {progress.failed} {t('generate.failed')}
          </p>
        </div>
      )}

      {/* Download button when complete with successful candidates */}
      {progress.status === 'completed' && progress.successful > 0 && (
        <div className="mt-4 flex justify-center">
          <button
            onClick={handleDownloadZip}
            disabled={isDownloading}
            className="px-6 py-2 rounded-lg bg-primary text-background font-medium hover:opacity-90 disabled:opacity-50"
          >
            {isDownloading ? t('generate.downloading') : t('generate.downloadAll')}
          </button>
        </div>
      )}
    </div>
  );
}

// Hook for SSE batch event handling
export function useBatchSSE(apiBase: string) {
  const [progress, setProgress] = useState<BatchProgress | null>(null);
  const [isGenerating, setIsGenerating] = useState(false);

  const startBatch = useCallback(async (prompt: string, numCandidates: number, options?: {
    mode?: string;
    model?: string;
    visualizerNode?: string;
  }) => {
    setIsGenerating(true);
    setProgress(null);

    try {
      const response = await fetch(`${apiBase}/api/v1/generate/batch`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          prompt,
          num_candidates: numCandidates,
          mode: options?.mode,
          model: options?.model,
          visualizer_node: options?.visualizerNode,
        }),
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const reader = response.body?.getReader();
      if (!reader) throw new Error('No response body');

      const decoder = new TextDecoder();
      let buffer = '';
      const candidates: BatchCandidate[] = [];

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';

        for (const line of lines) {
          if (line.startsWith('event:')) {
            // Event type line - we can ignore or use for routing
          } else if (line.startsWith('data:')) {
            try {
              const data = JSON.parse(line.slice(5).trim());

              // Handle different event types
              if (data.type === 'batch_start') {
                candidates.length = 0;
                for (let i = 0; i < numCandidates; i++) {
                  candidates.push({ candidateId: i, status: 'pending' });
                }
                setProgress({
                  batchId: data.batch_id,
                  status: 'running',
                  candidates: [...candidates],
                  successful: 0,
                  failed: 0,
                  startedAt: data.timing?.started_at || new Date().toISOString(),
                });
              } else if (data.type === 'candidate_start') {
                const idx = data.candidate_id;
                if (candidates[idx]) {
                  candidates[idx].status = 'running';
                }
                setProgress(prev => prev ? { ...prev, candidates: [...candidates] } : null);
              } else if (data.type === 'candidate_complete') {
                const idx = data.candidate_id;
                if (candidates[idx]) {
                  candidates[idx].status = data.status === 'completed' ? 'completed' : 'failed';
                  if (data.error) {
                    candidates[idx].error = data.error.message;
                  }
                }
                setProgress(prev => prev ? {
                  ...prev,
                  candidates: [...candidates],
                  successful: candidates.filter(c => c.status === 'completed').length,
                  failed: candidates.filter(c => c.status === 'failed').length,
                } : null);
              } else if (data.type === 'batch_complete') {
                setProgress(prev => prev ? {
                  ...prev,
                  status: 'completed',
                  completedAt: data.timing?.completed_at || new Date().toISOString(),
                } : null);
              } else if (data.type === 'batch_result') {
                // Final result with all artifacts
                const results = data.results || [];
                results.forEach((result: any) => {
                  const idx = result.candidate_id;
                  if (candidates[idx]) {
                    candidates[idx].artifacts = result.artifacts;
                  }
                });
                setProgress(prev => prev ? { ...prev, candidates: [...candidates] } : null);
              }
            } catch {
              // Skip unparseable lines
            }
          }
        }
      }
    } catch (error) {
      console.error('Batch generation error:', error);
      setProgress(prev => prev ? { ...prev, status: 'failed' } : null);
    } finally {
      setIsGenerating(false);
    }
  }, [apiBase]);

  return { progress, isGenerating, startBatch };
}
