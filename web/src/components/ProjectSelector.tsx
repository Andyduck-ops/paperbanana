import { useState, useEffect, useCallback, useRef } from 'react';

const PROJECTS_CACHE_KEY = 'paperbanana_projects_cache';
const PROJECTS_CACHE_TTL = 5 * 60 * 1000; // 5 minutes
const CURRENT_PROJECT_KEY = 'paperbanana_current_project';

export interface Project {
  id: string;
  name: string;
  description?: string;
  created_at: string;
  updated_at?: string;
}

interface CachedProjects {
  projects: Project[];
  timestamp: number;
}

function loadCachedProjects(): CachedProjects | null {
  try {
    const cached = localStorage.getItem(PROJECTS_CACHE_KEY);
    if (!cached) return null;
    const parsed = JSON.parse(cached) as CachedProjects;
    if (Date.now() - parsed.timestamp > PROJECTS_CACHE_TTL) {
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

function saveCachedProjects(projects: Project[]): void {
  try {
    localStorage.setItem(PROJECTS_CACHE_KEY, JSON.stringify({
      projects,
      timestamp: Date.now(),
    }));
  } catch {
    // Ignore localStorage errors
  }
}

function loadCurrentProjectId(): string | null {
  try {
    return localStorage.getItem(CURRENT_PROJECT_KEY);
  } catch {
    return null;
  }
}

function saveCurrentProjectId(id: string): void {
  try {
    localStorage.setItem(CURRENT_PROJECT_KEY, id);
  } catch {
    // Ignore localStorage errors
  }
}

export interface ProjectSelectorProps {
  onSelect: (project: Project) => void;
  selectedId?: string;
  disabled?: boolean;
  placeholder?: string;
}

export function ProjectSelector({
  onSelect,
  selectedId,
  disabled = false,
  placeholder,
}: ProjectSelectorProps) {
  const [projects, setProjects] = useState<Project[]>(() => {
    const cached = loadCachedProjects();
    return cached?.projects || [];
  });
  const [loading, setLoading] = useState(() => {
    const cached = loadCachedProjects();
    return !cached;
  });
  const [error, setError] = useState<string | null>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [currentProjectId, setCurrentProjectId] = useState<string | null>(() => loadCurrentProjectId());
  const hasFetchedRef = useRef(false);

  const effectiveSelectedId = selectedId ?? currentProjectId;

  const fetchProjects = useCallback(async (forceRefresh = false) => {
    if (!forceRefresh) {
      const cached = loadCachedProjects();
      if (cached) {
        setProjects(cached.projects);
        setLoading(false);
        return;
      }
    }

    setLoading(true);
    setError(null);
    try {
      const response = await fetch('/api/v1/projects');
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      const data = await response.json();
      const fetchedProjects = data.projects || [];
      setProjects(fetchedProjects);
      saveCachedProjects(fetchedProjects);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load projects');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!hasFetchedRef.current) {
      hasFetchedRef.current = true;
      fetchProjects();
    }
  }, [fetchProjects]);

  const selectedProject = projects.find((p) => p.id === effectiveSelectedId);

  const handleSelect = (project: Project) => {
    setCurrentProjectId(project.id);
    saveCurrentProjectId(project.id);
    onSelect(project);
    setIsOpen(false);
  };

  const handleRefresh = () => {
    fetchProjects(true);
  };

  return (
    <div className="project-selector relative">
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled || loading}
        className="
          w-full px-4 py-2 rounded-lg
          border border-border bg-background
          text-foreground text-left
          focus:outline-none focus:ring-2 focus:ring-primary
          disabled:opacity-50 disabled:cursor-not-allowed
          flex items-center justify-between
        "
        aria-expanded={isOpen}
        aria-haspopup="listbox"
      >
        <span className={selectedProject ? 'text-foreground' : 'text-muted-foreground'}>
          {loading
            ? 'Loading...'
            : selectedProject
            ? selectedProject.name
            : placeholder || 'Select a project'}
        </span>
        <svg
          className={`w-4 h-4 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div
          className="absolute z-10 w-full mt-1 bg-background border border-border rounded-lg shadow-lg max-h-60 overflow-auto"
          role="listbox"
        >
          {error && (
            <div className="px-4 py-2 text-sm text-red-500 flex items-center justify-between">
              <span>{error}</span>
              <button
                type="button"
                onClick={handleRefresh}
                className="text-xs underline hover:no-underline"
              >
                Retry
              </button>
            </div>
          )}
          {projects.length === 0 && !error && !loading && (
            <div className="px-4 py-2 text-sm text-muted-foreground">
              No projects available
            </div>
          )}
          {projects.map((project) => (
            <button
              key={project.id}
              type="button"
              onClick={() => handleSelect(project)}
              className={`
                w-full px-4 py-2 text-left text-sm
                hover:bg-primary/10 transition-colors
                ${project.id === effectiveSelectedId ? 'bg-primary/5 text-primary' : 'text-foreground'}
              `}
              role="option"
              aria-selected={project.id === effectiveSelectedId}
            >
              <div className="font-medium">{project.name}</div>
              {project.description && (
                <div className="text-xs text-muted-foreground">{project.description}</div>
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

export default ProjectSelector;
