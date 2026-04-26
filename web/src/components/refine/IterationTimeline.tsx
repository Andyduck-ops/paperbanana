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
    <div className={`iteration-timeline ${className}`}>
      <div className="iteration-timeline__header">
        <h4>Refinement Iterations</h4>
      </div>
      <div className="iteration-timeline__list">
        {iterations.map((iteration) => (
          <div
            key={iteration.id}
            className={`iteration-timeline__item status-${iteration.status} ${iteration.iteration === currentIteration ? 'current' : ''}`}
          >
            <span className="iteration-timeline__number">#{iteration.iteration}</span>
            <span className="iteration-timeline__status">{iteration.status}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
