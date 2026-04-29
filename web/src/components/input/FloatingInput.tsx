import { useState, forwardRef, InputHTMLAttributes } from 'react';

export interface FloatingInputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'placeholder'> {
  label: string;
  error?: string;
}

export const FloatingInput = forwardRef<HTMLInputElement, FloatingInputProps>(
  ({ label, error, className = '', id, ...props }, ref) => {
    const [isFocused, setIsFocused] = useState(false);
    const inputId = id || `floating-input-${label.toLowerCase().replace(/\s+/g, '-')}`;

    return (
      <div className={`relative ${isFocused ? 'focused' : ''} ${error ? 'error' : ''} ${className}`}>
        <input
          ref={ref}
          id={inputId}
          placeholder=" "
          onFocus={() => setIsFocused(true)}
          onBlur={() => setIsFocused(false)}
          className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary peer"
          {...props}
        />
        <label
          htmlFor={inputId}
          className="absolute left-3 top-2 text-sm text-muted-foreground transition-all peer-focus:-translate-y-5 peer-focus:text-xs peer-focus:text-primary peer-[:not(:placeholder-shown)]:-translate-y-5 peer-[:not(:placeholder-shown)]:text-xs"
        >
          {label}
        </label>
        {error && <span className="block mt-1 text-xs text-status-error">{error}</span>}
      </div>
    );
  }
);

FloatingInput.displayName = 'FloatingInput';
