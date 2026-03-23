export interface ExportOptions {
  filename?: string;
  dpi?: number;
}

export async function exportAsPng(
  canvasOrImage: HTMLCanvasElement | HTMLImageElement,
  options: ExportOptions = {}
): Promise<void> {
  const { filename = 'visualization', dpi = 300 } = options;
  const scale = dpi / 72;

  let canvas: HTMLCanvasElement;
  if (canvasOrImage instanceof HTMLCanvasElement) {
    canvas = canvasOrImage;
  } else {
    canvas = document.createElement('canvas');
    canvas.width = canvasOrImage.naturalWidth * scale;
    canvas.height = canvasOrImage.naturalHeight * scale;
    const ctx = canvas.getContext('2d');
    if (!ctx) throw new Error('Could not get canvas context');
    ctx.scale(scale, scale);
    ctx.drawImage(canvasOrImage, 0, 0);
  }

  const blob = await new Promise<Blob>((resolve, reject) => {
    canvas.toBlob((b) => {
      if (b) resolve(b);
      else reject(new Error('Failed to create blob'));
    }, 'image/png');
  });

  downloadBlob(blob, `${filename}.png`);
}

export async function exportAsSvg(
  svgElement: SVGSVGElement,
  options: ExportOptions = {}
): Promise<void> {
  const { filename = 'visualization' } = options;

  const serializer = new XMLSerializer();
  const svgString = serializer.serializeToString(svgElement);
  const blob = new Blob([svgString], { type: 'image/svg+xml' });

  downloadBlob(blob, `${filename}.svg`);
}

export async function exportAsPdf(
  canvas: HTMLCanvasElement,
  options: ExportOptions = {}
): Promise<void> {
  const { filename = 'visualization', dpi = 300 } = options;

  // Create high-res canvas for PDF
  const scale = dpi / 72;
  const pdfCanvas = document.createElement('canvas');
  pdfCanvas.width = canvas.width * scale;
  pdfCanvas.height = canvas.height * scale;
  const ctx = pdfCanvas.getContext('2d');
  if (!ctx) throw new Error('Could not get canvas context');
  ctx.scale(scale, scale);
  ctx.drawImage(canvas, 0, 0);

  // Get PNG data and download as PDF-like file
  // Note: For proper PDF export, use a library like jspdf
  // This implementation exports as high-res PNG for now
  const blob = await new Promise<Blob>((resolve, reject) => {
    pdfCanvas.toBlob((b) => {
      if (b) resolve(b);
      else reject(new Error('Failed to create blob'));
    }, 'image/png');
  });

  // Download as PDF extension (browser will handle as image)
  // TODO: Integrate jspdf for proper PDF generation
  downloadBlob(blob, `${filename}.pdf`);
}

function downloadBlob(blob: Blob, filename: string): void {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
