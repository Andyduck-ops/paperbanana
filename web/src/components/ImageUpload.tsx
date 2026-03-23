import { useState, useRef, useCallback } from 'react';
import { useLanguage } from '../hooks';

export interface ImageUploadProps {
  onImageSelect: (base64Data: string) => void;
  disabled?: boolean;
  className?: string;
}

export function ImageUpload({
  onImageSelect,
  disabled = false,
  className = '',
}: ImageUploadProps) {
  const { t } = useLanguage();
  const [isDragging, setIsDragging] = useState(false);
  const [preview, setPreview] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFile = useCallback((file: File) => {
    if (!file.type.match(/^image\/(png|jpeg|svg\+xml)$/)) {
      return;
    }

    const reader = new FileReader();
    reader.onload = (e) => {
      const base64 = e.target?.result as string;
      setPreview(base64);
      onImageSelect(base64);
    };
    reader.readAsDataURL(file);
  }, [onImageSelect]);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    if (!disabled) {
      setIsDragging(true);
    }
  }, [disabled]);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);

    if (disabled) return;

    const files = e.dataTransfer.files;
    if (files.length > 0) {
      handleFile(files[0]);
    }
  }, [disabled, handleFile]);

  const handleClick = useCallback(() => {
    if (!disabled) {
      fileInputRef.current?.click();
    }
  }, [disabled]);

  const handleFileChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      handleFile(files[0]);
    }
  }, [handleFile]);

  const handleClear = useCallback((e: React.MouseEvent) => {
    e.stopPropagation();
    setPreview(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  }, []);

  if (preview) {
    return (
      <div className={`relative ${className}`}>
        <div
          className="relative rounded-lg border border-border overflow-hidden bg-muted"
        >
          <img
            src={preview}
            alt="Uploaded preview"
            className="w-full h-auto max-h-64 object-contain"
          />
          <button
            type="button"
            onClick={handleClear}
            disabled={disabled}
            className="absolute top-2 right-2 px-3 py-1 rounded bg-red-500 text-white text-sm font-medium hover:bg-red-600 disabled:opacity-50 transition-colors"
          >
            {t('common.clear')}
          </button>
        </div>
      </div>
    );
  }

  return (
    <button
      type="button"
      onClick={handleClick}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
      disabled={disabled}
      className={`
        w-full p-8 rounded-lg border-2 border-dashed cursor-pointer
        transition-colors text-center
        ${isDragging
          ? 'border-primary bg-primary/5'
          : 'border-border hover:border-primary/50'
        }
        ${disabled ? 'opacity-50 cursor-not-allowed' : ''}
        ${className}
      `}
    >
      <input
        ref={fileInputRef}
        type="file"
        accept="image/png,image/jpeg,image/svg+xml"
        onChange={handleFileChange}
        disabled={disabled}
        className="hidden"
      />
      <div className="flex flex-col items-center gap-2">
        <svg
          className="w-12 h-12 text-muted-foreground"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={1.5}
            d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
          />
        </svg>
        <span className="text-sm text-muted-foreground">
          {t('refine.dropImage')}
        </span>
      </div>
    </button>
  );
}
