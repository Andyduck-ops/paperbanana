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
    method: `Database retrieval notes:
- matched three medical workflow diagrams with left-to-right sequencing
- useful reference traits: pale paper background, thin connector arrows, no 3D effects
- avoid crowded legends and avoid saturated gradients
- previous failed attempt was too dashboard-like and too colorful`,
    caption: 'Generate a clean process diagram for a tumor staging workflow. Keep the layout horizontal, highlight decision nodes, and leave enough white space for later annotation.',
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
  const [referenceImageData, setReferenceImageData] = useState<string | null>(null);
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
      return `Paper Context & References:\n${method}\n\nTarget Figure Brief:\n${cap}`;
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
        referenceImageData={referenceImageData}
        onReferenceImageChange={setReferenceImageData}
        onReferenceImageClear={() => setReferenceImageData(null)}
        disabled={isGenerating}
        examples={DEFAULT_EXAMPLES}
      />

      <section className="rounded-2xl border border-border/70 bg-card/70 px-4 py-4">
        <div className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
          <div className="space-y-1">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={batchMode}
                onChange={(e) => setBatchMode(e.target.checked)}
                disabled={isGenerating}
                className="w-4 h-4 rounded border-border text-primary focus:ring-primary"
              />
              <span className="text-sm font-medium text-foreground">{t('generate.batchMode')}</span>
            </label>
            <p className="text-xs text-muted-foreground">{t('generate.numCandidatesHint')}</p>
          </div>

          <div className={`grid gap-2 md:w-44 ${batchMode ? '' : 'opacity-60'}`}>
            <label htmlFor="num-candidates" className="text-xs font-medium text-muted-foreground">
              {t('generate.numCandidates')}
            </label>
            <input
              id="num-candidates"
              type="number"
              min={1}
              max={50}
              value={numCandidates}
              onChange={(e) => setNumCandidates(Math.max(1, Math.min(50, parseInt(e.target.value) || 1)))}
              disabled={!batchMode || isGenerating}
              className="w-full rounded-xl border border-border bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary disabled:cursor-not-allowed disabled:opacity-60"
            />
          </div>
        </div>
      </section>

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
