import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BatchProgressPanel, BatchProgress } from './BatchProgressPanel';

// Mock the useLanguage hook
vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const translations: Record<string, string> = {
        'generate.batchProgress': 'Batch Progress',
        'generate.candidate': `Candidate ${params?.number ?? ''}`,
        'generate.successful': 'successful',
        'generate.failed': 'failed',
        'generate.batchComplete': 'Batch Complete',
        'generate.downloadAll': 'Download All as ZIP',
        'generate.downloading': 'Downloading...',
      };
      return translations[key] || key;
    },
  }),
}));

describe('BatchProgressPanel', () => {
  let mockFetch: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockFetch = vi.fn();
    globalThis.fetch = mockFetch;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('renders nothing when no batchId', () => {
    const { container } = render(<BatchProgressPanel batchId={null} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders progress when batchId provided', () => {
    render(<BatchProgressPanel batchId="test-batch-123" />);
    expect(screen.getByText('Batch Progress')).toBeInTheDocument();
  });

  it('download button appears when batch status is completed', () => {
    const progress: BatchProgress = {
      batchId: 'test-batch-123',
      status: 'completed',
      candidates: [
        { candidateId: 0, status: 'completed', artifacts: [{ id: 'a1', kind: 'figure', mimeType: 'image/png' }] },
      ],
      successful: 1,
      failed: 0,
      startedAt: '2026-03-18T00:00:00Z',
      completedAt: '2026-03-18T00:01:00Z',
    };

    render(
      <BatchProgressPanel
        batchId="test-batch-123"
        initialProgress={progress}
      />
    );

    expect(screen.getByText('Download All as ZIP')).toBeInTheDocument();
  });

  it('download button is hidden while batch is running', () => {
    const progress: BatchProgress = {
      batchId: 'test-batch-123',
      status: 'running',
      candidates: [
        { candidateId: 0, status: 'running' },
        { candidateId: 1, status: 'pending' },
      ],
      successful: 0,
      failed: 0,
      startedAt: '2026-03-18T00:00:00Z',
    };

    render(
      <BatchProgressPanel
        batchId="test-batch-123"
        initialProgress={progress}
      />
    );

    expect(screen.queryByText('Download All as ZIP')).not.toBeInTheDocument();
  });

  it('click triggers POST to /api/v1/batch/download with batch_id', async () => {
    const mockBlob = new Blob(['fake zip content'], { type: 'application/zip' });
    mockFetch.mockResolvedValueOnce({
      ok: true,
      blob: () => Promise.resolve(mockBlob),
    });

    // Mock URL.createObjectURL
    const mockCreateObjectURL = vi.fn(() => 'blob:mock-url');
    const mockRevokeObjectURL = vi.fn();
    globalThis.URL.createObjectURL = mockCreateObjectURL;
    globalThis.URL.revokeObjectURL = mockRevokeObjectURL;

    const progress: BatchProgress = {
      batchId: 'test-batch-456',
      status: 'completed',
      candidates: [
        { candidateId: 0, status: 'completed', artifacts: [{ id: 'a1', kind: 'figure', mimeType: 'image/png' }] },
      ],
      successful: 1,
      failed: 0,
      startedAt: '2026-03-18T00:00:00Z',
      completedAt: '2026-03-18T00:01:00Z',
    };

    render(
      <BatchProgressPanel
        batchId="test-batch-456"
        initialProgress={progress}
        apiBase="http://localhost:8080"
      />
    );

    const downloadButton = screen.getByText('Download All as ZIP');
    fireEvent.click(downloadButton);

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/api/v1/batch/download', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ batch_id: 'test-batch-456' }),
      });
    });
  });

  it('button shows downloading state during fetch', async () => {
    let resolveFetch: (value: unknown) => void;
    const fetchPromise = new Promise((resolve) => {
      resolveFetch = resolve;
    });
    mockFetch.mockReturnValueOnce(fetchPromise);

    const progress: BatchProgress = {
      batchId: 'test-batch-789',
      status: 'completed',
      candidates: [
        { candidateId: 0, status: 'completed', artifacts: [{ id: 'a1', kind: 'figure', mimeType: 'image/png' }] },
      ],
      successful: 1,
      failed: 0,
      startedAt: '2026-03-18T00:00:00Z',
      completedAt: '2026-03-18T00:01:00Z',
    };

    render(
      <BatchProgressPanel
        batchId="test-batch-789"
        initialProgress={progress}
        apiBase="http://localhost:8080"
      />
    );

    const downloadButton = screen.getByText('Download All as ZIP');
    fireEvent.click(downloadButton);

    // Should show downloading state
    await waitFor(() => {
      expect(screen.getByText('Downloading...')).toBeInTheDocument();
    });

    // Resolve the fetch
    resolveFetch!({
      ok: true,
      blob: () => Promise.resolve(new Blob(['fake'], { type: 'application/zip' })),
    });

    // Should return to normal state
    await waitFor(() => {
      expect(screen.getByText('Download All as ZIP')).toBeInTheDocument();
    });
  });

  it('download button is disabled while downloading', async () => {
    let resolveFetch: (value: unknown) => void;
    const fetchPromise = new Promise((resolve) => {
      resolveFetch = resolve;
    });
    mockFetch.mockReturnValueOnce(fetchPromise);

    const progress: BatchProgress = {
      batchId: 'test-batch-dis',
      status: 'completed',
      candidates: [
        { candidateId: 0, status: 'completed', artifacts: [{ id: 'a1', kind: 'figure', mimeType: 'image/png' }] },
      ],
      successful: 1,
      failed: 0,
      startedAt: '2026-03-18T00:00:00Z',
      completedAt: '2026-03-18T00:01:00Z',
    };

    render(
      <BatchProgressPanel
        batchId="test-batch-dis"
        initialProgress={progress}
        apiBase="http://localhost:8080"
      />
    );

    const downloadButton = screen.getByRole('button', { name: 'Download All as ZIP' });
    fireEvent.click(downloadButton);

    // Button should be disabled during download
    await waitFor(() => {
      const downloadingButton = screen.getByRole('button', { name: 'Downloading...' });
      expect(downloadingButton).toBeDisabled();
    });

    // Resolve the fetch
    resolveFetch!({
      ok: true,
      blob: () => Promise.resolve(new Blob(['fake'], { type: 'application/zip' })),
    });
  });
});
