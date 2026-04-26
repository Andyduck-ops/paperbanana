export interface FieldErrorProps {
  error?: string;
  className?: string;
}

export function FieldError({ error, className = '' }: FieldErrorProps) {
  if (!error) return null;

  return (
    <p className={`text-sm text-red-600 dark:text-red-400 mt-1 ${className}`} role="alert">
      {error}
    </p>
  );
}

export default FieldError;
