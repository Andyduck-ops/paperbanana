import { ReactNode } from 'react';

export interface ModelPopoverProps {
  trigger: ReactNode;
  children: ReactNode;
  isOpen?: boolean;
  onOpenChange?: (open: boolean) => void;
  className?: string;
}

export function ModelPopover({ trigger, children, isOpen = false, className = '' }: ModelPopoverProps) {
  return (
    <div className={`relative ${className}`}>
      <div>{trigger}</div>
      {isOpen && (
        <div className="absolute z-50 mt-1 w-64 bg-card border border-border rounded-xl shadow-lg p-3">
          {children}
        </div>
      )}
    </div>
  );
}
