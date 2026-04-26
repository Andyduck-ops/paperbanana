import { useState } from 'react';
import { FolderTree, VersionTimeline, TemplateManager } from './';
import { useFolders, useVersions, useTemplates, useToast } from '../../hooks';

/**
 * P2 Batch 3 Demo Component
 * 
 * Demonstrates all features from P2-004 through O-002:
 * - P2-004: Folder Tree Sidebar with navigation
 * - P2-005: Folder CRUD with context menu
 * - P2-006: Version Timeline
 * - P2-007: Version Restore
 * - O-002: Template Management
 */
export function P2Batch3Demo() {
  const { addToast } = useToast();
  const [activeTab, setActiveTab] = useState<'folders' | 'versions' | 'templates'>('folders');
  const [demoProjectId] = useState('demo-project');
  const [demoVizId] = useState('demo-viz');

  // Folder hooks
  const {
    items,
    currentFolderId,
    isLoading: foldersLoading,
    navigateToFolder,
    createFolder,
    renameFolder,
    deleteFolder,
  } = useFolders(demoProjectId);

  // Version hooks
  const {
    versions,
    isLoading: versionsLoading,
    restoreVersion,
  } = useVersions(demoProjectId, demoVizId);

  // Template hooks
  const {
    templates,
    isLoading: templatesLoading,
    createTemplate,
    updateTemplate,
    deleteTemplate,
  } = useTemplates();

  const handleCreateFolder = async (parentId?: string) => {
    const name = prompt('Enter folder name:');
    if (name) {
      const result = await createFolder(name, parentId);
      if (result) {
        addToast('Folder created successfully', 'success');
      }
    }
  };

  const handleRenameFolder = async (folderId: string, currentName: string) => {
    const name = prompt('Enter new name:', currentName);
    if (name && name !== currentName) {
      const result = await renameFolder(folderId, name);
      if (result) {
        addToast('Folder renamed successfully', 'success');
      }
    }
  };

  const handleDeleteFolder = async (folderId: string, name: string) => {
    if (confirm(`Delete "${name}"?`)) {
      const success = await deleteFolder(folderId);
      if (success) {
        addToast('Folder deleted successfully', 'success');
      }
    }
  };

  const handleRestore = async (versionId: string) => {
    const result = await restoreVersion(versionId);
    if (result) {
      addToast('Version restored successfully', 'success');
    }
  };

  return (
    <div className="p-6 max-w-6xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">P2 Batch 3 Features Demo</h1>

      {/* Tab Navigation */}
      <div className="flex gap-2 mb-6 border-b border-border pb-2">
        {(['folders', 'versions', 'templates'] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`px-4 py-2 rounded-lg font-medium transition-colors ${
              activeTab === tab
                ? 'bg-primary text-background'
                : 'hover:bg-muted'
            }`}
          >
            {tab.charAt(0).toUpperCase() + tab.slice(1)}
          </button>
        ))}
      </div>

      {/* Content */}
      <div className="bg-card border border-border rounded-xl p-6">
        {activeTab === 'folders' && (
          <div>
            <h2 className="text-lg font-semibold mb-4">P2-004 & P2-005: Folder Tree & CRUD</h2>
            <div className="w-80 border border-border rounded-lg overflow-hidden">
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
            <p className="mt-4 text-sm text-muted-foreground">
              Right-click on folders to see context menu (Create, Rename, Delete)
            </p>
          </div>
        )}

        {activeTab === 'versions' && (
          <div>
            <h2 className="text-lg font-semibold mb-4">P2-006 & P2-007: Version Timeline & Restore</h2>
            <VersionTimeline
              versions={versions}
              isLoading={versionsLoading}
              onRestore={handleRestore}
            />
          </div>
        )}

        {activeTab === 'templates' && (
          <div>
            <h2 className="text-lg font-semibold mb-4">O-002: Template Management</h2>
            <TemplateManager
              templates={templates}
              isLoading={templatesLoading}
              onCreate={createTemplate}
              onUpdate={updateTemplate}
              onDelete={deleteTemplate}
            />
          </div>
        )}
      </div>

      {/* API Documentation */}
      <div className="mt-8 p-6 bg-muted rounded-xl">
        <h3 className="font-semibold mb-4">API Endpoints</h3>
        <div className="space-y-2 text-sm font-mono">
          <div className="flex gap-4">
            <span className="text-green-600 w-16">POST</span>
            <span>/api/v1/folders</span>
            <span className="text-muted-foreground">- Create folder</span>
          </div>
          <div className="flex gap-4">
            <span className="text-blue-600 w-16">PUT</span>
            <span>/api/v1/folders/:id</span>
            <span className="text-muted-foreground">- Rename folder</span>
          </div>
          <div className="flex gap-4">
            <span className="text-red-600 w-16">DELETE</span>
            <span>/api/v1/folders/:id</span>
            <span className="text-muted-foreground">- Delete folder</span>
          </div>
          <div className="flex gap-4">
            <span className="text-green-600 w-16">GET</span>
            <span>/api/v1/history/:project/:viz</span>
            <span className="text-muted-foreground">- Get version history</span>
          </div>
          <div className="flex gap-4">
            <span className="text-green-600 w-16">POST</span>
            <span>/api/v1/workspace/restore</span>
            <span className="text-muted-foreground">- Restore version</span>
          </div>
          <div className="flex gap-4">
            <span className="text-green-600 w-16">GET</span>
            <span>/api/v1/templates</span>
            <span className="text-muted-foreground">- List templates</span>
          </div>
        </div>
      </div>
    </div>
  );
}

export default P2Batch3Demo;
