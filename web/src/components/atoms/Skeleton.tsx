import { memo, type HTMLAttributes } from "react";

export type SkeletonVariant = "text" | "circular" | "rectangular" | "rounded";

export interface SkeletonProps extends HTMLAttributes<HTMLDivElement> {
  variant?: SkeletonVariant;
  width?: string | number;
  height?: string | number;
  className?: string;
  animate?: boolean;
}

const variantStyles: Record<SkeletonVariant, string> = {
  text: "rounded",
  circular: "rounded-full",
  rectangular: "rounded-none",
  rounded: "rounded-lg",
};

export const Skeleton = memo(function Skeleton({
  variant = "text",
  width,
  height,
  className = "",
  animate = true,
  style,
  ...props
}: SkeletonProps) {
  const sizeStyles = {
    width: typeof width === "number" ? `${width}px` : width,
    height: typeof height === "number" ? `${height}px` : height,
  };

  return (
    <div
      className={`bg-surface-2 ${variantStyles[variant]} ${animate ? "animate-pulse" : ""} ${className}`}
      style={{ ...sizeStyles, ...style }}
      {...props}
    />
  );
});

Skeleton.displayName = "Skeleton";

// Skeleton presets for common patterns
export const SkeletonText = memo(function SkeletonText({
  lines = 1,
  className = "",
}: {
  lines?: number;
  className?: string;
}) {
  return (
    <div className={`flex flex-col gap-2 ${className}`}>
      {Array.from({ length: lines }).map((_, i) => (
        <Skeleton
          key={i}
          variant="text"
          height={16}
          width={i === lines - 1 && lines > 1 ? "75%" : "100%"}
        />
      ))}
    </div>
  );
});

SkeletonText.displayName = "SkeletonText";

export const SkeletonCard = memo(function SkeletonCard({
  hasHeader = true,
  hasFooter = false,
  contentLines = 3,
  className = "",
}: {
  hasHeader?: boolean;
  hasFooter?: boolean;
  contentLines?: number;
  className?: string;
}) {
  return (
    <div className={`p-4 border border-border rounded-xl bg-surface ${className}`}>
      {hasHeader && (
        <div className="flex items-center gap-3 mb-4">
          <Skeleton variant="circular" width={40} height={40} />
          <div className="flex-1">
            <Skeleton variant="text" height={16} width="60%" className="mb-1" />
            <Skeleton variant="text" height={12} width="40%" />
          </div>
        </div>
      )}
      <SkeletonText lines={contentLines} />
      {hasFooter && (
        <div className="flex justify-end gap-2 mt-4">
          <Skeleton variant="rounded" width={80} height={32} />
          <Skeleton variant="rounded" width={80} height={32} />
        </div>
      )}
    </div>
  );
});

SkeletonCard.displayName = "SkeletonCard";

export const SkeletonImage = memo(function SkeletonImage({
  aspectRatio = "16/9",
  className = "",
}: {
  aspectRatio?: string;
  className?: string;
}) {
  return (
    <div className={`relative ${className}`} style={{ aspectRatio }}>
      <Skeleton variant="rounded" className="absolute inset-0" />
    </div>
  );
});

SkeletonImage.displayName = "SkeletonImage";
