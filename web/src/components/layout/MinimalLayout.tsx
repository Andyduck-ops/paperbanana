import { ReactNode } from 'react';

export interface MinimalLayoutProps {
  children: ReactNode;
  className?: string;
}

export function MinimalLayout({ children, className = '' }: MinimalLayoutProps) {
  return (
    <div className={`minimal-layout ${className}`}>
      {children}
    </div>
  );
}
