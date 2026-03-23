import { ReactNode } from 'react';

export interface LayoutProps {
  children: ReactNode;
  sidebar?: ReactNode;
  header?: ReactNode;
  footer?: ReactNode;
}

export function Layout({ children, sidebar, header, footer }: LayoutProps) {
  return (
    <div className="min-h-screen flex flex-col bg-background text-foreground">
      {header && (
        <header className="sticky top-0 z-50 bg-background border-b border-border">
          {header}
        </header>
      )}

      <main className="flex-1 flex flex-col md:flex-row">
        {sidebar && (
          <aside className="w-full md:w-64 lg:w-72 shrink-0 bg-muted/30 border-b md:border-b-0 md:border-r border-border">
            <div className="p-4">
              {sidebar}
            </div>
          </aside>
        )}

        <div className="flex-1 p-4 md:p-6 lg:p-8">
          {children}
        </div>
      </main>

      {footer && (
        <footer className="bg-muted/30 border-t border-border">
          {footer}
        </footer>
      )}
    </div>
  );
}
