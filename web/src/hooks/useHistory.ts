import { useState, useEffect, useCallback } from 'react';
import { apiClient } from '../lib/api';
import type { StageStatus } from '../components/StageCard';

export interface HistorySession {
  id: string;
  projectId: string;
  createdAt: string;
  status: string;
  prompt?: string;
  thumbnailUrl?: string;
  summary?: string;
  mode?: 'generate' | 'batch' | 'refine' | 'workspace';
  source?: 'local' | 'server';
}

export interface HistoryState {
  sessions: HistorySession[];
  isLoading: boolean;
  error: string | null;
}

export interface RestoredSession {
  id: string;
  status: string;
  mode: 'generate' | 'batch' | 'refine';
  prompt: string;
  artifacts: Array<{
    kind: string;
    mimeType: string;
    data?: string;
    assetId?: string;
    projectId?: string;
    uri?: string;
    summary?: string;
  }>;
  stages?: Array<{
    stage: string;
    status: StageStatus;
    summary?: string;
    error?: string;
    artifactCount?: number;
    artifactKinds?: string[];
  }>;
  error?: string;
  resumeMetadata?: {
    resumed_from_stage: string;
    stages_completed_before_resume: string[];
  };
  // Batch-specific fields
  batchId?: string;
  candidates?: Array<{
    candidateId: number;
    sessionId: string;
    status: string;
    artifacts: Array<{
      kind: string;
      mimeType: string;
      data?: string;
      assetId?: string;
      projectId?: string;
      summary?: string;
    }>;
    error?: string;
  }>;
  successful?: number;
  failed?: number;
  startedAt?: string;
  completedAt?: string;
}

export function useHistory(projectId?: string) {
  const [state, setState] = useState<HistoryState>({
    sessions: [],
    isLoading: false,
    error: null,
  });

  const fetchHistory = useCallback(async () => {
    setState((prev) => ({ ...prev, isLoading: true, error: null }));
    try {
      const response = await apiClient.listHistory(projectId);
      setState({
        sessions: response.sessions.map((s) => ({
          id: s.id,
          projectId: s.project_id,
          createdAt: s.created_at,
          status: s.status,
        })),
        isLoading: false,
        error: null,
      });
    } catch (err) {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error: err instanceof Error ? err.message : 'Failed to load history',
      }));
    }
  }, [projectId]);

  useEffect(() => {
    fetchHistory();
  }, [fetchHistory]);

  // Helper function to map backend status to StageStatus
  function mapStatusToStageStatus(status: string): StageStatus {
    switch (status) {
      case 'completed':
        return 'complete';
      case 'failed':
        return 'error';
      case 'running':
        return 'running';
      case 'not_run':
        return 'not_run';
      default:
        return 'pending';
    }
  }

  const restoreSession = useCallback(async (sessionId: string): Promise<RestoredSession | null> => {
    try {
      const response = await apiClient.getSession(sessionId);
      const snapshot = response.snapshot;

      if (!snapshot) {
        return null;
      }

      // Determine mode from snapshot metadata
      const mode = snapshot.metadata?.['history.mode'] ||
        (snapshot.current_stage === 'polish' ? 'refine' :
          (snapshot.metadata?.['batch.group_id'] ? 'batch' : 'generate'));

      // Extract artifacts from stage states or final output
      const artifacts: RestoredSession['artifacts'] = [];
      const stages: RestoredSession['stages'] = [];

      if (snapshot.stage_states) {
        for (const stageState of snapshot.stage_states) {
          stages.push({
            stage: stageState.stage,
            status: mapStatusToStageStatus(stageState.status),
            summary: stageState.output?.content?.slice(0, 100),
            error: stageState.error?.message,
            artifactCount: stageState.output?.generated_artifacts?.length || 0,
            artifactKinds: stageState.output?.generated_artifacts?.map((a) => a.kind),
          });

          if (stageState.output?.generated_artifacts) {
            for (const artifact of stageState.output.generated_artifacts) {
              artifacts.push({
                kind: artifact.kind,
                mimeType: artifact.mime_type,
                data: artifact.data || artifact.content,
                assetId: artifact.asset_id,
                projectId: artifact.metadata?.project_id,
                uri: artifact.uri,
              });
            }
          }
        }
      }

      // Also check final output for artifacts
      if (snapshot.final_output?.generated_artifacts) {
        for (const artifact of snapshot.final_output.generated_artifacts) {
          // Avoid duplicates
          if (!artifacts.find((a) => a.assetId === artifact.asset_id)) {
            artifacts.push({
              kind: artifact.kind,
              mimeType: artifact.mime_type,
              data: artifact.data || artifact.content,
              assetId: artifact.asset_id,
              projectId: artifact.metadata?.project_id,
              uri: artifact.uri,
            });
          }
        }
      }

      // Extract prompt from initial input
      const prompt = snapshot.initial_input?.content ||
        snapshot.initial_input?.visual_intent?.goal ||
        snapshot.initial_input?.metadata?.['http.visual_intent'] ||
        snapshot.metadata?.['http.prompt'] || '';

      // Handle batch mode
      if (mode === 'batch') {
        const batchGroupId = snapshot.metadata?.['batch.group_id'] || response.id;
        const sessionIdsRaw = snapshot.metadata?.['batch.session_ids'] || '';
        const sessionIds = sessionIdsRaw.split(',').map((s) => s.trim()).filter(Boolean);

        // For batch, we need to construct candidates from the session_ids
        // Since we only have one session's snapshot, we treat the current one as representative
        const candidates: RestoredSession['candidates'] = sessionIds.map((sid, index) => ({
          candidateId: index + 1,
          sessionId: sid,
          status: response.status === 'completed' ? 'completed' : 'failed',
          artifacts: artifacts.slice(0, 1), // Use first artifact as representative
        }));

        return {
          id: response.id,
          status: response.status,
          mode: 'batch',
          prompt,
          artifacts,
          stages,
          error: snapshot.error?.message,
          batchId: batchGroupId,
          candidates,
          successful: candidates.filter((c) => c.status === 'completed').length,
          failed: candidates.filter((c) => c.status === 'failed').length,
          startedAt: snapshot.started_at,
          completedAt: snapshot.completed_at,
        };
      }

      return {
        id: response.id,
        status: response.status,
        mode: mode as 'generate' | 'refine',
        prompt,
        artifacts,
        stages,
        error: snapshot.error?.message,
        resumeMetadata: snapshot.restore ? {
          resumed_from_stage: snapshot.restore.restored_from,
          stages_completed_before_resume: snapshot.pipeline?.filter((_, idx) =>
            snapshot.stage_states?.slice(0, idx).some(s => s.status === 'completed')
          ) || [],
        } : undefined,
      };
    } catch (err) {
      console.error('Failed to restore session:', err);
      return null;
    }
  }, []);

  return {
    ...state,
    count: state.sessions.length,
    refresh: fetchHistory,
    restoreSession,
  };
}
