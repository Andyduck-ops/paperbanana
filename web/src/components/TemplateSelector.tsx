import { useState, useRef, useEffect } from 'react';
import { useLanguage } from '../hooks';
import type { PromptTemplate } from '../hooks/usePromptTemplates';

export interface TemplateSelectorProps {
  templates: PromptTemplate[];
  onSelect: (template: PromptTemplate) => void;
  onSave?: (name: string) => void;
  onDelete?: (id: string) => void;
  disabled?: boolean;
  currentMethodContent?: string;
  currentCaption?: string;
}

export function TemplateSelector({
  templates,
  onSelect,
  onSave,
  onDelete,
  disabled = false,
  currentMethodContent = '',
  currentCaption = '',
}: TemplateSelectorProps) {
  const { t } = useLanguage();
  const [isOpen, setIsOpen] = useState(false);
  const [showSaveDialog, setShowSaveDialog] = useState(false);
  const [templateName, setTemplateName] = useState('');
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleSave = () => {
    if (templateName.trim() && onSave) {
      onSave(templateName.trim());
      setTemplateName('');
      setShowSaveDialog(false);
    }
  };

  const canSave = currentMethodContent.trim() || currentCaption.trim();

  return (
    <div ref={dropdownRef} className="relative">
      {/* Trigger Button */}
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        disabled={disabled}
        className="flex items-center gap-2 px-3 py-1.5 text-sm text-muted-foreground hover:text-foreground 
                   bg-muted/50 hover:bg-muted rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        title={t('templates.select')}
      >
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
        </svg>
        <span>{t('templates.select')}</span>
        <svg className={`w-3 h-3 transition-transform ${isOpen ? 'rotate-180' : ''}`} 
             fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {/* Dropdown */}
      {isOpen && (
        <div className="absolute bottom-full left-0 mb-2 w-80 bg-popover border border-border rounded-lg 
                        shadow-lg z-50 overflow-hidden">
          {/* Header */}
          <div className="flex items-center justify-between px-3 py-2 border-b border-border bg-muted/50">
            <span className="text-sm font-medium">{t('templates.title')}</span>
            {canSave && onSave && (
              <button
                type="button"
                onClick={() => setShowSaveDialog(true)}
                className="text-xs text-primary hover:text-primary/80 flex items-center gap-1"
              >
                <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
                {t('templates.save')}
              </button>
            )}
          </div>

          {/* Template List */}
          <div className="max-h-60 overflow-y-auto">
            {templates.length === 0 ? (
              <div className="px-3 py-4 text-center text-sm text-muted-foreground">
                <p>{t('templates.empty')}</p>
                {canSave && (
                  <p className="text-xs mt-1">{t('templates.createHint')}</p>
                )}
              </div>
            ) : (
              templates.map((template) => (
                <div
                  key={template.id}
                  className="group flex items-center justify-between px-3 py-2 hover:bg-muted 
                           cursor-pointer border-b border-border/50 last:border-b-0"
                >
                  <button
                    type="button"
                    onClick={() => {
                      onSelect(template);
                      setIsOpen(false);
                    }}
                    className="flex-1 text-left"
                  >
                    <div className="font-medium text-sm truncate">{template.name}</div>
                    <div className="text-xs text-muted-foreground truncate">
                      {template.caption.slice(0, 40) || template.methodContent.slice(0, 40)}...
                    </div>
                  </button>
                  {onDelete && (
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        onDelete(template.id);
                      }}
                      className="opacity-0 group-hover:opacity-100 p-1 text-muted-foreground 
                               hover:text-destructive transition-opacity"
                      title={t('templates.delete')}
                    >
                      <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} 
                              d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  )}
                </div>
              ))
            )}
          </div>
        </div>
      )}

      {/* Save Dialog */}
      {showSaveDialog && (
        <div className="absolute bottom-full left-0 mb-2 w-80 bg-popover border border-border rounded-lg 
                      shadow-lg z-50 p-3">
          <div className="text-sm font-medium mb-2">{t('templates.save')}</div>
          <input
            type="text"
            value={templateName}
            onChange={(e) => setTemplateName(e.target.value)}
            placeholder={t('templates.namePlaceholder')}
            className="w-full px-3 py-2 text-sm border border-border rounded-md bg-background 
                     focus:outline-none focus:ring-2 focus:ring-primary mb-3"
            autoFocus
          />
          <div className="flex justify-end gap-2">
            <button
              type="button"
              onClick={() => {
                setShowSaveDialog(false);
                setTemplateName('');
              }}
              className="px-3 py-1.5 text-xs text-muted-foreground hover:text-foreground"
            >
              {t('common.cancel')}
            </button>
            <button
              type="button"
              onClick={handleSave}
              disabled={!templateName.trim()}
              className="px-3 py-1.5 text-xs bg-primary text-primary-foreground rounded-md 
                       hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {t('common.save')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
