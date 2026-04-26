export interface ModelIndicatorProps {
  modelName?: string;
  isActive?: boolean;
  className?: string;
}

export function ModelIndicator({ modelName = 'Default Model', isActive = false, className = '' }: ModelIndicatorProps) {
  return (
    <div className={`model-indicator ${isActive ? 'active' : ''} ${className}`}>
      <span className="model-indicator__dot" />
      <span className="model-indicator__name">{modelName}</span>
    </div>
  );
}
