import type { Provider } from '../../stores';

export interface ChannelManagerProps {
  channels?: Provider[];
  onAddChannel?: (provider: Omit<Provider, 'id' | 'models'>) => Promise<void>;
  onUpdateChannel?: (id: string, updates: Partial<Provider>) => Promise<void>;
  onDeleteChannel?: (id: string) => Promise<void>;
  onFetchModels?: (channelId: string) => Promise<void>;
  onChannelSelect?: (channelId: string) => void;
  selectedChannelId?: string;
  loading?: boolean;
  className?: string;
}

export function ChannelManager({
  channels = [],
  onAddChannel: _onAddChannel,
  onUpdateChannel: _onUpdateChannel,
  onDeleteChannel: _onDeleteChannel,
  onFetchModels: _onFetchModels,
  onChannelSelect,
  selectedChannelId,
  loading = false,
  className = ''
}: ChannelManagerProps) {
  return (
    <div className={`bg-card rounded-xl border border-border/30 p-6 ${className}`}>
      <div className="mb-4">
        <h3 className="text-lg font-semibold text-foreground">Channels</h3>
      </div>
      <div className="space-y-1">
        {channels.map((channel) => (
          <button
            key={channel.id}
            className={`w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-left text-sm transition-colors ${selectedChannelId === channel.id ? 'bg-primary/10 text-primary' : 'text-foreground hover:bg-muted/40'}`}
            onClick={() => onChannelSelect?.(channel.id)}
          >
            <span className="font-medium">{channel.display_name || channel.name}</span>
            <span className="text-xs text-muted-foreground">{channel.models.length} models</span>
          </button>
        ))}
        {channels.length === 0 && !loading && (
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-sm">No channels configured</p>
          </div>
        )}
        {loading && (
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-sm">Loading channels...</p>
          </div>
        )}
      </div>
    </div>
  );
}
