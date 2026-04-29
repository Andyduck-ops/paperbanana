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
    <div className={`bg-card rounded-xl border border-border/30 p-6 ${className}`}>
      <div className="mb-4">
        <h4 className="text-lg font-semibold text-foreground">Variants</h4>
      </div>
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3">
        {variants.map((variant) => (
          <button
            key={variant.id}
            className={`relative aspect-square rounded-xl border overflow-hidden transition-all hover:scale-[1.02] ${selectedVariantId === variant.id ? 'border-primary ring-2 ring-primary/30' : 'border-border/50 hover:border-primary/30'}`}
            onClick={() => onVariantSelect?.(variant.id)}
          >
            {variant.imageUrl && (
              <img src={variant.imageUrl} alt={`Variant ${variant.index}`} className="w-full h-full object-cover" />
            )}
            <span className="absolute bottom-2 right-2 w-6 h-6 flex items-center justify-center rounded-full bg-black/50 text-white text-xs font-medium">
              {variant.index}
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}
