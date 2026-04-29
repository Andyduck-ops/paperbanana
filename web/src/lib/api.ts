export function getApiBase(): string {
  if (typeof window !== 'undefined' && (window as any).__TAURI_INTERNALS__) {
    const port = (window as any).__PAPERBANANA_BACKEND_PORT__;
    if (port) {
      return `http://127.0.0.1:${port}/api/v1`;
    }
  }
  return '/api/v1';
}

const API_BASE = getApiBase();

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

  async createProject(name: string, description?: string) {
    const response = await fetch(`${API_BASE}/projects`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, description }),
    });
    return handleResponse<{ id: string; name: string; description?: string; created_at: string }>(response);
  },

  async deleteProject(projectId: string) {
    const response = await fetch(`${API_BASE}/projects/${projectId}`, {
      method: 'DELETE',
    });
    if (!response.ok) {
      const text = await response.text();
      throw new ApiError(
        response.status,
        response.statusText,
        text || `HTTP ${response.status}: ${response.statusText}`
      );
    }
    return { success: true };
  },

  // Folder contents
  async listFolderContents(projectId: string, folderId?: string) {
    const params = new URLSearchParams();
    if (projectId) params.set('project_id', projectId);
    if (folderId) params.set('folder_id', folderId);
    const response = await fetch(`${API_BASE}/folders/contents?${params.toString()}`);
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

  async retrySession(sessionId: string) {
    const response = await fetch(`${API_BASE}/sessions/${sessionId}/retry`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    });
    return handleResponse<{
      session_id: string;
      status: string;
      resumed_from_stage?: string;
    }>(response);
  },

  // Folder CRUD endpoints
  async createFolder(data: { name: string; project_id: string; parent_id?: string }) {
    const response = await fetch(`${API_BASE}/folders`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse<{
      id: string;
      name: string;
      project_id: string;
      parent_id?: string;
      created_at: string;
    }>(response);
  },

  async updateFolder(folderId: string, data: { name?: string }) {
    const response = await fetch(`${API_BASE}/folders/${folderId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse<{
      id: string;
      name: string;
      project_id: string;
      parent_id?: string;
      updated_at: string;
    }>(response);
  },

  async deleteFolder(folderId: string) {
    const response = await fetch(`${API_BASE}/folders/${folderId}`, {
      method: 'DELETE',
    });
    return handleResponse<{ success: boolean }>(response);
  },

  // Version history endpoints
  async getVisualizationHistory(projectId: string, vizId: string) {
    const response = await fetch(`${API_BASE}/history/${projectId}/${vizId}`);
    return handleResponse<{
      versions: Array<{
        id: string;
        visualization_id: string;
        version: number;
        created_at: string;
        artifacts: Array<{
          id: string;
          kind: string;
          mime_type: string;
          data?: string;
          summary?: string;
        }>;
      }>;
    }>(response);
  },

  async restoreVersion(projectId: string, vizId: string, versionId: string) {
    const response = await fetch(`${API_BASE}/workspace/restore`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ project_id: projectId, visualization_id: vizId, version_id: versionId }),
    });
    return handleResponse<{
      success: boolean;
      session_id: string;
      artifacts: Array<{
        id: string;
        kind: string;
        mime_type: string;
        data?: string;
        summary?: string;
      }>;
    }>(response);
  },

  // Template management endpoints
  async listTemplates() {
    const response = await fetch(`${API_BASE}/templates`);
    return handleResponse<{
      templates: Array<{
        id: string;
        name: string;
        description?: string;
        category: string;
        thumbnail?: string;
        created_at: string;
      }>;
    }>(response);
  },

  async createTemplate(data: { name: string; description?: string; category: string; content: string }) {
    const response = await fetch(`${API_BASE}/templates`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse<{
      id: string;
      name: string;
      description?: string;
      category: string;
      created_at: string;
    }>(response);
  },

  async updateTemplate(templateId: string, data: { name?: string; description?: string; category?: string; content?: string }) {
    const response = await fetch(`${API_BASE}/templates/${templateId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    });
    return handleResponse<{
      id: string;
      name: string;
      description?: string;
      category: string;
      updated_at: string;
    }>(response);
  },

  async deleteTemplate(templateId: string) {
    const response = await fetch(`${API_BASE}/templates/${templateId}`, {
      method: 'DELETE',
    });
    return handleResponse<{ success: boolean }>(response);
  },
};
