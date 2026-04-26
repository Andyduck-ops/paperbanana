import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ProjectSelector, Project } from './ProjectSelector';

const mockProjects: Project[] = [
  { id: 'proj-1', name: 'Project Alpha', description: 'First project', created_at: '2024-01-01' },
  { id: 'proj-2', name: 'Project Beta', created_at: '2024-01-02' },
  { id: 'proj-3', name: 'Project Gamma', description: 'Third project', created_at: '2024-01-03' },
];

describe('ProjectSelector', () => {
  const mockOnSelect = vi.fn();
  const fetchMock = vi.fn();

  beforeEach(() => {
    localStorage.clear();
    mockOnSelect.mockClear();
    fetchMock.mockClear();
    fetchMock.mockResolvedValue({
      ok: true,
      json: async () => ({ projects: mockProjects }),
    });
    vi.stubGlobal('fetch', fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('shows loading state initially', () => {
    render(<ProjectSelector onSelect={mockOnSelect} />);
    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  it('fetches and displays projects', async () => {
    render(<ProjectSelector onSelect={mockOnSelect} />);
    await waitFor(() => {
      expect(screen.getByText('Select a project')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button'));
    await waitFor(() => {
      expect(screen.getByText('Project Alpha')).toBeInTheDocument();
      expect(screen.getByText('Project Beta')).toBeInTheDocument();
      expect(screen.getByText('Project Gamma')).toBeInTheDocument();
    });
  });

  it('shows project descriptions', async () => {
    render(<ProjectSelector onSelect={mockOnSelect} />);
    await waitFor(() => {
      expect(screen.getByText('Select a project')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button'));
    await waitFor(() => {
      expect(screen.getByText('First project')).toBeInTheDocument();
      expect(screen.getByText('Third project')).toBeInTheDocument();
    });
  });

  it('selects a project and calls onSelect', async () => {
    render(<ProjectSelector onSelect={mockOnSelect} />);
    await waitFor(() => {
      expect(screen.getByText('Select a project')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button'));
    await waitFor(() => {
      expect(screen.getByText('Project Alpha')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText('Project Alpha'));
    expect(mockOnSelect).toHaveBeenCalledWith(mockProjects[0]);
  });

  it('persists selected project to localStorage', async () => {
    render(<ProjectSelector onSelect={mockOnSelect} />);
    await waitFor(() => {
      expect(screen.getByText('Select a project')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button'));
    await waitFor(() => {
      expect(screen.getByText('Project Alpha')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText('Project Alpha'));
    expect(localStorage.getItem('paperbanana_current_project')).toBe('proj-1');
  });

  it('shows selected project name after selection', async () => {
    render(<ProjectSelector onSelect={mockOnSelect} />);
    await waitFor(() => {
      expect(screen.getByText('Select a project')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button'));
    await waitFor(() => {
      expect(screen.getByText('Project Alpha')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText('Project Alpha'));
    expect(screen.getByText('Project Alpha')).toBeInTheDocument();
  });

  it('uses cached projects from localStorage', async () => {
    localStorage.setItem('paperbanana_projects_cache', JSON.stringify({
      projects: mockProjects,
      timestamp: Date.now(),
    }));
    render(<ProjectSelector onSelect={mockOnSelect} />);
    expect(screen.getByText('Select a project')).toBeInTheDocument();
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it('ignores expired cache', async () => {
    localStorage.setItem('paperbanana_projects_cache', JSON.stringify({
      projects: mockProjects,
      timestamp: Date.now() - 10 * 60 * 1000,
    }));
    render(<ProjectSelector onSelect={mockOnSelect} />);
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalled();
    });
  });

  it('shows error state on fetch failure', async () => {
    fetchMock.mockResolvedValue({ ok: false, status: 500 });
    render(<ProjectSelector onSelect={mockOnSelect} />);
    await waitFor(() => {
      expect(screen.getByText('Select a project')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button'));
    await waitFor(() => {
      expect(screen.getByText('HTTP 500')).toBeInTheDocument();
    });
  });

  it('shows retry button on error', async () => {
    fetchMock.mockResolvedValueOnce({ ok: false, status: 500 });
    render(<ProjectSelector onSelect={mockOnSelect} />);
    await waitFor(() => {
      expect(screen.getByText('Select a project')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button'));
    await waitFor(() => {
      expect(screen.getByText('Retry')).toBeInTheDocument();
    });
  });

  it('retries fetch when Retry is clicked', async () => {
    fetchMock.mockResolvedValueOnce({ ok: false, status: 500 });
    fetchMock.mockResolvedValueOnce({
      ok: true,
      json: async () => ({ projects: mockProjects }),
    });
    render(<ProjectSelector onSelect={mockOnSelect} />);
    await waitFor(() => {
      expect(screen.getByText('Select a project')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button'));
    await waitFor(() => {
      expect(screen.getByText('Retry')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText('Retry'));
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(2);
    });
  });

  it('shows empty state when no projects', async () => {
    fetchMock.mockResolvedValue({
      ok: true,
      json: async () => ({ projects: [] }),
    });
    render(<ProjectSelector onSelect={mockOnSelect} />);
    await waitFor(() => {
      expect(screen.getByText('Select a project')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button'));
    await waitFor(() => {
      expect(screen.getByText('No projects available')).toBeInTheDocument();
    });
  });

  it('is disabled when disabled prop is true', async () => {
    render(<ProjectSelector onSelect={mockOnSelect} disabled={true} />);
    await waitFor(() => {
      expect(screen.getByRole('button')).toBeDisabled();
    });
  });

  it('uses custom placeholder', async () => {
    render(<ProjectSelector onSelect={mockOnSelect} placeholder="Choose a project..." />);
    await waitFor(() => {
      expect(screen.getByText('Choose a project...')).toBeInTheDocument();
    });
  });

  it('uses selectedId prop to show selected project', async () => {
    render(<ProjectSelector onSelect={mockOnSelect} selectedId="proj-2" />);
    await waitFor(() => {
      expect(screen.getByText('Project Beta')).toBeInTheDocument();
    });
  });

  it('toggles dropdown on click', async () => {
    render(<ProjectSelector onSelect={mockOnSelect} />);
    await waitFor(() => {
      expect(screen.getByText('Select a project')).toBeInTheDocument();
    });
    const button = screen.getByRole('button');
    fireEvent.click(button);
    await waitFor(() => {
      expect(screen.getByRole('listbox')).toBeInTheDocument();
    });
    fireEvent.click(button);
    await waitFor(() => {
      expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
    });
  });

  it('closes dropdown after selection', async () => {
    render(<ProjectSelector onSelect={mockOnSelect} />);
    await waitFor(() => {
      expect(screen.getByText('Select a project')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button'));
    await waitFor(() => {
      expect(screen.getByText('Project Alpha')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText('Project Alpha'));
    await waitFor(() => {
      expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
    });
  });

  it('shows aria-expanded correctly', async () => {
    render(<ProjectSelector onSelect={mockOnSelect} />);
    await waitFor(() => {
      expect(screen.getByText('Select a project')).toBeInTheDocument();
    });
    const button = screen.getByRole('button');
    expect(button).toHaveAttribute('aria-expanded', 'false');
    fireEvent.click(button);
    await waitFor(() => {
      expect(button).toHaveAttribute('aria-expanded', 'true');
    });
  });

  it('marks selected option with aria-selected', async () => {
    render(<ProjectSelector onSelect={mockOnSelect} selectedId="proj-1" />);
    await waitFor(() => {
      expect(screen.getByText('Project Alpha')).toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole('button'));
    await waitFor(() => {
      const options = screen.getAllByRole('option');
      expect(options[0]).toHaveAttribute('aria-selected', 'true');
      expect(options[1]).toHaveAttribute('aria-selected', 'false');
    });
  });
});
