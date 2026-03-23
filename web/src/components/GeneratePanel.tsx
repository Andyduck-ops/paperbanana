import { useState } from 'react';
import { useLanguage } from '../hooks';
import { useProviders } from '../hooks/useProviders';
import { DualInputPanel } from './DualInputPanel';
import { ConfigPanel, GenerationConfig } from './ConfigPanel';

export interface GenerateOptions {
  visualizerNode?: string;
  numCandidates?: number;
  config?: GenerationConfig;
}

export interface GeneratePanelProps {
  onGenerate: (prompt: string, options?: GenerateOptions) => void;
  isGenerating?: boolean;
  visualizerNodes?: string[];
  onNavigateToSettings?: () => void;
}

// Example content for users to load
const DEFAULT_EXAMPLES = [
  {
    method: `We propose a novel attention mechanism that combines sparse attention patterns with local context windows. The architecture consists of three main components: (1) a sparse global attention layer that captures long-range dependencies, (2) a local attention layer with sliding windows for fine-grained patterns, and (3) a gating mechanism that dynamically adjusts the contribution of each component based on input characteristics.`,
    caption: 'Figure 1: Architecture of the proposed sparse-local attention mechanism. The global attention layer (top) captures long-range dependencies, while the local attention layer (bottom) processes sliding windows. The gating network dynamically combines outputs from both layers.',
  },
];

export function GeneratePanel({
  onGenerate,
  isGenerating = false,
  visualizerNodes = [],
  onNavigateToSettings,
}: GeneratePanelProps) {
  const { t } = useLanguage();
  const { providers } = useProviders();
  const [methodContent, setMethodContent] = useState('');
  const [caption, setCaption] = useState('');
  const [selectedNode, setSelectedNode] = useState<string>('');
  const [batchMode, setBatchMode] = useState(false);
  const [numCandidates, setNumCandidates] = useState(3);
  const [config, setConfig] = useState<GenerationConfig>({
    aspectRatio: '16:9',
    criticRounds: 3,
    retrievalMode: 'auto',
    pipelineMode: 'full',
    queryModel: undefined,
    genModel: undefined,
  });

  const buildCombinedPrompt = (): string => {
    const method = methodContent.trim();
    const cap = caption.trim();

    if (method && cap) {
      return `Method Section:\n${method}\n\nFigure Caption:\n${cap}`;
    }
    return method || cap;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const combinedPrompt = buildCombinedPrompt();
    if (combinedPrompt) {
      onGenerate(combinedPrompt, {
        visualizerNode: selectedNode || undefined,
        numCandidates: batchMode ? numCandidates : undefined,
        config,
      });
    }
  };

  const hasContent = methodContent.trim() || caption.trim();

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <DualInputPanel
        methodContent={methodContent}
        caption={caption}
        onMethodChange={setMethodContent}
        onCaptionChange={setCaption}
        disabled={isGenerating}
        examples={DEFAULT_EXAMPLES}
      />

      {visualizerNodes.length > 0 && (
        <div>
          <label htmlFor="visualizer-node" className="block text-sm font-medium text-foreground mb-2">
            {t('generate.visualizerNode')}
          </label>
          <select
            id="visualizer-node"
            value={selectedNode}
            onChange={(e) => setSelectedNode(e.target.value)}
            disabled={isGenerating}
            className="w-full px-4 py-2 rounded-lg border border-border bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
          >
            <option value="">{t('generate.defaultVisualizer')}</option>
            {visualizerNodes.map((node) => (
              <option key={node} value={node}>{node}</option>
            ))}
          </select>
        </div>
      )}

      <div className="flex items-center gap-3 py-2">
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={batchMode}
            onChange={(e) => setBatchMode(e.target.checked)}
            disabled={isGenerating}
            className="w-4 h-4 rounded border-border text-primary focus:ring-primary"
          />
          <span className="text-sm text-foreground">{t('generate.batchMode')}</span>
        </label>
        {batchMode && (
          <div className="flex items-center gap-2">
            <input
              type="number"
              min={1}
              max={50}
              value={numCandidates}
              onChange={(e) => setNumCandidates(Math.max(1, Math.min(50, parseInt(e.target.value) || 1)))}
              disabled={isGenerating}
              className="w-16 px-2 py-1 rounded border border-border bg-background text-foreground text-sm focus:outline-none focus:ring-2 focus:ring-primary"
            />
            <span className="text-sm text-muted-foreground">{t('generate.numCandidatesHint')}</span>
          </div>
        )}
      </div>

      <div className="mt-4">
        <ConfigPanel
          config={config}
          onChange={setConfig}
          providers={providers}
          disabled={isGenerating}
          onNavigateToSettings={onNavigateToSettings}
        />
      </div>

      <button
        type="submit"
        disabled={!hasContent || isGenerating}
        className="w-full px-6 py-3 rounded-lg bg-primary text-background font-medium hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed transition-opacity"
      >
        {isGenerating ? t('generate.generating') : t('generate.submit')}
      </button>
    </form>
  );
}
