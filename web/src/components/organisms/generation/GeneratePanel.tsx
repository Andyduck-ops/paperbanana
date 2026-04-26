import { useState, useEffect, memo } from 'react';
import { useLanguage, usePromptTemplates } from '../../../hooks';
import { useProviders } from '../../../hooks/useProviders';
import { DualInputPanel } from '../../DualInputPanel';
import { ConfigPanel, GenerationConfig } from '../../ConfigPanel';
import { TemplateSelector } from '../../TemplateSelector';

export interface GenerateOptions {
  visualizerNode?: string;
  numCandidates?: number;
  content?: string;
  visualIntent?: string;
  config?: GenerationConfig;
}

export interface GeneratePanelProps {
  onGenerate: (prompt: string, options?: GenerateOptions) => void;
  isGenerating?: boolean;
  visualizerNodes?: string[];
  onNavigateToSettings?: () => void;
  initialMethodContent?: string;
  initialCaption?: string;
  collapsed?: boolean;
}

function GeneratePanelComponent({
  onGenerate,
  isGenerating = false,
  visualizerNodes = [],
  onNavigateToSettings,
  initialMethodContent,
  initialCaption,
  collapsed = false,
}: GeneratePanelProps) {
  const { t } = useLanguage();
  const { providers } = useProviders();
  const { templates, addTemplate, deleteTemplate } = usePromptTemplates();
  const [methodContent, setMethodContent] = useState(initialMethodContent || '');
  const [caption, setCaption] = useState(initialCaption || '');
  const [referenceImageData, setReferenceImageData] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<string>('');
  const [batchMode, setBatchMode] = useState(false);
  const [numCandidates, setNumCandidates] = useState(3);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [config, setConfig] = useState<GenerationConfig>({
    aspectRatio: '16:9',
    criticRounds: 3,
    retrievalMode: 'auto',
    pipelineMode: 'full',
    queryModel: undefined,
    genModel: undefined,
  });

  useEffect(() => {
    if (initialMethodContent !== undefined) {
      setMethodContent(initialMethodContent);
    }
    if (initialCaption !== undefined) {
      setCaption(initialCaption);
    }
  }, [initialMethodContent, initialCaption]);

  const buildCombinedPrompt = (): string => {
    const method = methodContent.trim();
    const cap = caption.trim();

    if (method && cap) {
      return `Paper Context & References:\n${method}\n\nTarget Figure Brief:\n${cap}`;
    }
    return method || cap;
  };

  const handleSelectTemplate = (template: { methodContent: string; caption: string }) => {
    setMethodContent(template.methodContent);
    setCaption(template.caption);
  };

  const handleSaveTemplate = (name: string) => {
    addTemplate(name, methodContent, caption);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const method = methodContent.trim();
    const cap = caption.trim();
    const combinedPrompt = buildCombinedPrompt();
    if (combinedPrompt) {
      onGenerate(combinedPrompt, {
        content: method || undefined,
        visualIntent: cap || undefined,
        visualizerNode: selectedNode || undefined,
        numCandidates: batchMode ? numCandidates : undefined,
        config,
      });
    }
  };

  const hasContent = methodContent.trim() || caption.trim();

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      {/* Value Proposition */}
      <div className="text-center pb-2">
        <h1 className="text-xl font-heading font-semibold text-foreground">
          {t('app.tagline')}
        </h1>
        <p className="text-sm text-muted-foreground mt-1">
          {t('generate.subtitle') || '输入论文上下文与目标描述，AI 自动生成学术插图'}
        </p>
      </div>

      {/* Core Input */}
      <DualInputPanel
        methodContent={methodContent}
        caption={caption}
        onMethodChange={setMethodContent}
        onCaptionChange={setCaption}
        referenceImageData={referenceImageData}
        onReferenceImageChange={setReferenceImageData}
        onReferenceImageClear={() => setReferenceImageData(null)}
        disabled={isGenerating}
        collapsed={collapsed || isGenerating}
      />

      {/* Primary CTA - visually dominant */}
      <button
        type="submit"
        disabled={!hasContent || isGenerating}
        className="w-full px-6 py-4 rounded-xl bg-primary text-primary-foreground font-semibold text-base hover:opacity-90 active:scale-[0.98] disabled:opacity-40 disabled:cursor-not-allowed disabled:active:scale-100 transition-all duration-200 flex items-center justify-center gap-2"
      >
        {isGenerating ? (
          <>
            <div className="w-5 h-5 border-2 border-primary-foreground border-t-transparent rounded-full animate-spin" />
            {t('generate.generating')}
          </>
        ) : (
          <>
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            {t('generate.submit')}
          </>
        )}
      </button>

      {/* Secondary Actions Row */}
      <div className="flex items-center justify-between gap-3">
        <TemplateSelector
          templates={templates}
          onSelect={handleSelectTemplate}
          onSave={handleSaveTemplate}
          onDelete={deleteTemplate}
          disabled={isGenerating}
          currentMethodContent={methodContent}
          currentCaption={caption}
        />

        <button
          type="button"
          onClick={() => setShowAdvanced(!showAdvanced)}
          className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
        >
          <svg
            className={`w-4 h-4 transition-transform ${showAdvanced ? 'rotate-180' : ''}`}
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
          {t('generate.advancedSettings') || '高级选项'}
        </button>
      </div>

      {/* Advanced Options - Collapsed by default */}
      {showAdvanced && (
        <div className="space-y-4 pt-2 border-t border-border/50">
          {/* Batch Mode */}
          <section className="rounded-xl border border-border/50 bg-card/50 px-4 py-3">
            <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
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

              <div className={`grid gap-1 md:w-40 ${batchMode ? '' : 'opacity-50'}`}>
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
                  className="w-full rounded-lg border border-border bg-background px-3 py-1.5 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary disabled:cursor-not-allowed disabled:opacity-50"
                />
              </div>
            </div>
          </section>

          {/* Visualizer Node */}
          {visualizerNodes.length > 0 && (
            <div>
              <label htmlFor="visualizer-node" className="block text-sm font-medium text-foreground mb-1.5">
                {t('generate.visualizerNode')}
              </label>
              <select
                id="visualizer-node"
                value={selectedNode}
                onChange={(e) => setSelectedNode(e.target.value)}
                disabled={isGenerating}
                className="w-full px-3 py-2 rounded-lg border border-border bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary text-sm"
              >
                <option value="">{t('generate.defaultVisualizer')}</option>
                {visualizerNodes.map((node) => (
                  <option key={node} value={node}>{node}</option>
                ))}
              </select>
            </div>
          )}

          {/* Config Panel */}
          <ConfigPanel
            config={config}
            onChange={setConfig}
            providers={providers}
            disabled={isGenerating}
            onNavigateToSettings={onNavigateToSettings}
          />
        </div>
      )}
    </form>
  );
}

export const GeneratePanel = memo(GeneratePanelComponent);
GeneratePanel.displayName = 'GeneratePanel';
