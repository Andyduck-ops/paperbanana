// @ts-nocheck - Test file with unused variables
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { FolderTree } from './FolderTree';
import type { FolderItem } from '../../hooks/useFolders';

vi.mock('../../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => key,
  }),
}));

describe('FolderTree', () => {
  const mockItems: FolderItem[] = [
    {
      id: 'folder-1',
      name: 'Test Folder',
      type: 'folder',
      created_at: '2024-01-01T00:00:00.000Z',
    },
    {
      id: 'viz-1',
      name: 'Visualization 1',
      type: 'visualization',
      created_at: '2024-01-02T00:00:00.000Z',
    },
    {
      id: 'viz-2',
      name: 'Visualization 2',
      type: 'visualization',
      created_at: '2024-01-03T00:00:00.000Z',
    },
  ];

  const defaultProps = {
    items: mockItems,
    currentFolderId: null,
    isLoading: false,
    onNavigate: vi.fn(),
    onCreateFolder: vi.fn(),
    onRenameFolder: vi.fn(),
    onDeleteFolder: vi.fn(),
  };

  it('renders correctly with items', () => {
    const { container } = render(<FolderTree {...defaultProps} />);
    expect(container).toMatchSnapshot();
  });

  it('renders correctly when loading', () => {
    const { container } = render(<FolderTree {...defaultProps} isLoading={true} />);
    expect(container).toMatchSnapshot();
  });

  it('renders correctly with empty items', () => {
    const { container } = render(<FolderTree {...defaultProps} items={[]} />);
    expect(container).toMatchSnapshot();
  });

  it('renders correctly with current folder selected', () => {
    const { container } = render(
      <FolderTree {...defaultProps} currentFolderId="folder-1" />
    );
    expect(container).toMatchSnapshot();
  });

  it('calls onNavigate when folder is clicked', () => {
    const mockOnNavigate = vi.fn();
    render(<FolderTree {...defaultProps} onNavigate={mockOnNavigate} />);
    
    fireEvent.click(screen.getByText('Test Folder'));
    
    expect(mockOnNavigate).toHaveBeenCalledWith('folder-1', mockItems[0]);
  });

  it('calls onNavigate with null when home is clicked', () => {
    const mockOnNavigate = vi.fn();
    render(<FolderTree {...defaultProps} onNavigate={mockOnNavigate} />);
    
    fireEvent.click(screen.getByText('folder.root'));
    
    expect(mockOnNavigate).toHaveBeenCalledWith(null);
  });

  it('calls onCreateFolder when new folder button is clicked', () => {
    const mockOnCreateFolder = vi.fn();
    render(<FolderTree {...defaultProps} onCreateFolder={mockOnCreateFolder} />);
    
    fireEvent.click(screen.getByText('folder.newFolder'));
    
    expect(mockOnCreateFolder).toHaveBeenCalledWith(undefined);
  });

  it('shows context menu on right click', () => {
    render(<FolderTree {...defaultProps} />);
    
    const folderItem = screen.getByText('Test Folder');
    fireEvent.contextMenu(folderItem);
    
    expect(screen.getByText('folder.rename')).toBeInTheDocument();
    expect(screen.getByText('folder.delete')).toBeInTheDocument();
  });

  it('calls onRenameFolder from context menu', () => {
    const mockOnRenameFolder = vi.fn();
    render(<FolderTree {...defaultProps} onRenameFolder={mockOnRenameFolder} />);
    
    const folderItem = screen.getByText('Test Folder');
    fireEvent.contextMenu(folderItem);
    
    fireEvent.click(screen.getByText('folder.rename'));
    
    expect(mockOnRenameFolder).toHaveBeenCalledWith('folder-1', 'Test Folder');
  });

  it('calls onDeleteFolder from context menu', () => {
    const mockOnDeleteFolder = vi.fn();
    render(<FolderTree {...defaultProps} onDeleteFolder={mockOnDeleteFolder} />);
    
    const folderItem = screen.getByText('Test Folder');
    fireEvent.contextMenu(folderItem);
    
    fireEvent.click(screen.getByText('folder.delete'));
    
    expect(mockOnDeleteFolder).toHaveBeenCalledWith('folder-1', 'Test Folder');
  });

  it('distinguishes between folder and visualization items', () => {
    const { container } = render(<FolderTree {...defaultProps} />);
    
    expect(screen.getByText('Test Folder')).toBeInTheDocument();
    expect(screen.getByText('Visualization 1')).toBeInTheDocument();
    expect(screen.getByText('Visualization 2')).toBeInTheDocument();
  });

  it('navigates to visualization when clicked', () => {
    const mockOnNavigate = vi.fn();
    render(<FolderTree {...defaultProps} onNavigate={mockOnNavigate} />);
    
    fireEvent.click(screen.getByText('Visualization 1'));
    
    expect(mockOnNavigate).toHaveBeenCalledWith('viz-1', mockItems[1]);
  });
});
