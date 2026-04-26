import { memo, type ReactNode, type HTMLAttributes } from "react";

export type CardVariant = "default" | "elevated" | "outlined" | "ghost";

export interface CardProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
  variant?: CardVariant;
  padding?: "none" | "sm" | "md" | "lg";
  hover?: boolean;
  className?: string;
}

const variantStyles: Record<CardVariant, string> = {
  default: "bg-surface border border-border shadow-sm",
  elevated: "bg-surface border border-border shadow-md",
  outlined: "bg-transparent border border-border",
  ghost: "bg-surface-1/50 border-transparent",
};

const paddingStyles = {
  none: "",
  sm: "p-3",
  md: "p-4",
  lg: "p-6",
};

export const Card = memo(function Card({
  children,
  variant = "default",
  padding = "md",
  hover = false,
  className = "",
  ...props
}: CardProps) {
  return (
    <div
      className={`rounded-xl transition-all ${variantStyles[variant]} ${paddingStyles[padding]} ${hover ? "hover:shadow-lg hover:border-primary/20 cursor-pointer" : ""} ${className}`}
      {...props}
    >
      {children}
    </div>
  );
});

Card.displayName = "Card";

// Card sub-components
export interface CardHeaderProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
  className?: string;
}

export const CardHeader = memo(function CardHeader({
  children,
  className = "",
  ...props
}: CardHeaderProps) {
  return (
    <div className={`flex items-center justify-between mb-4 ${className}`} {...props}>
      {children}
    </div>
  );
});

CardHeader.displayName = "CardHeader";

export interface CardTitleProps extends HTMLAttributes<HTMLHeadingElement> {
  children: ReactNode;
  className?: string;
}

export const CardTitle = memo(function CardTitle({
  children,
  className = "",
  ...props
}: CardTitleProps) {
  return (
    <h3 className={`text-lg font-semibold text-foreground ${className}`} {...props}>
      {children}
    </h3>
  );
});

CardTitle.displayName = "CardTitle";

export interface CardContentProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
  className?: string;
}

export const CardContent = memo(function CardContent({
  children,
  className = "",
  ...props
}: CardContentProps) {
  return (
    <div className={`text-muted-foreground ${className}`} {...props}>
      {children}
    </div>
  );
});

CardContent.displayName = "CardContent";
