import { useState, useCallback, useEffect } from 'react';
import { useLanguage } from '../hooks';
import { ImageUpload } from './ImageUpload';

export interface RefineRequest {
  imageData: string;
  instructions: string;
  resolution: '2K' | '4K';
  enableIteration?: boolean;
  maxIterations?: number;
}

export interface RefinePanelProps {
  onRefine: (request: RefineRequest) => void;
  isRefining?: boolean;
  apiBase?: string;
  initialImageData?: string | null;
}

export function RefinePanel({
  onRefine,
  isRefining = false,
  initialImageData = null,
}: RefinePanelProps) {
  const { t } = useLanguage();
  const [imageData, setImageData] = useState<string | null>(null);
  const [instructions, setInstructions] = useState('');
  const [resolution, setResolution] = useState<'2K' | '4K'>('2K');
  const [enableIteration, setEnableIteration] = useState(false);
  const [maxIterations, setMaxIterations] = useState(3);

  useEffect(() => {
    if (initialImageData) {
      setImageData(initialImageData);
    }
  }, [initialImageData]);

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
        enableIteration,
        maxIterations,
      });
    }
  }, [enableIteration, imageData, instructions, maxIterations, onRefine, resolution]);

  const canSubmit = imageData && !isRefining;

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <ImageUpload
        onImageSelect={handleImageSelect}
        onClear={() => setImageData(null)}
        disabled={isRefining}
        initialPreview={imageData}
      />

      <div>
        <label htmlFor="refine-instructions" className="mb-2 block text-sm font-medium text-foreground">
          {t('refine.instructions')}
        </label>
        <textarea
          id="refine-instructions"
          value={instructions}
          onChange={(e) => setInstructions(e.target.value)}
          placeholder={t('refine.instructionsPlaceholder')}
          disabled={isRefining}
          rows={4}
          className="w-full resize-none rounded-lg border border-border bg-background px-4 py-2 text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary"
        />
      </div>

      <div>
        <label className="mb-2 block text-sm font-medium text-foreground">
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
              className="h-4 w-4 border-border text-primary focus:ring-primary"
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
              className="h-4 w-4 border-border text-primary focus:ring-primary"
            />
            <span className="text-sm text-foreground">{t('refine.resolution4K')}</span>
          </label>
        </div>
      </div>

      <section className="space-y-4 rounded-2xl border border-border/70 bg-card/80 px-4 py-4">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-sm font-medium text-foreground">{t('refine.enableIteration')}</p>
            <p className="mt-1 text-xs text-muted-foreground">{t('refine.iterationHint')}</p>
          </div>
          <input
            aria-label={t('refine.enableIteration')}
            type="checkbox"
            checked={enableIteration}
            onChange={(event) => setEnableIteration(event.target.checked)}
            disabled={isRefining}
            className="mt-1 h-4 w-4 rounded border-border text-primary focus:ring-primary"
          />
        </div>

        <div className={enableIteration ? '' : 'opacity-50'}>
          <div className="mb-2 flex items-center justify-between gap-3">
            <label htmlFor="refine-max-iterations" className="block text-sm font-medium text-foreground">
              {t('refine.maxIterations')}
            </label>
            <span className="min-w-10 rounded-full bg-muted px-3 py-1 text-center text-sm font-medium text-foreground">
              {maxIterations}
            </span>
          </div>
          <input
            id="refine-max-iterations"
            type="range"
            min={1}
            max={5}
            step={1}
            value={maxIterations}
            onChange={(event) => setMaxIterations(Number(event.target.value))}
            disabled={!enableIteration || isRefining}
            className="w-full accent-primary"
          />
        </div>
      </section>

      <button
        type="submit"
        disabled={!canSubmit}
        className="w-full rounded-lg bg-primary px-6 py-3 font-medium text-background transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
      >
        {isRefining ? t('refine.refining') : t('refine.refineButton')}
      </button>
    </form>
  );
}
