// @ts-nocheck - This page requires react-router-dom which is not installed
import { useState } from 'react';
// import { useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { VersionTimeline, FolderTree } from '../components';
import { useVersions, useFolders, useToast } from '../hooks';

/**
 * Visualization Detail Page
 * 
 * Demonstrates:
 * - FolderTree component for navigation
 * - VersionTimeline component for version history and restore
 */
export function VisualizationDetailPage() {
  const { t } = useTranslation();
  // const { projectId, vizId } = useParams<{ projectId: string; vizId: string }>();
  const projectId = 'default';
  const vizId = 'default';
  const { addToast } = useToast();
  const [showVersionPanel, setShowVersionPanel] = useState(false);

  const {
    versions,
    isLoading: versionsLoading,
    restoreVersion,
    refresh: _refreshVersions,
  } = useVersions(projectId, vizId);

  const {
    items,
    currentFolderId,
    isLoading: foldersLoading,
    navigateToFolder,
    createFolder,
    renameFolder,
    deleteFolder,
    refresh: _refreshFolders,
  } = useFolders(projectId);

  const handleRestore = async (versionId: string) => {
    const result = await restoreVersion(versionId);
    if (result) {
      addToast(t('version.restored') || 'Version restored successfully', 'success');
    } else {
      addToast(t('version.error') || 'Failed to restore version', 'error');
    }
  };

  const handleCreateFolder = async (parentId?: string) => {
    const name = prompt(t('folder.createPlaceholder') || 'Enter folder name...');
    if (name) {
      const result = await createFolder(name, parentId);
      if (result) {
        addToast(t('folder.created') || 'Folder created', 'success');
      }
    }
  };

  const handleRenameFolder = async (folderId: string, currentName: string) => {
    const name = prompt(t('folder.renamePlaceholder') || 'Enter new name...', currentName);
    if (name && name !== currentName) {
      const result = await renameFolder(folderId, name);
      if (result) {
        addToast(t('folder.renamed') || 'Folder renamed', 'success');
      }
    }
  };

  const handleDeleteFolder = async (folderId: string, name: string) => {
    const confirmed = confirm((t('folder.confirmDelete') || 'Delete {{name}}?').replace('{{name}}', name));
    if (confirmed) {
      const success = await deleteFolder(folderId);
      if (success) {
        addToast(t('folder.deleted') || 'Folder deleted', 'success');
      }
    }
  };

  return (
    <div className="min-h-screen bg-background">
      <div className="flex h-screen">
        {/* Folder Sidebar */}
        <aside className="folder-sidebar">
          <div className="folder-sidebar__header">
            <h2 className="folder-sidebar__title">{t('folder.title')}</h2>
          </div>
          <div className="folder-sidebar__content">
            <FolderTree
              items={items}
              currentFolderId={currentFolderId}
              isLoading={foldersLoading}
              onNavigate={navigateToFolder}
              onCreateFolder={handleCreateFolder}
              onRenameFolder={handleRenameFolder}
              onDeleteFolder={handleDeleteFolder}
            />
          </div>
        </aside>

        {/* Main Content */}
        <main className="flex-1 overflow-auto p-6">
          <div className="max-w-4xl mx-auto">
            <div className="flex items-center justify-between mb-6">
              <h1 className="text-2xl font-bold">{t('version.title')}</h1>
              <button
                onClick={() => setShowVersionPanel(!showVersionPanel)}
                className="px-4 py-2 bg-primary text-background rounded-lg hover:opacity-90 transition-opacity"
              >
                {showVersionPanel ? t('common.hide') : t('common.show')} {t('version.versions')}
              </button>
            </div>

            {showVersionPanel && (
              <div className="bg-card border border-border rounded-xl p-4">
                <VersionTimeline
                  versions={versions}
                  isLoading={versionsLoading}
                  onRestore={handleRestore}
                />
              </div>
            )}

            {/* Placeholder for visualization content */}
            <div className="mt-8 p-8 bg-muted rounded-xl text-center text-muted-foreground">
              <p>{t('common.select')} {t('version.preview')}</p>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}

export default VisualizationDetailPage;
