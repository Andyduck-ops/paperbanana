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
      <div className={`floating-input ${isFocused ? 'focused' : ''} ${error ? 'error' : ''} ${className}`}>
        <input
          ref={ref}
          id={inputId}
          placeholder=" "
          onFocus={() => setIsFocused(true)}
          onBlur={() => setIsFocused(false)}
          {...props}
        />
        <label htmlFor={inputId}>{label}</label>
        {error && <span className="floating-input__error">{error}</span>}
      </div>
    );
  }
);

FloatingInput.displayName = 'FloatingInput';
