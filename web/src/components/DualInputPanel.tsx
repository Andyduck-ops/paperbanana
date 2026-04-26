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
  collapsed?: boolean;
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
  collapsed = false,
}: DualInputPanelProps) {
  const { t } = useLanguage();

  const textareaClasses =
    'w-full px-4 py-3 rounded-xl border border-border bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary resize-none transition-colors';

  if (collapsed) {
    return (
      <div className="p-4 rounded-xl border border-border bg-muted/50 animate-pulse">
        <div className="flex items-center gap-3 text-muted-foreground">
          <div className="w-4 h-4 border-2 border-current border-t-transparent rounded-full animate-spin" />
          <span className="text-sm">{t('generate.generating')}</span>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Method Section */}
      <div>
        <label htmlFor="method-content" className="block text-sm font-medium text-foreground mb-1.5">
          {t('generate.methodSection')}
        </label>
        <textarea
          id="method-content"
          value={methodContent}
          onChange={(e) => onMethodChange(e.target.value)}
          placeholder={t('generate.methodPlaceholder')}
          disabled={disabled}
          rows={5}
          className={textareaClasses}
        />
      </div>

      {/* Figure Caption */}
      <div>
        <label htmlFor="figure-caption" className="block text-sm font-medium text-foreground mb-1.5">
          {t('generate.figureCaption')}
        </label>
        <textarea
          id="figure-caption"
          value={caption}
          onChange={(e) => onCaptionChange(e.target.value)}
          placeholder={t('generate.captionPlaceholder')}
          disabled={disabled}
          rows={3}
          className={textareaClasses}
        />
      </div>

      {/* Reference Image - compact */}
      {onReferenceImageChange && (
        <div>
          <label className="block text-sm font-medium text-foreground mb-1.5">
            {t('generate.referenceImage')}
          </label>
          <ImageUpload
            onImageSelect={onReferenceImageChange}
            onClear={onReferenceImageClear}
            disabled={disabled}
            initialPreview={referenceImageData}
            className="min-h-[6rem] rounded-xl"
          />
        </div>
      )}
    </div>
  );
}
