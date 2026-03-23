import { useState, useCallback } from 'react';
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

export interface GenerateResult {
  sessionId: string;
  artifacts: Array<{
    kind: string;
    mimeType: string;
    summary: string;
    data?: string;
    assetId?: string;
  }>;
}

export interface GenerateState {
  isGenerating: boolean;
  stages: StageState[];
  result: GenerateResult | null;
  error: string | null;
}

interface GenerateOptions {
  visualizerNode?: string;
  config?: Pick<
    GenerateRequest,
    'aspect_ratio' | 'critic_rounds' | 'retrieval_mode' | 'pipeline_mode' | 'query_model' | 'gen_model'
  >;
}

export function useGenerate() {
  const [state, setState] = useState<GenerateState>({
    isGenerating: false,
    stages: [],
    result: null,
    error: null,
  });

  const generate = useCallback(async (
    prompt: string,
    options?: GenerateOptions
  ) => {
    setState({
      isGenerating: true,
      stages: [],
      result: null,
      error: null,
    });

    const stageOrder = ['retriever', 'planner', 'visualizer', 'critic'];
    const agentNames: Record<string, string> = {
      retriever: 'Retriever',
      planner: 'Planner',
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
          visualizer_node: options?.visualizerNode,
          aspect_ratio: options?.config?.aspect_ratio,
          critic_rounds: options?.config?.critic_rounds,
          retrieval_mode: options?.config?.retrieval_mode,
          pipeline_mode: options?.config?.pipeline_mode,
          query_model: options?.config?.query_model,
          gen_model: options?.config?.gen_model,
        },
        {
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
                  assetId: (a as { asset_id?: string }).asset_id,
                })),
              },
            }));
          },
          onError: (data) => {
            setState((prev) => ({
              ...prev,
              isGenerating: false,
              error: data.message,
              stages: data.stage
                ? prev.stages.map((s) =>
                    s.stage === data.stage
                      ? { ...s, status: 'error', error: data.message }
                      : s
                  )
                : prev.stages,
            }));
          },
        }
      );
    } catch (err) {
      setState((prev) => ({
        ...prev,
        isGenerating: false,
        error: err instanceof Error ? err.message : 'Unknown error',
      }));
    }
  }, []);

  const reset = useCallback(() => {
    setState({
      isGenerating: false,
      stages: [],
      result: null,
      error: null,
    });
  }, []);

  return {
    ...state,
    generate,
    reset,
  };
}
