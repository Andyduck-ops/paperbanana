import { useState } from 'react';
import { useLanguage } from '../hooks';

export interface CriticSuggestionsProps {
  suggestions: string;
  roundNumber: number;
  hasChanges: boolean;
}

export function CriticSuggestions({
  suggestions,
  roundNumber,
  hasChanges
}: CriticSuggestionsProps) {
  const { t } = useLanguage();
  const [isExpanded, setIsExpanded] = useState(false);

  if (!suggestions) return null;

  return (
    <div className="critic-suggestions mt-2">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-center justify-between p-2 text-sm bg-blue-500/10 rounded border border-blue-500/30 hover:bg-blue-500/20"
      >
        <span className="flex items-center gap-2">
          <span>💡</span>
          <span>{t('generate.criticSuggestions', { round: roundNumber })}</span>
        </span>
        <span className="text-xs text-muted-foreground">
          {isExpanded ? '▲' : '▼'}
        </span>
      </button>

      {isExpanded && (
        <div className="mt-1 p-3 bg-background border border-border rounded text-sm">
          {hasChanges ? (
            <p className="text-foreground whitespace-pre-wrap">{suggestions}</p>
          ) : (
            <p className="text-green-600 flex items-center gap-2">
              <span>✅</span>
              {t('generate.noChangesNeeded')}
            </p>
          )}
        </div>
      )}
    </div>
  );
}
