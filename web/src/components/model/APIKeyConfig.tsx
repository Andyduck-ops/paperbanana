import { useState } from 'react';

interface APIKeyConfigProps {
  providerId: string;
  providerName: string;
  onKeyAdded: (key: string) => void;
  onClose: () => void;
}

export function APIKeyConfig({
  providerId,
  providerName,
  onKeyAdded,
  onClose,
}: APIKeyConfigProps) {
  const [apiKey, setApiKey] = useState('');
  const [testing, setTesting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!apiKey.trim()) return;

    setTesting(true);
    setError(null);

    try {
      const response = await fetch(`/api/v1/providers/${providerId}/keys`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key: apiKey.trim() }),
      });

      if (!response.ok) {
        const data = await response.json().catch(() => ({}));
        throw new Error(data.error || 'Failed to add API key');
      }

      setSuccess(true);
      onKeyAdded(apiKey.trim());
      setTimeout(() => onClose(), 1500);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setTesting(false);
    }
  };

  const handleTestConnection = async () => {
    if (!apiKey.trim()) return;

    setTesting(true);
    setError(null);

    try {
      const response = await fetch(`/api/v1/providers/${providerId}/test`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ key: apiKey.trim() }),
      });

      const data = await response.json();

      if (!response.ok || !data.success) {
        throw new Error(data.error || 'Connection test failed');
      }

      setSuccess(true);
      setTimeout(() => setSuccess(false), 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Test failed');
    } finally {
      setTesting(false);
    }
  };

  return (
    <div className="p-4 bg-card/95 backdrop-blur-lg rounded-xl border border-border/30 shadow-2xl">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-foreground">
          Configure {providerName}
        </h3>
        <button
          onClick={onClose}
          className="w-6 h-6 rounded-md flex items-center justify-center hover:bg-muted/50 transition-colors"
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <form onSubmit={handleSubmit} className="space-y-3">
        <div>
          <label className="block text-xs text-muted-foreground mb-1">
            API Key
          </label>
          <input
            type="password"
            value={apiKey}
            onChange={(e) => setApiKey(e.target.value)}
            placeholder="Enter your API key..."
            className="w-full px-3 py-2 rounded-lg bg-muted/30 border border-border/30 text-sm
                       focus:outline-none focus:ring-2 focus:ring-primary/30 focus:border-primary/30"
          />
        </div>

        {error && (
          <div className="px-3 py-2 rounded-lg bg-red-500/10 border border-red-500/20 text-xs text-red-500">
            {error}
          </div>
        )}

        {success && (
          <div className="px-3 py-2 rounded-lg bg-green-500/10 border border-green-500/20 text-xs text-green-500">
            ✓ Success!
          </div>
        )}

        <div className="flex gap-2">
          <button
            type="button"
            onClick={handleTestConnection}
            disabled={!apiKey.trim() || testing}
            className="flex-1 px-3 py-2 rounded-lg bg-muted/50 text-sm text-foreground
                       hover:bg-muted transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {testing ? 'Testing...' : 'Test'}
          </button>
          <button
            type="submit"
            disabled={!apiKey.trim() || testing}
            className="flex-1 px-3 py-2 rounded-lg bg-primary text-sm text-background
                       hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Save
          </button>
        </div>
      </form>
    </div>
  );
}
