import { useState } from 'react';
import { useLanguage } from '../hooks';

export interface DualInputPanelProps {
  methodContent: string;
  caption: string;
  onMethodChange: (value: string) => void;
  onCaptionChange: (value: string) => void;
  disabled?: boolean;
  examples?: { method: string; caption: string }[];
}

export function DualInputPanel({
  methodContent,
  caption,
  onMethodChange,
  onCaptionChange,
  disabled = false,
  examples = [],
}: DualInputPanelProps) {
  const { t } = useLanguage();
  const [showMethodPreview, setShowMethodPreview] = useState(false);
  const [showCaptionPreview, setShowCaptionPreview] = useState(false);

  const textareaClasses =
    'w-full px-4 py-3 rounded-lg border border-border bg-background text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary resize-none';

  return (
    <div className="grid grid-cols-5 gap-4">
      {/* Method Section - 3/5 width */}
      <div className="col-span-3">
        <div className="flex items-center justify-between mb-2">
          <label htmlFor="method-content" className="block text-sm font-medium text-foreground">
            {t('generate.methodSection')}
          </label>
          <button
            type="button"
            onClick={() => setShowMethodPreview(!showMethodPreview)}
            className="text-xs text-muted-foreground hover:text-foreground"
            disabled={disabled}
          >
            {t('generate.previewMarkdown')}
          </button>
        </div>
        {examples.length > 0 && (
          <div className="mb-2">
            <select
              className="w-full px-2 py-1 text-sm rounded border border-border bg-background text-foreground"
              onChange={(e) => {
                const idx = parseInt(e.target.value, 10);
                if (!isNaN(idx) && examples[idx]) {
                  onMethodChange(examples[idx].method);
                }
              }}
              disabled={disabled}
            >
              <option value="">{t('generate.loadExample')}</option>
              {examples.map((_, i) => (
                <option key={i} value={i}>
                  Example {i + 1}
                </option>
              ))}
            </select>
          </div>
        )}
        {showMethodPreview ? (
          <div className="prose prose-sm dark:prose-invert max-w-none min-h-[100px] p-4 rounded-lg border border-border bg-background">
            {methodContent || <span className="text-muted-foreground">No content</span>}
          </div>
        ) : (
          <textarea
            id="method-content"
            value={methodContent}
            onChange={(e) => onMethodChange(e.target.value)}
            placeholder={t('generate.methodPlaceholder')}
            disabled={disabled}
            rows={6}
            className={textareaClasses}
          />
        )}
      </div>

      {/* Figure Caption - 2/5 width */}
      <div className="col-span-2">
        <div className="flex items-center justify-between mb-2">
          <label htmlFor="figure-caption" className="block text-sm font-medium text-foreground">
            {t('generate.figureCaption')}
          </label>
          <button
            type="button"
            onClick={() => setShowCaptionPreview(!showCaptionPreview)}
            className="text-xs text-muted-foreground hover:text-foreground"
            disabled={disabled}
          >
            {t('generate.previewMarkdown')}
          </button>
        </div>
        {examples.length > 0 && (
          <div className="mb-2">
            <select
              className="w-full px-2 py-1 text-sm rounded border border-border bg-background text-foreground"
              onChange={(e) => {
                const idx = parseInt(e.target.value, 10);
                if (!isNaN(idx) && examples[idx]) {
                  onCaptionChange(examples[idx].caption);
                }
              }}
              disabled={disabled}
            >
              <option value="">{t('generate.loadExample')}</option>
              {examples.map((_, i) => (
                <option key={i} value={i}>
                  Example {i + 1}
                </option>
              ))}
            </select>
          </div>
        )}
        {showCaptionPreview ? (
          <div className="prose prose-sm dark:prose-invert max-w-none min-h-[100px] p-4 rounded-lg border border-border bg-background">
            {caption || <span className="text-muted-foreground">No content</span>}
          </div>
        ) : (
          <textarea
            id="figure-caption"
            value={caption}
            onChange={(e) => onCaptionChange(e.target.value)}
            placeholder={t('generate.captionPlaceholder')}
            disabled={disabled}
            rows={6}
            className={textareaClasses}
          />
        )}
      </div>
    </div>
  );
}
