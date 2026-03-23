package agent

import "context"

// BaseAgent defines the lifecycle every pipeline stage must support.
type BaseAgent interface {
	Initialize(ctx context.Context) error
	Execute(ctx context.Context, input AgentInput) (AgentOutput, error)
	Cleanup(ctx context.Context) error
	GetState() AgentState
	RestoreState(state AgentState) error
}
