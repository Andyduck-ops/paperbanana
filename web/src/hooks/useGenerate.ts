import { useState, useCallback, useRef, useEffect } from 'react';
import { streamGenerate } from '../lib/sse';
import { apiClient } from '../lib/api';
import type { StageStatus } from '../components/StageCard';
import type { GenerateRequest } from '../types/api';
import { useNetworkStatus } from './useNetworkStatus';

export interface StageState {
  stage: string;
  agent: string;
  status: StageStatus;
  summary?: string;
  error?: string;
  artifactCount?: number;
  artifactKinds?: string[];
  startedAt?: string;
  completedAt?: string;
  duration?: number;
}

export interface ResumeMetadata {
  resumed_from_stage: string;
  stages_completed_before_resume: string[];
}

export interface GenerateResult {
  sessionId: string;
  artifacts: Array<{
    kind: string;
    mimeType: string;
    summary: string;
    data?: string;
    assetId?: string;
    projectId?: string;
  }>;
}

export interface GenerateState {
  isGenerating: boolean;
  isCanceled: boolean;
  stages: StageState[];
  result: GenerateResult | null;
  error: string | null;
  resumeMetadata: ResumeMetadata | null;
  stagesNotRun: string[];
  isRetrying: boolean;
}

export interface RestoreGenerateSnapshot {
  sessionId: string;
  artifacts: GenerateResult['artifacts'];
  stages: Array<{
    stage: string;
    status: StageStatus;
    summary?: string;
    error?: string;
    artifactCount?: number;
    artifactKinds?: string[];
    startedAt?: string;
    completedAt?: string;
    duration?: number;
  }>;
  error?: string | null;
  resumeMetadata?: ResumeMetadata | null;
}

interface GenerateOptions {
  visualizerNode?: string;
  content?: string;
  visualIntent?: string;
  config?: Pick<
    GenerateRequest,
    'aspect_ratio' | 'critic_rounds' | 'retrieval_mode' | 'pipeline_mode' | 'query_model' | 'gen_model'
  >;
}

function createInitialState(): GenerateState {
  return {
    isGenerating: false,
    isCanceled: false,
    stages: [],
    result: null,
    error: null,
    resumeMetadata: null,
    stagesNotRun: [],
    isRetrying: false,
  };
}

function getAgentLabel(stage: string) {
  const agentNames: Record<string, string> = {
    retriever: 'Retriever',
    planner: 'Planner',
    stylist: 'Stylist',
    visualizer: 'Visualizer',
    critic: 'Critic',
    polish: 'Polish',
  };

  return agentNames[stage] || stage;
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`;
  if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
  return `${Math.floor(ms / 60000)}m ${Math.floor((ms % 60000) / 1000)}s`;
}

export function useGenerate() {
  const [state, setState] = useState<GenerateState>(createInitialState);
  const abortControllerRef = useRef<AbortController | null>(null);
  const sessionIdRef = useRef<string | null>(null);
  const lastPromptRef = useRef<string | null>(null);
  const lastOptionsRef = useRef<GenerateOptions | null>(null);
  const generateRef = useRef<((prompt: string, options?: GenerateOptions) => Promise<void>) | null>(null);
  const { online, onReconnect } = useNetworkStatus();

  const cancel = useCallback(async () => {
    const sessionId = sessionIdRef.current;
    if (!sessionId) {
      return;
    }

    try {
      await fetch(`/api/v1/sessions/${sessionId}/cancel`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      });
    } catch {
      // Ignore network errors, abort locally anyway
    }

    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }

    setState((prev) => ({
      ...prev,
      isGenerating: false,
      isCanceled: true,
      error: 'Generation canceled by user',
    }));
  }, []);

  const generate = useCallback(async (
    prompt: string,
    options?: GenerateOptions
  ) => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    const abortController = new AbortController();
    abortControllerRef.current = abortController;
    lastPromptRef.current = prompt;
    lastOptionsRef.current = options || null;

    setState({
      isGenerating: true,
      isCanceled: false,
      stages: [],
      result: null,
      error: null,
      resumeMetadata: null,
      stagesNotRun: [],
      isRetrying: false,
    });

    const stageOrder = ['retriever', 'planner', 'stylist', 'visualizer', 'critic'];
    const agentNames: Record<string, string> = {
      retriever: 'Retriever',
      planner: 'Planner',
      stylist: 'Stylist',
      visualizer: 'Visualizer',
      critic: 'Critic',
    };

    const initialStages: StageState[] = stageOrder.map((stage) => ({
      stage,
      agent: agentNames[stage] || stage,
      status: 'pending' as StageStatus,
    }));

    setState((prev) => ({ ...prev, stages: initialStages }));

    try {
      await streamGenerate(
        {
          prompt,
          content: options?.content,
          visual_intent: options?.visualIntent,
          visualizer_node: options?.visualizerNode,
          aspect_ratio: options?.config?.aspect_ratio,
          critic_rounds: options?.config?.critic_rounds,
          retrieval_mode: options?.config?.retrieval_mode,
          pipeline_mode: options?.config?.pipeline_mode,
          query_model: options?.config?.query_model,
          gen_model: options?.config?.gen_model,
        },
        {
          signal: abortController.signal,
          onStageStart: (data) => {
            const startedAt = data.occurred_at || new Date().toISOString();
            setState((prev) => ({
              ...prev,
              stages: prev.stages.map((s) =>
                s.stage === data.stage ? { ...s, status: 'running', startedAt } : s
              ),
            }));
          },
          onStageComplete: (data) => {
            const completedAt = data.occurred_at || new Date().toISOString();
            setState((prev) => ({
              ...prev,
              stages: prev.stages.map((s) => {
                if (s.stage !== data.stage) return s;
                const startedAt = s.startedAt ? new Date(s.startedAt).getTime() : Date.now();
                const endTime = new Date(completedAt).getTime();
                const duration = endTime - startedAt;
                return {
                  ...s,
                  status: 'complete',
                  summary: data.metadata?.summary,
                  artifactCount: data.metadata?.artifact_count ? parseInt(data.metadata.artifact_count, 10) : undefined,
                  artifactKinds: data.metadata?.artifact_kinds?.split(','),
                  completedAt,
                  duration,
                };
              }),
            }));
          },
          onResult: (data) => {
            sessionIdRef.current = data.session_id;
            setState((prev) => ({
              ...prev,
              isGenerating: false,
              result: {
                sessionId: data.session_id,
                artifacts: data.generated_artifacts.map((a) => ({
                  kind: a.kind,
                  mimeType: a.mime_type,
                  summary: a.summary,
                  data: a.data,
                  assetId: a.asset_id,
                  projectId: a.project_id || data.project_id,
                })),
              },
            }));
          },
          onError: (data) => {
            setState((prev) => {
              const failedStageIndex = data.stage
                ? prev.stages.findIndex((s) => s.stage === data.stage)
                : -1;

              const stagesNotRun = failedStageIndex >= 0
                ? prev.stages.slice(failedStageIndex + 1).map((s) => s.stage)
                : [];

              return {
                ...prev,
                isGenerating: false,
                error: data.message,
                stagesNotRun,
                stages: prev.stages.map((s) => {
                  if (data.stage && s.stage === data.stage) {
                    return { ...s, status: 'error', error: data.message };
                  }
                  if (failedStageIndex >= 0) {
                    const currentIndex = prev.stages.findIndex((st) => st.stage === s.stage);
                    if (currentIndex > failedStageIndex) {
                      return { ...s, status: 'not_run' as StageStatus };
                    }
                  }
                  return s;
                }),
              };
            });
          },
          onResumeStart: (data) => {
            sessionIdRef.current = data.session_id;
            setState((prev) => ({
              ...prev,
              resumeMetadata: {
                resumed_from_stage: data.resumed_from_stage,
                stages_completed_before_resume: data.stages_completed_before_resume,
              },
              stages: prev.stages.map((s) => {
                if (data.stages_completed_before_resume.includes(s.stage)) {
                  return { ...s, status: 'complete' as StageStatus };
                }
                return s;
              }),
            }));
          },
        }
      );
    } catch (err) {
      if (err instanceof Error && err.name === 'AbortError') {
        setState((prev) => ({
          ...prev,
          isGenerating: false,
          isCanceled: true,
          error: 'Generation canceled by user',
        }));
        return;
      }
      setState((prev) => ({
        ...prev,
        isGenerating: false,
        error: err instanceof Error ? err.message : 'Unknown error',
      }));
    }
  }, []);

  generateRef.current = generate;

  useEffect(() => {
    if (!online) return;

    return onReconnect(() => {
      if (state.isGenerating && lastPromptRef.current && generateRef.current) {
        const lastFailedStage = state.stages.find(s => s.status === 'error' || s.status === 'running');
        if (lastFailedStage && sessionIdRef.current) {
          apiClient.retrySession(sessionIdRef.current)
            .then((res) => {
              if (res.session_id && lastPromptRef.current && generateRef.current) {
                generateRef.current(lastPromptRef.current, lastOptionsRef.current || undefined);
              }
            })
            .catch(() => {
              // Silently fail retry
            });
        }
      }
    });
  }, [online, onReconnect, state.isGenerating, state.stages]);

  const reset = useCallback(() => {
    lastPromptRef.current = null;
    lastOptionsRef.current = null;
    abortControllerRef.current = null;
    setState(createInitialState());
  }, []);

  const restore = useCallback((snapshot: RestoreGenerateSnapshot) => {
    const stages = snapshot.stages.map((stage) => ({
      agent: getAgentLabel(stage.stage),
      ...stage,
    }));

    sessionIdRef.current = snapshot.sessionId;

    setState({
      isGenerating: false,
      isCanceled: false,
      stages,
      result: snapshot.artifacts.length > 0
        ? {
            sessionId: snapshot.sessionId,
            artifacts: snapshot.artifacts,
          }
        : null,
      error: snapshot.error || null,
      resumeMetadata: snapshot.resumeMetadata || null,
      stagesNotRun: stages
        .filter((stage) => stage.status === 'not_run')
        .map((stage) => stage.stage),
      isRetrying: false,
    });
  }, []);

  return {
    ...state,
    generate,
    cancel,
    reset,
    restore,
    formatDuration,
  };
}
