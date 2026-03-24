import { useLanguage } from '../../hooks';

export type EmptyStateAction = 'generate' | 'example';

export interface EmptyStateProps {
  mode: 'generate' | 'refine';
  onAction?: (action: EmptyStateAction) => void;
}

// Example prompts for different use cases
const EXAMPLE_PROMPTS = [
  {
    id: 'architecture',
    title: 'Neural Network Architecture',
    description: 'Visualize transformer attention patterns',
    prompt: 'Create a detailed diagram showing multi-head attention mechanism in transformers with query, key, value matrices and attention weights visualization.',
  },
  {
    id: 'experiment',
    title: 'Experimental Results',
    description: 'Plot comparative performance metrics',
    prompt: 'Generate a grouped bar chart comparing accuracy across 5 different models on 3 benchmark datasets with error bars.',
  },
  {
    id: 'pipeline',
    title: 'Data Pipeline',
    description: 'Illustrate processing workflow',
    prompt: 'Create a flowchart showing data preprocessing pipeline from raw input through normalization, augmentation to model training.',
  },
  {
    id: 'math',
    title: 'Mathematical Concept',
    description: 'Visualize equations and relationships',
    prompt: 'Generate a visualization of gradient descent optimization showing contour lines, optimization path, and convergence point.',
  },
];

/**
 * Empty State Component
 *
 * Displays when the workspace has no active session.
 * Provides:
 * - Brief capability statement
 * - Example prompts for quick start
 * - Visual invitation to generation mode
 *
 * Design principles:
 * - Not hollow - feels intentional and welcoming
 * - Clear entry points
 * - Establishes product tone
 */
export function EmptyState({
  mode,
  onAction,
}: EmptyStateProps) {
  useLanguage();

  const handleExampleClick = (prompt: string) => {
    // Dispatch a custom event that the parent can listen to
    const event = new CustomEvent('workspace:loadExample', {
      detail: { prompt },
    });
    window.dispatchEvent(event);
    onAction?.('example');
  };

  return (
    <div className="empty-state flex flex-col items-center justify-center min-h-[60vh] px-6">
      {/* Main illustration/icon */}
      <div className="empty-state__hero mb-8">
        <div className="relative w-32 h-32 mx-auto">
          {/* Animated rings */}
          <div className="absolute inset-0 rounded-full bg-primary/5 animate-pulse" />
          <div className="absolute inset-2 rounded-full bg-primary/10 animate-pulse" style={{ animationDelay: '0.2s' }} />
          <div className="absolute inset-4 rounded-full bg-primary/15 animate-pulse" style={{ animationDelay: '0.4s' }} />

          {/* Center icon */}
          <div className="absolute inset-6 rounded-full bg-primary/20 flex items-center justify-center">
            {mode === 'generate' ? (
              <svg
                className="w-10 h-10 text-primary"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1.5}
                  d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z"
                />
              </svg>
            ) : (
              <svg
                className="w-10 h-10 text-primary"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={1.5}
                  d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z"
                />
              </svg>
            )}
          </div>
        </div>
      </div>

      {/* Capability statement */}
      <div className="empty-state__content text-center max-w-lg mb-10">
        <h2 className="text-2xl font-heading font-semibold text-foreground mb-3">
          {mode === 'generate'
            ? 'Create Scientific Visualizations'
            : 'Refine Your Images'}
        </h2>
        <p className="text-muted-foreground leading-relaxed">
          {mode === 'generate'
            ? 'Transform your research into publication-ready figures. Describe your method, and let AI generate precise, academic-quality visualizations.'
            : 'Upload an image and provide refinement instructions to enhance resolution, adjust styling, or modify specific elements.'}
        </p>
      </div>

      {/* Example prompts - only for generate mode */}
      {mode === 'generate' && (
        <div className="empty-state__examples w-full max-w-3xl">
          <p className="text-sm text-muted-foreground text-center mb-4">
            Try an example to get started
          </p>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {EXAMPLE_PROMPTS.map((example) => (
              <button
                key={example.id}
                onClick={() => handleExampleClick(example.prompt)}
                className="
                  group text-left p-4 rounded-xl
                  border border-border/50 bg-card/50
                  hover:border-primary/30 hover:bg-primary/5
                  transition-all duration-200
                  focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/50
                "
              >
                <div className="flex items-start gap-3">
                  <div className="w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center flex-shrink-0 group-hover:bg-primary/20 transition-colors">
                    <svg
                      className="w-4 h-4 text-primary"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M13 10V3L4 14h7v7l9-11h-7z"
                      />
                    </svg>
                  </div>
                  <div>
                    <h3 className="font-medium text-foreground text-sm mb-0.5">
                      {example.title}
                    </h3>
                    <p className="text-xs text-muted-foreground">
                      {example.description}
                    </p>
                  </div>
                </div>
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Quick action hint */}
      <div className="empty-state__hint mt-8 text-center">
        <p className="text-xs text-muted-foreground">
          {mode === 'generate'
            ? 'Or type your own description below to begin'
            : 'Upload an image below to start refining'}
        </p>
      </div>
    </div>
  );
}

export default EmptyState;
