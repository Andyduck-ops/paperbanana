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
  projectId: string;
  visualizationId?: string;
  status: string;
  currentStage?: string;
  mode: 'generate' | 'batch' | 'refine';
  prompt: string;
  artifacts: Array<{
    kind: string;
    mimeType: string;
    summary?: string;
    data?: string;
    assetId?: string;
    projectId?: string;
    uri?: string;
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
      summary?: string;
      data?: string;
      assetId?: string;
      projectId?: string;
      uri?: string;
    }>;
    stages?: Array<{
      stage: string;
      status: StageStatus;
      summary?: string;
      error?: string;
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
        sessions: (response.sessions || []).map((s) => ({
          id: s.id,
          projectId: s.project_id,
          createdAt: s.created_at,
          status: s.status,
          prompt: (s as { prompt?: string }).prompt,
          summary: (s as { summary?: string }).summary,
          mode: (s as { mode?: 'generate' | 'batch' | 'refine' | 'workspace' }).mode,
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
        (response.current_stage === 'polish' ? 'refine' :
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
                summary: artifact.metadata?.summary,
                data: artifact.bytes || artifact.content,
                assetId: artifact.id,
                projectId: artifact.metadata?.project_id,
                uri: artifact.uri,
              });
            }
          }
        }
      }

      // Also check final output for artifacts (used for refine mode)
      if (snapshot.final_output?.generated_artifacts) {
        for (const artifact of snapshot.final_output.generated_artifacts) {
          // Avoid duplicates
          if (!artifacts.find((a) => a.assetId === artifact.id)) {
            artifacts.push({
              kind: artifact.kind,
              mimeType: artifact.mime_type,
              summary: artifact.metadata?.summary,
              data: artifact.bytes || artifact.content,
              assetId: artifact.id,
              projectId: artifact.metadata?.project_id,
              uri: artifact.uri,
            });
          }
        }
      }

      // Extract prompt from initial input
      const prompt = snapshot.initial_input?.content ||
        snapshot.initial_input?.visual_intent?.goal ||
        snapshot.metadata?.['http.prompt'] || '';

      // Handle batch mode
      if (mode === 'batch') {
        // Check if this is a batch root session with candidate session IDs
        const batchGroupId = snapshot.metadata?.['batch.group_id'] || response.id;
        const sessionIdsRaw = snapshot.metadata?.['batch.session_ids'] || '';
        const candidateSessionIds = sessionIdsRaw.split(',').map((s) => s.trim()).filter(Boolean);

        // If we have candidate session IDs, fetch each candidate
        const candidates: RestoredSession['candidates'] = [];

        if (candidateSessionIds.length > 0) {
          for (let i = 0; i < candidateSessionIds.length; i++) {
            const candidateSessionId = candidateSessionIds[i];
            try {
              const candidateResponse = await apiClient.getSession(candidateSessionId);
              const candidateSnapshot = candidateResponse.snapshot;

              const candidateArtifacts: RestoredSession['artifacts'] = [];
              const candidateStages: RestoredSession['stages'] = [];

              if (candidateSnapshot?.stage_states) {
                for (const stageState of candidateSnapshot.stage_states) {
                  candidateStages.push({
                    stage: stageState.stage,
                    status: mapStatusToStageStatus(stageState.status),
                    summary: stageState.output?.content?.slice(0, 100),
                    error: stageState.error?.message,
                  });

                  if (stageState.output?.generated_artifacts) {
                    for (const artifact of stageState.output.generated_artifacts) {
                      candidateArtifacts.push({
                        kind: artifact.kind,
                        mimeType: artifact.mime_type,
                        summary: artifact.metadata?.summary,
                        data: artifact.bytes || artifact.content,
                        assetId: artifact.id,
                        projectId: artifact.metadata?.project_id,
                      });
                    }
                  }
                }
              }

              if (candidateSnapshot?.final_output?.generated_artifacts) {
                for (const artifact of candidateSnapshot.final_output.generated_artifacts) {
                  if (!candidateArtifacts.find((a) => a.assetId === artifact.id)) {
                    candidateArtifacts.push({
                      kind: artifact.kind,
                      mimeType: artifact.mime_type,
                      summary: artifact.metadata?.summary,
                      data: artifact.bytes || artifact.content,
                      assetId: artifact.id,
                      projectId: artifact.metadata?.project_id,
                    });
                  }
                }
              }

              candidates.push({
                candidateId: i,
                sessionId: candidateSessionId,
                status: candidateResponse.status,
                artifacts: candidateArtifacts,
                stages: candidateStages,
                error: candidateSnapshot?.error?.message,
              });
            } catch {
              // If fetching a candidate fails, add it as failed
              candidates.push({
                candidateId: i,
                sessionId: candidateSessionId,
                status: 'failed',
                artifacts: [],
                error: 'Failed to fetch candidate session',
              });
            }
          }
        } else {
          // Fallback: treat current session artifacts as the only candidate
          candidates.push({
            candidateId: 0,
            sessionId: response.id,
            status: response.status,
            artifacts: artifacts.slice(0, 1),
            stages: stages,
          });
        }

        return {
          id: response.id,
          projectId: response.project_id,
          visualizationId: response.visualization_id,
          status: response.status,
          currentStage: response.current_stage,
          mode: 'batch',
          prompt,
          artifacts,
          stages,
          error: snapshot.error?.message,
          batchId: batchGroupId,
          candidates,
          successful: candidates.filter((c) => c.status === 'completed').length,
          failed: candidates.filter((c) => c.status === 'failed').length,
          startedAt: snapshot.started_at || response.created_at,
          completedAt: snapshot.completed_at || response.completed_at,
        };
      }

      return {
        id: response.id,
        projectId: response.project_id,
        visualizationId: response.visualization_id,
        status: response.status,
        currentStage: response.current_stage,
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
