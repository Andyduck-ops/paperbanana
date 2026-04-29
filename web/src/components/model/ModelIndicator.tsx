export interface ModelIndicatorProps {
  modelName?: string;
  isActive?: boolean;
  className?: string;
}

export function ModelIndicator({ modelName = 'Default Model', isActive = false, className = '' }: ModelIndicatorProps) {
  return (
    <div className={`flex items-center gap-2 ${isActive ? 'opacity-100' : 'opacity-60'} ${className}`}>
      <span className={`w-2 h-2 rounded-full ${isActive ? 'bg-status-success' : 'bg-muted-foreground'}`} />
      <span className="text-xs text-muted-foreground">{modelName}</span>
    </div>
  );
}
