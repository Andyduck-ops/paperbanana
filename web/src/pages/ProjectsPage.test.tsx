import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ProjectsPage } from './ProjectsPage';
import type { Project } from '../types/api';

// Mock the hooks and API
vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => key,
  }),
  useFocusTrap: () => ({ current: null }),
}));

vi.mock('../lib/api', () => ({
  apiClient: {
    listProjects: vi.fn(),
    deleteProject: vi.fn(),
    createProject: vi.fn(),
  },
}));

import { apiClient } from '../lib/api';

describe('ProjectsPage', () => {
  const mockProjects: Project[] = [
    {
      id: 'project-1',
      name: 'Test Project 1',
      description: 'Description 1',
      created_at: '2024-01-01T00:00:00.000Z',
      updated_at: '2024-01-01T00:00:00.000Z',
    },
    {
      id: 'project-2',
      name: 'Test Project 2',
      created_at: '2024-01-02T00:00:00.000Z',
      updated_at: '2024-01-02T00:00:00.000Z',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders loading state initially', () => {
    (apiClient.listProjects as ReturnType<typeof vi.fn>).mockImplementation(() => new Promise(() => {}));
    const { container } = render(<ProjectsPage />);
    expect(container).toMatchSnapshot();
    expect(screen.getByText('common.loading')).toBeInTheDocument();
  });

  it('renders correctly with projects', async () => {
    (apiClient.listProjects as ReturnType<typeof vi.fn>).mockResolvedValue({ projects: mockProjects });
    const { container } = render(<ProjectsPage />);
    
    await waitFor(() => {
      expect(screen.getByText('Test Project 1')).toBeInTheDocument();
    });
    
    expect(container).toMatchSnapshot();
  });

  it('renders correctly with empty projects', async () => {
    (apiClient.listProjects as ReturnType<typeof vi.fn>).mockResolvedValue({ projects: [] });
    const { container } = render(<ProjectsPage />);
    
    await waitFor(() => {
      expect(screen.getByText('projects.noProjects')).toBeInTheDocument();
    });
    
    expect(container).toMatchSnapshot();
  });

  it('renders correctly with error', async () => {
    (apiClient.listProjects as ReturnType<typeof vi.fn>).mockRejectedValue(new Error('Failed to load'));
    const { container } = render(<ProjectsPage />);
    
    await waitFor(() => {
      expect(screen.getByText('Failed to load')).toBeInTheDocument();
    });
    
    expect(container).toMatchSnapshot();
  });

  it('creates project successfully', async () => {
    const newProject: Project = {
      id: 'project-3',
      name: 'New Project',
      created_at: '2024-01-03T00:00:00.000Z',
      updated_at: '2024-01-03T00:00:00.000Z',
    };
    
    (apiClient.listProjects as ReturnType<typeof vi.fn>).mockResolvedValue({ projects: mockProjects });
    (apiClient.createProject as ReturnType<typeof vi.fn>).mockResolvedValue(newProject);
    
    render(<ProjectsPage />);
    
    await waitFor(() => {
      expect(screen.getByText('Test Project 1')).toBeInTheDocument();
    });
    
    fireEvent.click(screen.getByText('projects.newProject'));
    
    fireEvent.change(screen.getByPlaceholderText('projects.namePlaceholder'), {
      target: { value: 'New Project' },
    });
    
    fireEvent.click(screen.getByText('common.create'));
    
    await waitFor(() => {
      expect(apiClient.createProject).toHaveBeenCalledWith('New Project', undefined);
    });
  });

  it('deletes project successfully', async () => {
    (apiClient.listProjects as ReturnType<typeof vi.fn>).mockResolvedValue({ projects: mockProjects });
    (apiClient.deleteProject as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);
    
    render(<ProjectsPage />);
    
    await waitFor(() => {
      expect(screen.getByText('Test Project 1')).toBeInTheDocument();
    });
    
    const deleteButtons = screen.getAllByTitle('common.delete');
    fireEvent.click(deleteButtons[0]);
    
    fireEvent.click(screen.getByText('common.delete'));
    
    await waitFor(() => {
      expect(apiClient.deleteProject).toHaveBeenCalledWith('project-1');
    });
  });

  it('navigates to project workspace when project is clicked', async () => {
    (apiClient.listProjects as ReturnType<typeof vi.fn>).mockResolvedValue({ projects: mockProjects });
    render(<ProjectsPage />);
    
    await waitFor(() => {
      expect(screen.getByText('Test Project 1')).toBeInTheDocument();
    });
    
    const projectLink = screen.getByText('Test Project 1').closest('a');
    expect(projectLink).toHaveAttribute('href', '/?project=project-1');
  });
});
