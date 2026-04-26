export interface BatchVariant {
  id: string;
  index: number;
  imageUrl?: string;
  isSelected?: boolean;
}

export interface BatchVariantGridProps {
  variants?: BatchVariant[];
  onVariantSelect?: (variantId: string) => void;
  selectedVariantId?: string;
  className?: string;
}

export function BatchVariantGrid({
  variants = [],
  onVariantSelect,
  selectedVariantId,
  className = ''
}: BatchVariantGridProps) {
  return (
    <div className={`batch-variant-grid ${className}`}>
      <div className="batch-variant-grid__header">
        <h4>Variants</h4>
      </div>
      <div className="batch-variant-grid__grid">
        {variants.map((variant) => (
          <button
            key={variant.id}
            className={`batch-variant-grid__item ${selectedVariantId === variant.id ? 'selected' : ''}`}
            onClick={() => onVariantSelect?.(variant.id)}
          >
            {variant.imageUrl && (
              <img src={variant.imageUrl} alt={`Variant ${variant.index}`} />
            )}
            <span className="batch-variant-grid__index">{variant.index}</span>
          </button>
        ))}
      </div>
    </div>
  );
}
