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
    <div className={`role-mapping ${className}`}>
      <div className="role-mapping__header">
        <h3>Role Mappings</h3>
      </div>
      <div className="role-mapping__list">
        {roleList.map((role) => (
          <div key={role.id} className="role-mapping__item">
            <span className="role-mapping__role-name">{role.name}</span>
            <div className="role-mapping__actions">
              <span className="role-mapping__model-id">{role.modelId || 'Unassigned'}</span>
              {role.modelId && onClear && (
                <button 
                  onClick={() => handleClear(role.id)}
                  type="button"
                  className="role-mapping__clear-btn"
                  title="Clear assignment"
                >
                  ×
                </button>
              )}
            </div>
          </div>
        ))}
        {roleList.length === 0 && (
          <div className="role-mapping__empty">
            <p>No roles defined</p>
          </div>
        )}
      </div>
    </div>
  );
}
