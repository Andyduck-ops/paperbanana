import { memo } from 'react';

export interface SkeletonLoaderProps {
  variant?: 'text' | 'card' | 'image' | 'stage' | 'input';
  lines?: number;
  className?: string;
}

function SkeletonLoaderComponent({
  variant = 'text',
  lines = 1,
  className = '',
}: SkeletonLoaderProps) {
  const baseClasses = 'animate-pulse rounded-lg bg-muted/60';

  if (variant === 'text') {
    return (
      <div className={`space-y-2 ${className}`}>
        {Array.from({ length: lines }).map((_, i) => (
          <div
            key={i}
            className={`${baseClasses} h-4`}
            style={{ width: i === lines - 1 && lines > 1 ? '75%' : '100%' }}
          />
        ))}
      </div>
    );
  }

  if (variant === 'card') {
    return (
      <div className={`rounded-2xl border border-border/50 bg-card/50 p-6 space-y-4 ${className}`}>
        <div className={`${baseClasses} h-6 w-1/3`} />
        <div className={`${baseClasses} h-4 w-full`} />
        <div className={`${baseClasses} h-4 w-5/6`} />
        <div className="flex gap-3 pt-2">
          <div className={`${baseClasses} h-10 w-24`} />
          <div className={`${baseClasses} h-10 w-24`} />
        </div>
      </div>
    );
  }

  if (variant === 'image') {
    return (
      <div className={`rounded-2xl border border-border/50 bg-card/50 overflow-hidden ${className}`}>
        <div className={`${baseClasses} w-full aspect-video`} />
        <div className="p-4 space-y-2">
          <div className={`${baseClasses} h-5 w-1/2`} />
          <div className={`${baseClasses} h-3 w-1/3`} />
        </div>
      </div>
    );
  }

  if (variant === 'stage') {
    return (
      <div className={`space-y-3 ${className}`}>
        {Array.from({ length: lines }).map((_, i) => (
          <div
            key={i}
            className="flex items-center gap-3 p-3 rounded-xl border border-border/50 bg-card/30"
          >
            <div className={`${baseClasses} w-8 h-8 rounded-full flex-shrink-0`} />
            <div className="flex-1 space-y-2">
              <div className={`${baseClasses} h-4 w-1/4`} />
              <div className={`${baseClasses} h-3 w-1/2`} />
            </div>
            <div className={`${baseClasses} h-6 w-16 rounded-full flex-shrink-0`} />
          </div>
        ))}
      </div>
    );
  }

  if (variant === 'input') {
    return (
      <div className={`space-y-4 ${className}`}>
        <div className={`${baseClasses} h-10 w-full rounded-xl`} />
        <div className={`${baseClasses} h-32 w-full rounded-xl`} />
        <div className={`${baseClasses} h-10 w-full rounded-xl`} />
      </div>
    );
  }

  return null;
}

export const SkeletonLoader = memo(SkeletonLoaderComponent);
SkeletonLoader.displayName = 'SkeletonLoader';
