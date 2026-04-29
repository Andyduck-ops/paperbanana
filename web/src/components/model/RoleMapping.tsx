import type { Provider, RoleAssignment, WorkflowRole } from '../../stores';

export interface RoleMappingProps {
  channels?: Provider[];
  roles?: Array<{ id: string; name: string; modelId?: string }>;
  assignments?: Record<WorkflowRole, RoleAssignment | null>;
  onRoleUpdate?: (roleId: string, modelId: string) => void;
  onAssign?: (role: WorkflowRole, provider_id: string, model_id: string) => Promise<void>;
  onClear?: (role: WorkflowRole) => void;
  className?: string;
}

export function RoleMapping({
  channels: _channels = [],
  roles = [],
  assignments = {} as Record<WorkflowRole, RoleAssignment | null>,
  onRoleUpdate: _onRoleUpdate,
  onAssign: _onAssign,
  onClear,
  className = ''
}: RoleMappingProps) {
  // Build role list from assignments if provided, otherwise fall back to roles prop
  const roleList = Object.keys(assignments).length > 0
    ? Object.entries(assignments).map(([key, value]) => ({
        id: key,
        name: key === 'image_generation' ? 'Image Generation' :
              key === 'retrieval_reasoning' ? 'Retrieval & Reasoning' : key,
        modelId: value?.model_id,
        providerId: value?.provider_id,
      }))
    : roles;

  const handleClear = (roleId: string) => {
    if (onClear) {
      onClear(roleId as WorkflowRole);
    }
  };

  return (
    <div className={`bg-card rounded-xl border border-border/30 p-6 ${className}`}>
      <div className="mb-4">
        <h3 className="text-lg font-semibold text-foreground">Role Mappings</h3>
      </div>
      <div className="space-y-2">
        {roleList.map((role) => (
          <div key={role.id} className="flex items-center justify-between py-2 px-3 rounded-lg hover:bg-muted/40 transition-colors">
            <span className="text-sm font-medium text-foreground">{role.name}</span>
            <div className="flex items-center gap-2">
              <span className="text-xs text-muted-foreground">{role.modelId || 'Unassigned'}</span>
              {role.modelId && onClear && (
                <button
                  onClick={() => handleClear(role.id)}
                  type="button"
                  className="w-6 h-6 flex items-center justify-center rounded-md text-muted-foreground hover:text-status-error hover:bg-status-error/10 transition-colors"
                  title="Clear assignment"
                >
                  ×
                </button>
              )}
            </div>
          </div>
        ))}
        {roleList.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            <p className="text-sm">No roles defined</p>
          </div>
        )}
      </div>
    </div>
  );
}
