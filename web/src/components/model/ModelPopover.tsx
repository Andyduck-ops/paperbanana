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
    <div className={`model-popover ${className}`}>
      <div className="model-popover__trigger">{trigger}</div>
      {isOpen && (
        <div className="model-popover__content">
          {children}
        </div>
      )}
    </div>
  );
}
