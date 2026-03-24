import { useState } from 'react';
import { useLanguage } from '../hooks';
import { ImageUpload } from './ImageUpload';

export interface DualInputPanelProps {
  methodContent: string;
  caption: string;
  onMethodChange: (value: string) => void;
  onCaptionChange: (value: string) => void;
  referenceImageData?: string | null;
  onReferenceImageChange?: (value: string) => void;
  onReferenceImageClear?: () => void;
  disabled?: boolean;
  examples?: { method: string; caption: string }[];
}

export function DualInputPanel({
  methodContent,
  caption,
  onMethodChange,
  onCaptionChange,
  referenceImageData = null,
  onReferenceImageChange,
  onReferenceImageClear,
  disabled = false,
  examples = [],
}: DualInputPanelProps) {
  const { t } = useLanguage();
  const [showMethodPreview, setShowMethodPreview] = useState(false);
  const [showCaptionPreview, setShowCaptionPreview] = useState(false);

  const textareaClasses =
    'w-full px-4 py-3 rounded-2xl border border-border bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary resize-none';

  return (
    <div className="space-y-4">
      {examples.length > 0 && (
        <select
          className="w-full rounded-2xl border border-border bg-background px-3 py-2 text-sm text-foreground"
          onChange={(e) => {
            const idx = parseInt(e.target.value, 10);
            if (!isNaN(idx) && examples[idx]) {
              onMethodChange(examples[idx].method);
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
      )}

      <div className="dual-input-panel grid grid-cols-1 gap-4 md:grid-cols-5">
      {/* Method Section - 3/5 width */}
      <div className="dual-input-panel__context md:col-span-3">
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

        {onReferenceImageChange && (
          <div className="mt-4 space-y-2">
            <label className="block text-sm font-medium text-foreground">
              {t('generate.referenceImage')}
            </label>
            <ImageUpload
              onImageSelect={onReferenceImageChange}
              onClear={onReferenceImageClear}
              disabled={disabled}
              initialPreview={referenceImageData}
              className="min-h-[9rem] rounded-2xl"
            />
          </div>
        )}
      </div>

      {/* Figure Caption - 2/5 width */}
      <div className="dual-input-panel__brief md:col-span-2">
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
    </div>
  );
}
