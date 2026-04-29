export interface ProgressLineProps {
  progress: number;
  className?: string;
}

export function ProgressLine({ progress, className = '' }: ProgressLineProps) {
  const clampedProgress = Math.min(100, Math.max(0, progress));

  return (
    <div className={`w-full h-1.5 bg-muted rounded-full overflow-hidden ${className}`}>
      <div
        className="h-full bg-primary rounded-full transition-all duration-300"
        style={{ width: `${clampedProgress}%` }}
      />
    </div>
  );
}
