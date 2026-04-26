import { useState, useRef, useEffect } from 'react';
import { useLanguage } from '../../hooks';
import type { FolderItem } from '../../hooks/useFolders';

export interface FolderTreeProps {
  items: FolderItem[];
  currentFolderId: string | null;
  isLoading: boolean;
  onNavigate: (folderId: string | null, item?: FolderItem) => void;
  onCreateFolder: (parentId?: string) => void;
  onRenameFolder: (folderId: string, currentName: string) => void;
  onDeleteFolder: (folderId: string, name: string) => void;
}

interface ContextMenuState {
  visible: boolean;
  x: number;
  y: number;
  targetId?: string;
  targetName?: string;
  targetType?: 'folder' | 'visualization';
}

export function FolderTree({
  items,
  currentFolderId,
  isLoading,
  onNavigate,
  onCreateFolder,
  onRenameFolder,
  onDeleteFolder,
}: FolderTreeProps) {
  const { t } = useLanguage();
  const [contextMenu, setContextMenu] = useState<ContextMenuState>({
    visible: false,
    x: 0,
    y: 0,
  });
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setContextMenu((prev) => ({ ...prev, visible: false }));
      }
    };

    if (contextMenu.visible) {
      document.addEventListener('click', handleClickOutside);
    }

    return () => document.removeEventListener('click', handleClickOutside);
  }, [contextMenu.visible]);

  const handleContextMenu = (e: React.MouseEvent, item: FolderItem) => {
    e.preventDefault();
    e.stopPropagation();
    setContextMenu({
      visible: true,
      x: e.clientX,
      y: e.clientY,
      targetId: item.id,
      targetName: item.name,
      targetType: item.type,
    });
  };

  const handleCloseContextMenu = () => {
    setContextMenu((prev) => ({ ...prev, visible: false }));
  };

  const handleCreateFolder = () => {
    onCreateFolder(currentFolderId || undefined);
    handleCloseContextMenu();
  };

  const handleRename = () => {
    if (contextMenu.targetId && contextMenu.targetName && contextMenu.targetType === 'folder') {
      onRenameFolder(contextMenu.targetId, contextMenu.targetName);
    }
    handleCloseContextMenu();
  };

  const handleDelete = () => {
    if (contextMenu.targetId && contextMenu.targetName && contextMenu.targetType === 'folder') {
      onDeleteFolder(contextMenu.targetId, contextMenu.targetName);
    }
    handleCloseContextMenu();
  };

  if (isLoading) {
    return (
      <div className="folder-tree-loading">
        <div className="w-5 h-5 border-2 border-primary/25 border-t-primary rounded-full animate-spin" />
      </div>
    );
  }

  return (
    <div className="folder-tree">
      {/* Root/Home item */}
      <div
        className={`folder-tree-item ${currentFolderId === null ? 'folder-tree-item--active' : ''}`}
        onClick={() => onNavigate(null)}
        onContextMenu={(e) => handleContextMenu(e, { id: 'root', name: 'Root', type: 'folder', created_at: '' })}
      >
        <svg className="folder-tree-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 12l8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25" />
        </svg>
        <span className="folder-tree-label">{t('folder.root') || 'Home'}</span>
      </div>

      {/* Create new folder button */}
      <button
        className="folder-tree-action"
        onClick={() => onCreateFolder(currentFolderId || undefined)}
      >
        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
        </svg>
        <span>{t('folder.newFolder') || 'New Folder'}</span>
      </button>

      {/* Items list */}
      <div className="folder-tree-list">
        {items.length === 0 ? (
          <div className="folder-tree-empty">
            {t('folder.empty') || 'No items'}
          </div>
        ) : (
          items.map((item) => (
            <div
              key={item.id}
              className={`folder-tree-item ${item.type === 'visualization' ? 'folder-tree-item--viz' : ''}`}
              onClick={() => onNavigate(item.id, item)}
              onContextMenu={(e) => handleContextMenu(e, item)}
            >
              {item.type === 'folder' ? (
                <svg className="folder-tree-icon folder-tree-icon--folder" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 12.75V12A2.25 2.25 0 014.5 9.75h15A2.25 2.25 0 0121.75 12v.75m-8.69-6.44l-2.12-2.12a1.5 1.5 0 00-1.061-.44H4.5A2.25 2.25 0 002.25 6v12a2.25 2.25 0 002.25 2.25h15A2.25 2.25 0 0021.75 18V9a2.25 2.25 0 00-2.25-2.25h-5.379a1.5 1.5 0 01-1.06-.44z" />
                </svg>
              ) : (
                <svg className="folder-tree-icon folder-tree-icon--viz" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 3v11.25A2.25 2.25 0 006 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0118 16.5h-2.25m-7.5 0h7.5m-7.5 0l-1 3m8.5-3l1 3m0 0l.5 1.5m-.5-1.5h-9.5m0 0l-.5 1.5m.75-9l3-3 2.148 2.148A12.061 12.061 0 0116.5 7.605" />
                </svg>
              )}
              <span className="folder-tree-label" title={item.name}>
                {item.name}
              </span>
            </div>
          ))
        )}
      </div>

      {/* Context Menu */}
      {contextMenu.visible && (
        <div
          ref={menuRef}
          className="folder-context-menu"
          style={{
            position: 'fixed',
            left: contextMenu.x,
            top: contextMenu.y,
            zIndex: 100,
          }}
        >
          {contextMenu.targetType === 'folder' && contextMenu.targetId !== 'root' && (
            <>
              <button className="folder-context-menu-item" onClick={handleRename}>
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                </svg>
                <span>{t('folder.rename') || 'Rename'}</span>
              </button>
              <button className="folder-context-menu-item folder-context-menu-item--danger" onClick={handleDelete}>
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12.562 3.034a2.25 2.25 0 01-1.52-2.228m0 0a48.394 48.394 0 013.36-.379m12.562 3.034a2.25 2.25 0 011.52-2.228m0 0a48.394 48.394 0 013.36-.379m0 0a1.5 1.5 0 01-1.484-1.475v-.993a1.5 1.5 0 011.484-1.475" />
                </svg>
                <span>{t('folder.delete') || 'Delete'}</span>
              </button>
              <div className="folder-context-menu-divider" />
            </>
          )}
          <button className="folder-context-menu-item" onClick={handleCreateFolder}>
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
            </svg>
            <span>{t('folder.newFolder') || 'New Folder'}</span>
          </button>
        </div>
      )}
    </div>
  );
}
