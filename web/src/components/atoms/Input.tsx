import { forwardRef, memo, useId, type InputHTMLAttributes } from "react";

export type InputSize = "sm" | "md" | "lg";

export interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'size'> {
  inputSize?: InputSize;
  error?: string;
  label?: string;
  helperText?: string;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
}

const sizeStyles: Record<InputSize, string> = {
  sm: "h-8 px-3 text-sm",
  md: "h-10 px-4 text-base",
  lg: "h-12 px-4 text-lg",
};

export const Input = memo(forwardRef<HTMLInputElement, InputProps>(
  function Input(
    {
      inputSize = "md",
      error,
      label,
      helperText,
      leftIcon,
      rightIcon,
      className = "",
      disabled,
      id: externalId,
      ...props
    },
    ref
  ) {
    const generatedId = useId();
    const inputId = externalId || generatedId;
    const errorId = `${inputId}-error`;
    const helperId = `${inputId}-helper`;

    const baseStyles = "w-full rounded-lg border bg-surface transition-colors focus:outline-none focus:ring-2 focus:ring-primary/50";
    const stateStyles = error
      ? "border-danger focus:border-danger"
      : "border-border focus:border-primary";
    const iconPadding = leftIcon ? "pl-10" : "";
    const rightIconPadding = rightIcon ? "pr-10" : "";

    return (
      <div className="w-full">
        {label && (
          <label htmlFor={inputId} className="block text-sm font-medium text-foreground mb-1">
            {label}
          </label>
        )}
        <div className="relative">
          {leftIcon && (
            <div className="absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none">
              {leftIcon}
            </div>
          )}
          <input
            ref={ref}
            id={inputId}
            aria-invalid={!!error}
            aria-describedby={error ? errorId : helperText ? helperId : undefined}
            className={`${baseStyles} ${sizeStyles[inputSize]} ${stateStyles} ${iconPadding} ${rightIconPadding} ${disabled ? "opacity-50 cursor-not-allowed" : ""} ${className}`}
            disabled={disabled}
            {...props}
          />
          {rightIcon && (
            <div className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground pointer-events-none">
              {rightIcon}
            </div>
          )}
        </div>
        {error ? (
          <p id={errorId} className="mt-1 text-sm text-danger" role="alert">{error}</p>
        ) : helperText ? (
          <p id={helperId} className="mt-1 text-sm text-muted-foreground">{helperText}</p>
        ) : null}
      </div>
    );
  }
));

Input.displayName = "Input";
