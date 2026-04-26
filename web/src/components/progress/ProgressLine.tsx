export interface ProgressLineProps {
  progress: number;
  className?: string;
}

export function ProgressLine({ progress, className = '' }: ProgressLineProps) {
  const clampedProgress = Math.min(100, Math.max(0, progress));

  return (
    <div className={`progress-line ${className}`}>
      <div
        className="progress-line__fill"
        style={{ width: `${clampedProgress}%` }}
      />
    </div>
  );
}
