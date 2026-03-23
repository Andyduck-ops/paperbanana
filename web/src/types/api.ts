/**
 * API Types - 统一的 API 类型定义
 * 与后端 DTO 保持同步
 */

// ============================================
// Refine API - 精修相关类型
// ============================================

export interface RefineImageAsset {
  file: File;
  previewUrl: string;
}

export interface RefineRequest {
  session_id?: string;
  image: RefineImageAsset;
  instructions: string; // Refinement instructions
  resolution: "2K" | "4K";
  model?: string; // Optional model override
  provider_id?: string; // Optional provider ID
  max_iterations?: number; // Maximum refinement iterations (default: 3, max: 5)
  enable_iteration?: boolean; // Enable iterative refinement with critic
}

export interface RefineApiRequest {
  session_id?: string;
  image_data: string; // Base64 encoded image
  instructions: string; // Refinement instructions
  resolution: "2K" | "4K";
  model?: string; // Optional model override
  provider_id?: string; // Optional provider ID
  max_iterations?: number; // Maximum refinement iterations (default: 3, max: 5)
  enable_iteration?: boolean; // Enable iterative refinement with critic
}

export interface RefineResponseMetadata {
  iterations?: string;
  stop_reason?:
    | "accepted"
    | "max_iterations"
    | "no_critic"
    | "critic_failed"
    | "critic_init_failed";
  quality_score?: string;
}

export interface RefineImagePayload {
  data?: string;
  mime_type?: string;
  metadata?: Record<string, string>;
}

export interface RefineResponse {
  session_id: string;
  status: "completed" | "failed";
  image?: RefineImagePayload;
  image_data?: string; // Compatibility field for older callers
  content?: string; // Textual fallback or final model text content
  error?: string;
  metadata?: RefineResponseMetadata;
}

export interface RefineResult {
  sessionId: string;
  status: "completed" | "failed";
  image: {
    data: string;
    mimeType: string;
    metadata?: Record<string, string>;
  };
  content?: string;
  metadata?: RefineResponseMetadata;
}

// ============================================
// Generate API - 生成相关类型
// ============================================

export interface GenerateRequest {
  prompt: string;
  mode?: 'diagram' | 'plot';
  model?: string;
  session_id?: string;
  resume?: boolean;
  visualizer_node?: string;
  project_id?: string;
  folder_id?: string;
  num_candidates?: number;
  
  // Config fields
  aspect_ratio?: '21:9' | '16:9' | '3:2';
  critic_rounds?: number; // 1-5
  retrieval_mode?: 'auto' | 'manual' | 'random' | 'none';
  pipeline_mode?: 'full' | 'planner-critic' | 'vanilla';
  query_model?: string;
  gen_model?: string;
}

export interface GenerateResponse {
  session_id: string;
  status: 'completed' | 'failed';
  generated_artifacts?: Artifact[];
  error?: string;
}

// ============================================
// Batch API - 批次相关类型
// ============================================

export interface BatchGenerateRequest extends GenerateRequest {
  num_candidates: number; // Required for batch
}

export interface BatchGenerateResponse {
  batch_id: string;
  results: CandidateResult[];
  successful: number;
  failed: number;
  timing: BatchTiming;
}

export interface CandidateResult {
  candidate_id: number;
  session_id: string;
  status: 'completed' | 'failed';
  artifacts?: Artifact[];
  error?: ErrorDetail;
}

// ============================================
// Common Types - 通用类型
// ============================================

export interface Artifact {
  id: string;
  kind: ArtifactKind;
  mime_type: string;
  uri?: string;
  content?: string;
  bytes?: string; // Base64 encoded
  metadata?: Record<string, string>;
}

export type ArtifactKind =
  | 'reference_bundle'
  | 'plan'
  | 'rendered_figure'
  | 'prompt_trace'
  | 'critique'
  | 'polished_image';

export interface ErrorDetail {
  message: string;
  stage?: string;
  code?: string;
}

export interface BatchTiming {
  started_at: string;
  completed_at: string;
  duration: number; // milliseconds
}

// ============================================
// History API - 历史记录类型
// ============================================

export interface HistorySession {
  id: string;
  project_id: string;
  created_at: string;
  status: 'completed' | 'failed' | 'running';
  current_stage?: string;
  prompt?: string;
}

// ============================================
// Project API - 项目相关类型
// ============================================

export interface Project {
  id: string;
  name: string;
  created_at: string;
}

export interface Folder {
  id: string;
  name: string;
  type: 'folder' | 'visualization';
  created_at: string;
}
