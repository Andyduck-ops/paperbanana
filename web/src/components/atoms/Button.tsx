import { forwardRef, memo, type ButtonHTMLAttributes, type ReactNode } from "react";

export type ButtonVariant = "primary" | "secondary" | "ghost" | "danger" | "success";
export type ButtonSize = "sm" | "md" | "lg";

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  isLoading?: boolean;
  leftIcon?: ReactNode;
  rightIcon?: ReactNode;
  fullWidth?: boolean;
}

const variantStyles: Record<ButtonVariant, string> = {
  primary: "bg-primary text-background hover:opacity-90 active:opacity-80",
  secondary: "bg-muted text-foreground hover:bg-muted/80 active:bg-muted/70 border border-border",
  ghost: "bg-transparent text-foreground hover:bg-muted active:bg-muted/80",
  danger: "bg-status-error text-background hover:opacity-90 active:opacity-80",
  success: "bg-status-success text-background hover:opacity-90 active:opacity-80",
};

const sizeStyles: Record<ButtonSize, string> = {
  sm: "h-8 px-3 text-sm rounded-lg",
  md: "h-10 px-4 text-base rounded-lg",
  lg: "h-12 px-6 text-base rounded-xl",
};

export const Button = memo(forwardRef<HTMLButtonElement, ButtonProps>(
  function Button(
    {
      children,
      variant = "primary",
      size = "md",
      isLoading = false,
      leftIcon,
      rightIcon,
      fullWidth = false,
      disabled,
      className = "",
      ...props
    },
    ref
  ) {
    const baseStyles = "inline-flex items-center justify-center gap-2 font-medium transition-all duration-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/50 focus-visible:ring-offset-2 disabled:opacity-40 disabled:cursor-not-allowed active:scale-[0.98]";
    const widthStyles = fullWidth ? "w-full" : "";

    return (
      <button
        ref={ref}
        className={`${baseStyles} ${variantStyles[variant]} ${sizeStyles[size]} ${widthStyles} ${className}`}
        disabled={disabled || isLoading}
        {...props}
      >
        {isLoading ? (
          <span className="animate-spin h-4 w-4 border-2 border-current border-t-transparent rounded-full" />
        ) : leftIcon}
        {children}
        {!isLoading && rightIcon}
      </button>
    );
  }
));

Button.displayName = "Button";
