export interface RefineIteration {
  id: string;
  iteration: number;
  status: 'pending' | 'running' | 'completed' | 'failed';
  imageUrl?: string;
  instructions?: string;
  timestamp?: string;
}

export interface IterationTimelineProps {
  iterations?: RefineIteration[];
  currentIteration?: number;
  className?: string;
}

export function IterationTimeline({ iterations = [], currentIteration = 0, className = '' }: IterationTimelineProps) {
  return (
    <div className={`bg-card rounded-xl border border-border/30 p-6 ${className}`}>
      <div className="mb-4">
        <h4 className="text-lg font-semibold text-foreground">Refinement Iterations</h4>
      </div>
      <div className="space-y-2">
        {iterations.map((iteration) => (
          <div
            key={iteration.id}
            className={`flex items-center gap-3 px-3 py-2 rounded-lg text-sm ${iteration.iteration === currentIteration ? 'bg-primary/10 text-primary' : 'text-foreground hover:bg-muted/40'} ${iteration.status === 'completed' ? 'border-l-2 border-status-success' : iteration.status === 'failed' ? 'border-l-2 border-status-error' : iteration.status === 'running' ? 'border-l-2 border-status-running' : ''}`}
          >
            <span className="font-mono text-xs text-muted-foreground">#{iteration.iteration}</span>
            <span className="flex-1">{iteration.instructions || `Iteration ${iteration.iteration}`}</span>
            <span className={`text-xs px-2 py-0.5 rounded-full ${iteration.status === 'completed' ? 'bg-status-success/10 text-status-success' : iteration.status === 'failed' ? 'bg-status-error/10 text-status-error' : iteration.status === 'running' ? 'bg-status-running/10 text-status-running' : 'bg-muted text-muted-foreground'}`}>
              {iteration.status}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
