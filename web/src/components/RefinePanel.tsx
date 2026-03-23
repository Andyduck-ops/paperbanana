import { useState, useCallback } from 'react';
import { useLanguage } from '../hooks';
import { ImageUpload } from './ImageUpload';

export interface RefineRequest {
  imageData: string;
  instructions: string;
  resolution: '2K' | '4K';
}

export interface RefinePanelProps {
  onRefine: (request: RefineRequest) => void;
  isRefining?: boolean;
  apiBase?: string;
}

export function RefinePanel({
  onRefine,
  isRefining = false,
}: RefinePanelProps) {
  const { t } = useLanguage();
  const [imageData, setImageData] = useState<string | null>(null);
  const [instructions, setInstructions] = useState('');
  const [resolution, setResolution] = useState<'2K' | '4K'>('2K');

  const handleImageSelect = useCallback((base64Data: string) => {
    setImageData(base64Data);
  }, []);

  const handleSubmit = useCallback((e: React.FormEvent) => {
    e.preventDefault();
    if (imageData) {
      onRefine({
        imageData,
        instructions: instructions.trim(),
        resolution,
      });
    }
  }, [imageData, instructions, resolution, onRefine]);

  const canSubmit = imageData && !isRefining;

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <h2 className="text-xl font-semibold text-foreground">
        {t('refine.title')}
      </h2>

      <ImageUpload
        onImageSelect={handleImageSelect}
        disabled={isRefining}
      />

      <div>
        <label htmlFor="refine-instructions" className="block text-sm font-medium text-foreground mb-2">
          {t('refine.instructions')}
        </label>
        <textarea
          id="refine-instructions"
          value={instructions}
          onChange={(e) => setInstructions(e.target.value)}
          placeholder={t('refine.instructionsPlaceholder')}
          disabled={isRefining}
          rows={4}
          className="w-full px-4 py-2 rounded-lg border border-border bg-background text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary resize-none"
        />
      </div>

      <div>
        <label className="block text-sm font-medium text-foreground mb-2">
          {t('refine.resolution')}
        </label>
        <div className="flex gap-4">
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="radio"
              name="resolution"
              value="2K"
              checked={resolution === '2K'}
              onChange={() => setResolution('2K')}
              disabled={isRefining}
              className="w-4 h-4 border-border text-primary focus:ring-primary"
            />
            <span className="text-sm text-foreground">{t('refine.resolution2K')}</span>
          </label>
          <label className="flex items-center gap-2 cursor-pointer">
            <input
              type="radio"
              name="resolution"
              value="4K"
              checked={resolution === '4K'}
              onChange={() => setResolution('4K')}
              disabled={isRefining}
              className="w-4 h-4 border-border text-primary focus:ring-primary"
            />
            <span className="text-sm text-foreground">{t('refine.resolution4K')}</span>
          </label>
        </div>
      </div>

      <button
        type="submit"
        disabled={!canSubmit}
        className="w-full px-6 py-3 rounded-lg bg-primary text-background font-medium hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed transition-opacity"
      >
        {isRefining ? t('refine.refining') : t('refine.refineButton')}
      </button>
    </form>
  );
}
