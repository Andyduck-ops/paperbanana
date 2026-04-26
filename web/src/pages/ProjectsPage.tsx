import { useState, useEffect, useCallback } from 'react';
import { useLanguage, useFocusTrap } from '../hooks';
import { apiClient } from '../lib/api';
import type { Project } from '../types/api';

/**
 * Projects Page
 *
 * Displays a list of projects with:
 * - Project cards with key information
 * - Create new project functionality
 * - Delete project with confirmation
 * - Navigate to project workspace
 */
export function ProjectsPage() {
  const { t } = useLanguage();
  const [projects, setProjects] = useState<Project[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState<string | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  // Load projects
  const loadProjects = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const response = await apiClient.listProjects();
      setProjects(response.projects);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load projects');
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadProjects();
  }, [loadProjects]);

  const handleCreateSuccess = (newProject: Project) => {
    setProjects(prev => [newProject, ...prev]);
    setShowCreateModal(false);
  };

  const handleDelete = async (projectId: string) => {
    try {
      setIsDeleting(true);
      await apiClient.deleteProject(projectId);
      setProjects(prev => prev.filter(p => p.id !== projectId));
      setShowDeleteConfirm(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete project');
    } finally {
      setIsDeleting(false);
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="flex items-center gap-2 text-muted-foreground">
          <div className="w-5 h-5 border-2 border-current border-t-transparent rounded-full animate-spin" />
          {t('common.loading') || 'Loading...'}
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b border-border bg-card">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <a
                href="/"
                className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 19l-7-7m0 0l7-7m-7 7h18" />
                </svg>
                {t('common.back') || 'Back'}
              </a>
              <h1 className="text-2xl font-semibold text-foreground">
                {t('projects.title') || 'Projects'}
              </h1>
            </div>
            <button
              onClick={() => setShowCreateModal(true)}
              className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg font-medium hover:opacity-90 transition-opacity"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              {t('projects.newProject') || 'New Project'}
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {error && (
          <div className="mb-6 p-4 rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400">
            <div className="flex items-center gap-2">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              {error}
            </div>
          </div>
        )}

        {projects.length === 0 ? (
          <div className="text-center py-16">
            <div className="w-16 h-16 mx-auto mb-4 rounded-full bg-muted flex items-center justify-center">
              <svg className="w-8 h-8 text-muted-foreground" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-foreground mb-2">
              {t('projects.noProjects') || 'No projects yet'}
            </h3>
            <p className="text-muted-foreground mb-4">
              {t('projects.createFirst') || 'Create your first project to get started'}
            </p>
            <button
              onClick={() => setShowCreateModal(true)}
              className="inline-flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg font-medium hover:opacity-90 transition-opacity"
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              {t('projects.createProject') || 'Create Project'}
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {projects.map((project) => (
              <ProjectCard
                key={project.id}
                project={project}
                onDelete={() => setShowDeleteConfirm(project.id)}
                formatDate={formatDate}
              />
            ))}
          </div>
        )}
      </main>

      {/* Create Project Modal */}
      {showCreateModal && (
        <CreateProjectModal
          onClose={() => setShowCreateModal(false)}
          onSuccess={handleCreateSuccess}
        />
      )}

      {/* Delete Confirmation Modal */}
      {showDeleteConfirm && (
        <DeleteConfirmModal
          projectName={projects.find(p => p.id === showDeleteConfirm)?.name || ''}
          onClose={() => setShowDeleteConfirm(null)}
          onConfirm={() => handleDelete(showDeleteConfirm)}
          isDeleting={isDeleting}
        />
      )}
    </div>
  );
}

// ============================================
// Project Card Component
// ============================================

interface ProjectCardProps {
  project: Project;
  onDelete: () => void;
  formatDate: (date: string) => string;
}

function ProjectCard({ project, onDelete, formatDate }: ProjectCardProps) {
  const { t } = useLanguage();

  return (
    <div className="group relative bg-card rounded-xl border border-border p-6 hover:shadow-lg hover:border-primary/20 transition-all">
      <div className="flex items-start justify-between mb-4">
        <div className="w-12 h-12 rounded-lg bg-primary/10 flex items-center justify-center">
          <svg className="w-6 h-6 text-primary" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
        </div>
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete();
          }}
          className="opacity-0 group-hover:opacity-100 p-2 rounded-lg text-muted-foreground hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-all"
          title={t('common.delete') || 'Delete'}
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
        </button>
      </div>

      <a href={`/?project=${project.id}`} className="block">
        <h3 className="text-lg font-semibold text-foreground mb-1 hover:text-primary transition-colors">
          {project.name}
        </h3>
        {project.description && (
          <p className="text-sm text-muted-foreground mb-3 line-clamp-2">
            {project.description}
          </p>
        )}
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          {t('projects.created') || 'Created'} {formatDate(project.created_at)}
        </div>
      </a>
    </div>
  );
}

// ============================================
// Create Project Modal
// ============================================

interface CreateProjectModalProps {
  onClose: () => void;
  onSuccess: (project: Project) => void;
}

function CreateProjectModal({ onClose, onSuccess }: CreateProjectModalProps) {
  const { t } = useLanguage();
  const dialogRef = useFocusTrap<HTMLDivElement>(true);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    try {
      setIsSubmitting(true);
      setError(null);
      const project = await apiClient.createProject(name.trim(), description.trim() || undefined);
      onSuccess(project);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create project');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div
        ref={dialogRef}
        className="bg-background rounded-2xl shadow-2xl max-w-md w-full"
        role="dialog"
        aria-modal="true"
        aria-labelledby="create-project-title"
      >
        <div className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 id="create-project-title" className="text-xl font-semibold">
              {t('projects.newProject') || 'New Project'}
            </h3>
            <button
              type="button"
              onClick={onClose}
              className="p-2 rounded-lg hover:bg-muted transition-colors"
              aria-label={t('common.close') || 'Close'}
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {error && (
            <div className="mb-4 p-3 rounded-lg bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit}>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">
                  {t('projects.name') || 'Name'} *
                </label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder={t('projects.namePlaceholder') || 'Enter project name'}
                  className="w-full px-3 py-2 rounded-lg border border-border bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
                  autoFocus
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-foreground mb-1">
                  {t('projects.description') || 'Description'}
                </label>
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  placeholder={t('projects.descriptionPlaceholder') || 'Enter project description (optional)'}
                  rows={3}
                  className="w-full px-3 py-2 rounded-lg border border-border bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 resize-none"
                />
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 mt-6">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 rounded-lg border border-border text-foreground hover:bg-muted transition-colors"
              >
                {t('common.cancel') || 'Cancel'}
              </button>
              <button
                type="submit"
                disabled={!name.trim() || isSubmitting}
                className="px-4 py-2 rounded-lg bg-primary text-primary-foreground font-medium hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isSubmitting ? (
                  <span className="flex items-center gap-2">
                    <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
                    {t('common.creating') || 'Creating...'}
                  </span>
                ) : (
                  t('common.create') || 'Create'
                )}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

// ============================================
// Delete Confirmation Modal
// ============================================

interface DeleteConfirmModalProps {
  projectName: string;
  onClose: () => void;
  onConfirm: () => void;
  isDeleting: boolean;
}

function DeleteConfirmModal({ projectName, onClose, onConfirm, isDeleting }: DeleteConfirmModalProps) {
  const { t } = useLanguage();
  const dialogRef = useFocusTrap<HTMLDivElement>(true);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div
        ref={dialogRef}
        className="bg-background rounded-2xl shadow-2xl max-w-md w-full"
        role="alertdialog"
        aria-modal="true"
        aria-labelledby="delete-project-title"
        aria-describedby="delete-project-desc"
      >
        <div className="p-6">
          <div className="flex items-center gap-4 mb-4">
            <div className="w-12 h-12 rounded-full bg-red-100 dark:bg-red-900/30 flex items-center justify-center flex-shrink-0">
              <svg className="w-6 h-6 text-red-600 dark:text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <div>
              <h3 id="delete-project-title" className="text-lg font-semibold text-foreground">
                {t('projects.deleteTitle') || 'Delete Project'}
              </h3>
              <p id="delete-project-desc" className="text-sm text-muted-foreground mt-1">
                {t('projects.deleteConfirm') || 'This action cannot be undone'}
              </p>
            </div>
          </div>

          <div className="bg-muted rounded-lg p-4 mb-6">
            <p className="text-sm text-muted-foreground mb-1">{t('projects.projectName') || 'Project name'}:</p>
            <p className="font-medium text-foreground">{projectName}</p>
          </div>

          <div className="flex items-center justify-end gap-3">
            <button
              type="button"
              onClick={onClose}
              disabled={isDeleting}
              className="px-4 py-2 rounded-lg border border-border text-foreground hover:bg-muted transition-colors disabled:opacity-50"
            >
              {t('common.cancel') || 'Cancel'}
            </button>
            <button
              type="button"
              onClick={onConfirm}
              disabled={isDeleting}
              className="px-4 py-2 rounded-lg bg-red-600 text-white font-medium hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isDeleting ? (
                <span className="flex items-center gap-2">
                  <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  {t('common.deleting') || 'Deleting...'}
                </span>
              ) : (
                t('common.delete') || 'Delete'
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default ProjectsPage;
