import { useState } from 'react';
import { useLanguage } from '../hooks';
import { Button } from './atoms';
import { exportAsPng, exportAsSvg, exportAsPdf } from '../lib/export';

export type ExportFormat = 'png' | 'svg' | 'pdf';

export interface ExportModalProps {
  isOpen: boolean;
  onClose: () => void;
  imageData?: string;
  svgElement?: SVGSVGElement;
  canvas?: HTMLCanvasElement;
}

export function ExportModal({
  isOpen,
  onClose,
  imageData,
  svgElement,
  canvas,
}: ExportModalProps) {
  const { t } = useLanguage();
  const [format, setFormat] = useState<ExportFormat>('png');
  const [dpi, setDpi] = useState(300);
  const [isExporting, setIsExporting] = useState(false);

  if (!isOpen) return null;

  const handleExport = async () => {
    setIsExporting(true);
    try {
      if (format === 'png') {
        if (canvas) {
          await exportAsPng(canvas, { dpi });
        } else if (imageData) {
          const img = new window.Image();
          img.src = `data:image/png;base64,${imageData}`;
          await new Promise((resolve) => { img.onload = resolve; });
          await exportAsPng(img, { dpi });
        }
      } else if (format === 'svg' && svgElement) {
        await exportAsSvg(svgElement);
      } else if (format === 'pdf') {
        if (canvas) {
          await exportAsPdf(canvas, { dpi });
        } else if (imageData) {
          const img = new window.Image();
          img.src = `data:image/png;base64,${imageData}`;
          await new Promise((resolve) => { img.onload = resolve; });
          const tempCanvas = document.createElement('canvas');
          tempCanvas.width = img.naturalWidth;
          tempCanvas.height = img.naturalHeight;
          const ctx = tempCanvas.getContext('2d');
          if (ctx) {
            ctx.drawImage(img, 0, 0);
            await exportAsPdf(tempCanvas, { dpi });
          }
        }
      }
      onClose();
    } finally {
      setIsExporting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-background border border-border rounded-lg p-6 w-full max-w-sm">
        <h3 className="text-lg font-heading text-foreground mb-4">
          {t('export.title')}
        </h3>

        <div className="space-y-4">
          <div>
            <label className="block text-sm text-foreground mb-2">
              {t('export.format')}
            </label>
            <div className="flex gap-2">
              {(['png', 'svg', 'pdf'] as ExportFormat[]).map((f) => (
                <button
                  key={f}
                  onClick={() => setFormat(f)}
                  className={`
                    px-4 py-2 rounded border transition-colors uppercase
                    ${format === f
                      ? 'bg-primary text-background border-primary'
                      : 'bg-background text-foreground border-border hover:bg-muted'
                    }
                  `}
                >
                  {f}
                </button>
              ))}
            </div>
          </div>

          {format !== 'svg' && (
            <div>
              <label className="block text-sm text-foreground mb-2">
                {t('export.dpi')}
              </label>
              <input
                type="number"
                value={dpi}
                onChange={(e) => setDpi(Math.max(72, Math.min(600, parseInt(e.target.value) || 300)))}
                min={72}
                max={600}
                className="w-full px-3 py-2 rounded border border-border bg-background text-foreground"
              />
              <p className="text-xs text-muted-foreground mt-1">
                {t('export.dpiRange')} (72-600)
              </p>
            </div>
          )}

          <div className="flex gap-2 pt-4">
            <Button
              type="button"
              variant="secondary"
              fullWidth
              onClick={onClose}
            >
              {t('common.cancel')}
            </Button>
            <Button
              type="button"
              variant="primary"
              fullWidth
              isLoading={isExporting}
              onClick={handleExport}
            >
              {t('export.download')}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
