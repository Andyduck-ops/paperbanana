import { useState, useCallback, useRef } from 'react';
import { streamGenerate } from '../lib/sse';
import type { StageStatus } from '../components/StageCard';
import type { GenerateRequest } from '../types/api';

export interface StageState {
  stage: string;
  agent: string;
  status: StageStatus;
  summary?: string;
  error?: string;
  artifactCount?: number;
  artifactKinds?: string[];
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

export function useGenerate() {
  const [state, setState] = useState<GenerateState>(createInitialState);
  const abortControllerRef = useRef<AbortController | null>(null);
  const sessionIdRef = useRef<string | null>(null);

  const cancel = useCallback(async () => {
    const sessionId = sessionIdRef.current;
    if (!sessionId) {
      return;
    }

    // Call the cancel API endpoint
    try {
      await fetch(`/api/v1/sessions/${sessionId}/cancel`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      });
    } catch {
      // Ignore network errors, abort locally anyway
    }

    // Abort the fetch request
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
    }

    // Update state to show canceled status
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
    // Abort any existing request
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    // Create new AbortController for this request
    const abortController = new AbortController();
    abortControllerRef.current = abortController;

    setState({
      isGenerating: true,
      isCanceled: false,
      stages: [],
      result: null,
      error: null,
      resumeMetadata: null,
      stagesNotRun: [],
    });

    // GD-UI-001: Full 5-stage pipeline order
    const stageOrder = ['retriever', 'planner', 'stylist', 'visualizer', 'critic'];
    const agentNames: Record<string, string> = {
      retriever: 'Retriever',
      planner: 'Planner',
      stylist: 'Stylist',
      visualizer: 'Visualizer',
      critic: 'Critic',
    };

    // Initialize stages
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
            setState((prev) => ({
              ...prev,
              stages: prev.stages.map((s) =>
                s.stage === data.stage ? { ...s, status: 'running' } : s
              ),
            }));
          },
          onStageComplete: (data) => {
            setState((prev) => ({
              ...prev,
              stages: prev.stages.map((s) =>
                s.stage === data.stage
                  ? {
                      ...s,
                      status: 'complete',
                      summary: data.summary,
                      artifactCount: data.artifact_count,
                      artifactKinds: data.artifact_kinds,
                    }
                  : s
              ),
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
            // GD-UI-002: Mark failed stage and stages_not_run
            setState((prev) => {
              const failedStageIndex = data.stage
                ? prev.stages.findIndex((s) => s.stage === data.stage)
                : -1;

              // Identify stages that will not run (after the failed stage)
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
                    // Mark the failed stage
                    return { ...s, status: 'error', error: data.message };
                  }
                  if (failedStageIndex >= 0) {
                    const currentIndex = prev.stages.findIndex((st) => st.stage === s.stage);
                    if (currentIndex > failedStageIndex) {
                      // Mark subsequent stages as not_run
                      return { ...s, status: 'not_run' as StageStatus };
                    }
                  }
                  return s;
                }),
              };
            });
          },
          onResumeStart: (data) => {
            // GD-UI-004: Handle resumed task - mark completed stages and set resume metadata
            sessionIdRef.current = data.session_id;
            setState((prev) => ({
              ...prev,
              resumeMetadata: {
                resumed_from_stage: data.resumed_from_stage,
                stages_completed_before_resume: data.stages_completed_before_resume,
              },
              stages: prev.stages.map((s) => {
                if (data.stages_completed_before_resume.includes(s.stage)) {
                  // Mark previously completed stages
                  return { ...s, status: 'complete' as StageStatus };
                }
                return s;
              }),
            }));
          },
        }
      );
    } catch (err) {
      // Check if this was an abort error
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

  const reset = useCallback(() => {
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
    });
  }, []);

  return {
    ...state,
    generate,
    cancel,
    reset,
    restore,
  };
}
