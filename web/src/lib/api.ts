const API_BASE = '/api/v1';

export class ApiError extends Error {
  constructor(
    public status: number,
    public statusText: string,
    message: string
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const text = await response.text();
    throw new ApiError(
      response.status,
      response.statusText,
      text || `HTTP ${response.status}: ${response.statusText}`
    );
  }
  return response.json();
}

export const apiClient = {
  // Generation endpoints
  async generate(data: { prompt: string; visualizer_node?: string }) {
    const response = await fetch(`${API_BASE}/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse<{
      session_id: string;
      generated_artifacts: Array<{
        kind: string;
        mime_type: string;
        summary: string;
        data?: string;
      }>;
    }>(response);
  },

  // Project endpoints
  async listProjects() {
    const response = await fetch(`${API_BASE}/projects`);
    return handleResponse<{
      projects: Array<{ id: string; name: string; created_at: string }>;
    }>(response);
  },

  async createProject(name: string) {
    const response = await fetch(`${API_BASE}/projects`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name }),
    });
    return handleResponse<{ id: string; name: string; created_at: string }>(response);
  },

  // Folder contents
  async listFolderContents(projectId: string, folderId?: string) {
    const path = folderId
      ? `${API_BASE}/projects/${projectId}/folders/${folderId}/contents`
      : `${API_BASE}/projects/${projectId}/folders`;
    const response = await fetch(path);
    return handleResponse<{
      items: Array<{
        id: string;
        name: string;
        type: 'folder' | 'visualization';
        created_at: string;
      }>;
    }>(response);
  },

  // History
  async listHistory(projectId?: string) {
    const query = projectId ? `?project_id=${projectId}` : '';
    const response = await fetch(`${API_BASE}/history${query}`);
    return handleResponse<{
      sessions: Array<{
        id: string;
        project_id: string;
        created_at: string;
        status: string;
      }>;
    }>(response);
  },

  // Asset download (returns blob URL)
  async getAssetUrl(assetId: string): Promise<string> {
    return `${API_BASE}/assets/${assetId}`;
  },

  // Session restore
  async getSession(sessionId: string) {
    const response = await fetch(`${API_BASE}/session/${sessionId}`);
    return handleResponse<{
      id: string;
      project_id: string;
      visualization_id?: string;
      status: string;
      current_stage: string;
      schema_version: string;
      created_at: string;
      updated_at: string;
      completed_at?: string;
      snapshot?: {
        schema_version: string;
        session_id: string;
        request_id: string;
        status: string;
        current_stage: string;
        failed_stage?: string;
        pipeline: string[];
        initial_input: {
          session_id: string;
          request_id: string;
          stage: string;
          content: string;
          messages?: unknown[];
          visual_intent: {
            mode: string;
            goal: string;
            audience: string;
            style: string;
            constraints?: string[];
            preferred_outputs?: string[];
          };
          retrieved_references?: unknown[];
          prompt: {
            system_instruction: string;
            version: string;
            template: string;
            variables?: Record<string, string>;
          };
          generated_artifacts?: Array<{
            id: string;
            kind: string;
            mime_type: string;
            uri: string;
            content?: string;
            bytes?: string;
            metadata?: Record<string, string>;
          }>;
          critique_rounds?: unknown[];
          restore: {
            snapshot_version: string;
            restored_from: string;
            restored_at: string;
            resume_token: string;
          };
          metadata?: Record<string, string>;
        };
        stage_states?: Array<{
          stage: string;
          status: string;
          timing: {
            started_at: string;
            completed_at: string;
            duration: number;
          };
          input: unknown;
          output: {
            stage: string;
            content?: string;
            messages?: unknown[];
            visual_intent: unknown;
            retrieved_references?: unknown[];
            prompt: unknown;
            generated_artifacts?: Array<{
              id: string;
              kind: string;
              mime_type: string;
              uri: string;
              content?: string;
              bytes?: string;
              metadata?: Record<string, string>;
            }>;
            critique_rounds?: unknown[];
            error?: {
              message: string;
              code?: string;
              category?: string;
              retryable: boolean;
              suggestion?: string;
              stage?: string;
            };
            metadata?: Record<string, string>;
          };
          error?: {
            message: string;
            code?: string;
            category?: string;
            retryable: boolean;
            suggestion?: string;
            stage?: string;
          };
          restore: unknown;
        }>;
        final_output: {
          stage: string;
          content?: string;
          messages?: unknown[];
          visual_intent: unknown;
          retrieved_references?: unknown[];
          prompt: unknown;
          generated_artifacts?: Array<{
            id: string;
            kind: string;
            mime_type: string;
            uri: string;
            content?: string;
            bytes?: string;
            metadata?: Record<string, string>;
          }>;
          critique_rounds?: unknown[];
          error?: unknown;
          metadata?: Record<string, string>;
        };
        error?: {
          message: string;
          code?: string;
          category?: string;
          retryable: boolean;
          suggestion?: string;
          stage?: string;
        };
        restore: {
          snapshot_version: string;
          restored_from: string;
          restored_at: string;
          resume_token: string;
        };
        metadata?: Record<string, string>;
        started_at: string;
        updated_at: string;
        completed_at: string;
      };
    }>(response);
  },
};
