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
    <div className={`channel-manager ${className}`}>
      <div className="channel-manager__header">
        <h3>Channels</h3>
      </div>
      <div className="channel-manager__list">
        {channels.map((channel) => (
          <button
            key={channel.id}
            className={`channel-manager__item ${selectedChannelId === channel.id ? 'selected' : ''}`}
            onClick={() => onChannelSelect?.(channel.id)}
          >
            <span className="channel-manager__name">{channel.display_name || channel.name}</span>
            <span className="channel-manager__count">{channel.models.length} models</span>
          </button>
        ))}
        {channels.length === 0 && !loading && (
          <div className="channel-manager__empty">
            <p>No channels configured</p>
          </div>
        )}
        {loading && (
          <div className="channel-manager__loading">
            <p>Loading channels...</p>
          </div>
        )}
      </div>
    </div>
  );
}
