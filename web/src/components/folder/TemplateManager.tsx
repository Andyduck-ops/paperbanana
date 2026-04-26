import { useState } from 'react';
import { useLanguage } from '../../hooks';
import type { Template } from '../../hooks/useTemplates';

export interface TemplateManagerProps {
  templates: Template[];
  isLoading: boolean;
  onCreate: (template: { name: string; description?: string; category: string; content: string }) => void;
  onUpdate: (id: string, template: { name?: string; description?: string; category?: string; content?: string }) => void;
  onDelete: (id: string) => void;
}

export function TemplateManager({
  templates,
  isLoading,
  onCreate,
  onUpdate,
  onDelete,
}: TemplateManagerProps) {
  const { t } = useLanguage();
  const [isEditing, setIsEditing] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<Template | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    category: 'general',
    content: '',
  });
  const [selectedCategory, setSelectedCategory] = useState<string>('all');

  const categories = ['all', ...Array.from(new Set(templates.map((t) => t.category)))];

  const filteredTemplates = selectedCategory === 'all'
    ? templates
    : templates.filter((t) => t.category === selectedCategory);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (editingTemplate) {
      onUpdate(editingTemplate.id, formData);
    } else {
      onCreate(formData);
    }
    resetForm();
  };

  const resetForm = () => {
    setIsEditing(false);
    setEditingTemplate(null);
    setFormData({ name: '', description: '', category: 'general', content: '' });
  };

  const handleEdit = (template: Template) => {
    setEditingTemplate(template);
    setFormData({
      name: template.name,
      description: template.description || '',
      category: template.category,
      content: '',
    });
    setIsEditing(true);
  };

  const handleDelete = (id: string) => {
    if (confirm(t('template.confirmDelete') || 'Are you sure you want to delete this template?')) {
      onDelete(id);
    }
  };

  if (isLoading) {
    return (
      <div className="template-manager-loading">
        <div className="w-6 h-6 border-2 border-primary/25 border-t-primary rounded-full animate-spin" />
        <span>{t('template.loading') || 'Loading templates...'}</span>
      </div>
    );
  }

  return (
    <div className="template-manager">
      <div className="template-manager-header">
        <h3 className="template-manager-title">{t('template.title') || 'Template Management'}</h3>
        <button
          className="template-manager-btn template-manager-btn--primary"
          onClick={() => setIsEditing(true)}
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
          {t('template.new') || 'New Template'}
        </button>
      </div>

      {/* Category Filter */}
      <div className="template-manager-filters">
        {categories.map((cat) => (
          <button
            key={cat}
            className={`template-manager-filter ${selectedCategory === cat ? 'template-manager-filter--active' : ''}`}
            onClick={() => setSelectedCategory(cat)}
          >
            {cat === 'all' ? t('template.allCategories') || 'All' : cat}
          </button>
        ))}
      </div>

      {/* Edit/Create Form */}
      {isEditing && (
        <div className="template-manager-modal">
          <div className="template-manager-modal-content">
            <div className="template-manager-modal-header">
              <h4>{editingTemplate ? t('template.edit') || 'Edit Template' : t('template.create') || 'Create Template'}</h4>
              <button className="template-manager-modal-close" onClick={resetForm}>
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <form onSubmit={handleSubmit} className="template-manager-form">
              <div className="template-manager-field">
                <label>{t('template.name') || 'Name'}</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                />
              </div>
              <div className="template-manager-field">
                <label>{t('template.category') || 'Category'}</label>
                <input
                  type="text"
                  value={formData.category}
                  onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                  required
                />
              </div>
              <div className="template-manager-field">
                <label>{t('template.description') || 'Description'}</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  rows={2}
                />
              </div>
              <div className="template-manager-field">
                <label>{t('template.content') || 'Content'}</label>
                <textarea
                  value={formData.content}
                  onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                  rows={6}
                  required
                />
              </div>
              <div className="template-manager-actions">
                <button type="button" className="template-manager-btn" onClick={resetForm}>
                  {t('common.cancel') || 'Cancel'}
                </button>
                <button type="submit" className="template-manager-btn template-manager-btn--primary">
                  {editingTemplate ? t('common.save') || 'Save' : t('common.create') || 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Templates List */}
      <div className="template-manager-list">
        {filteredTemplates.length === 0 ? (
          <div className="template-manager-empty">
            <svg className="w-12 h-12 text-muted-foreground/50" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
            </svg>
            <p>{t('template.noTemplates') || 'No templates yet'}</p>
          </div>
        ) : (
          filteredTemplates.map((template) => (
            <div key={template.id} className="template-manager-item">
              <div className="template-manager-item-info">
                <h4 className="template-manager-item-name">{template.name}</h4>
                <span className="template-manager-item-category">{template.category}</span>
                {template.description && (
                  <p className="template-manager-item-description">{template.description}</p>
                )}
              </div>
              <div className="template-manager-item-actions">
                <button
                  className="template-manager-item-btn"
                  onClick={() => handleEdit(template)}
                  title={t('common.edit') || 'Edit'}
                >
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                  </svg>
                </button>
                <button
                  className="template-manager-item-btn template-manager-item-btn--danger"
                  onClick={() => handleDelete(template.id)}
                  title={t('common.delete') || 'Delete'}
                >
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12.562 3.034a2.25 2.25 0 01-1.52-2.228m0 0a48.394 48.394 0 013.36-.379m12.562 3.034a2.25 2.25 0 011.52-2.228m0 0a48.394 48.394 0 013.36-.379m0 0a1.5 1.5 0 01-1.484-1.475v-.993a1.5 1.5 0 011.484-1.475" />
                  </svg>
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
