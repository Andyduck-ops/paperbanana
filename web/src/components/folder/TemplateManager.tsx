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
      <div className="flex items-center gap-2 py-8 text-muted-foreground">
        <div className="w-6 h-6 border-2 border-primary/25 border-t-primary rounded-full animate-spin" />
        <span>{t('template.loading') || 'Loading templates...'}</span>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-foreground">{t('template.title') || 'Template Management'}</h3>
        <button
          className="inline-flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-background bg-primary rounded-lg hover:opacity-90 transition-opacity"
          onClick={() => setIsEditing(true)}
        >
          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
          {t('template.new') || 'New Template'}
        </button>
      </div>

      {/* Category Filter */}
      <div className="flex flex-wrap gap-2">
        {categories.map((cat) => (
          <button
            key={cat}
            className={`px-3 py-1 text-xs font-medium rounded-full transition-colors ${selectedCategory === cat ? 'bg-primary text-background' : 'bg-muted text-muted-foreground hover:text-foreground'}`}
            onClick={() => setSelectedCategory(cat)}
          >
            {cat === 'all' ? t('template.allCategories') || 'All' : cat}
          </button>
        ))}
      </div>

      {/* Edit/Create Form */}
      {isEditing && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div className="fixed inset-0 bg-black/50 backdrop-blur-sm" onClick={resetForm} />
          <div className="relative bg-card rounded-2xl border border-border shadow-2xl p-6 w-full max-w-lg mx-4 max-h-[90vh] overflow-auto">
            <div className="flex items-center justify-between mb-4">
              <h4 className="text-lg font-semibold text-foreground">{editingTemplate ? t('template.edit') || 'Edit Template' : t('template.create') || 'Create Template'}</h4>
              <button className="w-8 h-8 flex items-center justify-center rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-colors" onClick={resetForm}>
                <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-1.5">
                <label className="block text-sm font-medium text-foreground">{t('template.name') || 'Name'}</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>
              <div className="space-y-1.5">
                <label className="block text-sm font-medium text-foreground">{t('template.category') || 'Category'}</label>
                <input
                  type="text"
                  value={formData.category}
                  onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                  required
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                />
              </div>
              <div className="space-y-1.5">
                <label className="block text-sm font-medium text-foreground">{t('template.description') || 'Description'}</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  rows={2}
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary resize-none"
                />
              </div>
              <div className="space-y-1.5">
                <label className="block text-sm font-medium text-foreground">{t('template.content') || 'Content'}</label>
                <textarea
                  value={formData.content}
                  onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                  rows={6}
                  required
                  className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary resize-none"
                />
              </div>
              <div className="flex items-center justify-end gap-3 pt-2">
                <button type="button" className="px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors" onClick={resetForm}>
                  {t('common.cancel') || 'Cancel'}
                </button>
                <button type="submit" className="inline-flex items-center px-4 py-2 text-sm font-medium text-background bg-primary rounded-xl hover:opacity-90 transition-opacity">
                  {editingTemplate ? t('common.save') || 'Save' : t('common.create') || 'Create'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Templates List */}
      <div className="space-y-2">
        {filteredTemplates.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 gap-3 text-muted-foreground">
            <svg className="w-12 h-12 text-muted-foreground/50" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 00-3.375-3.375h-1.5A1.125 1.125 0 0113.5 7.125v-1.5a3.375 3.375 0 00-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 00-9-9z" />
            </svg>
            <p className="text-sm">{t('template.noTemplates') || 'No templates yet'}</p>
          </div>
        ) : (
          filteredTemplates.map((template) => (
            <div key={template.id} className="flex items-start justify-between gap-3 p-3 rounded-xl border border-border/50 bg-background/50 hover:bg-muted/30 transition-colors">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <h4 className="text-sm font-semibold text-foreground">{template.name}</h4>
                  <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs bg-muted text-muted-foreground">{template.category}</span>
                </div>
                {template.description && (
                  <p className="text-xs text-muted-foreground mt-1">{template.description}</p>
                )}
              </div>
              <div className="flex items-center gap-1 shrink-0">
                <button
                  className="w-8 h-8 flex items-center justify-center rounded-lg text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
                  onClick={() => handleEdit(template)}
                  title={t('common.edit') || 'Edit'}
                >
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                  </svg>
                </button>
                <button
                  className="w-8 h-8 flex items-center justify-center rounded-lg text-muted-foreground hover:text-status-error hover:bg-status-error/10 transition-colors"
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
