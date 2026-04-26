import { forwardRef, memo, useId, type SelectHTMLAttributes } from "react";

export type SelectSize = "sm" | "md" | "lg";

export interface SelectOption {
  value: string;
  label: string;
  disabled?: boolean;
}

export interface SelectProps extends Omit<SelectHTMLAttributes<HTMLSelectElement>, 'size'> {
  selectSize?: SelectSize;
  error?: string;
  label?: string;
  helperText?: string;
  options: SelectOption[];
  placeholder?: string;
}

const sizeStyles: Record<SelectSize, string> = {
  sm: "h-8 px-3 text-sm",
  md: "h-10 px-4 text-base",
  lg: "h-12 px-4 text-lg",
};

export const Select = memo(forwardRef<HTMLSelectElement, SelectProps>(
  function Select(
    {
      selectSize = "md",
      error,
      label,
      helperText,
      options,
      placeholder,
      className = "",
      disabled,
      id: externalId,
      ...props
    },
    ref
  ) {
    const generatedId = useId();
    const selectId = externalId || generatedId;
    const errorId = `${selectId}-error`;
    const helperId = `${selectId}-helper`;

    const baseStyles = "w-full rounded-lg border bg-surface transition-colors focus:outline-none focus:ring-2 focus:ring-primary/50 appearance-none cursor-pointer";
    const stateStyles = error
      ? "border-danger focus:border-danger"
      : "border-border focus:border-primary";

    return (
      <div className="w-full">
        {label && (
          <label htmlFor={selectId} className="block text-sm font-medium text-foreground mb-1">
            {label}
          </label>
        )}
        <div className="relative">
          <select
            ref={ref}
            id={selectId}
            aria-invalid={!!error}
            aria-describedby={error ? errorId : helperText ? helperId : undefined}
            className={`${baseStyles} ${sizeStyles[selectSize]} ${stateStyles} pr-10 ${disabled ? "opacity-50 cursor-not-allowed" : ""} ${className}`}
            disabled={disabled}
            {...props}
          >
            {placeholder && (
              <option value="" disabled>
                {placeholder}
              </option>
            )}
            {options.map((option) => (
              <option key={option.value} value={option.value} disabled={option.disabled}>
                {option.label}
              </option>
            ))}
          </select>
          <div className="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none text-muted-foreground">
            <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M2.5 4.5L6 8L9.5 4.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </div>
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

Select.displayName = "Select";
