import { useState } from 'react';
import { useTranslation } from 'react-i18next';

export interface DangerZoneProps {
  onReset?: () => Promise<void>;
}

export function DangerZone({ onReset }: DangerZoneProps) {
  const { t } = useTranslation();
  const [showModal, setShowModal] = useState(false);
  const [confirmText, setConfirmText] = useState('');
  const [isResetting, setIsResetting] = useState(false);
  const [result, setResult] = useState<{ success?: boolean; message?: string } | null>(null);

  const handleReset = async () => {
    if (confirmText !== 'RESET') return;

    setIsResetting(true);
    setResult(null);

    try {
      if (onReset) {
        await onReset();
        setResult({ success: true, message: t('settings.resetSuccess', { count: 0 }) });
      } else {
        const response = await fetch('/api/v1/providers/reset', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ confirm: 'RESET' }),
        });

        const data = await response.json();

        if (response.ok) {
          setResult({ success: true, message: t('settings.resetSuccess', { count: data.keys_cleared }) });
          setConfirmText('');
        } else {
          setResult({ success: false, message: data.error || t('common.error') });
        }
      }
    } catch {
      setResult({ success: false, message: t('error.networkError') });
    } finally {
      setIsResetting(false);
      setShowModal(false);
    }
  };

  const canConfirm = confirmText === 'RESET';

  return (
    <section className="bg-card rounded-xl border border-status-error/30 p-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-status-error/80">System safety</p>
          <h3 className="text-lg font-semibold text-foreground mt-1">{t('settings.dangerZone')}</h3>
          <p className="text-sm text-muted-foreground mt-1">{t('settings.dangerZoneHint')}</p>
        </div>

        <button onClick={() => setShowModal(true)} className="inline-flex items-center px-4 py-2 text-sm font-medium text-status-error bg-status-error/10 hover:bg-status-error/20 border border-status-error/30 rounded-xl transition-colors">
          {t('settings.resetProviders')}
        </button>
      </div>

      {result && (
        <p className={`mt-4 text-sm ${result.success ? 'text-status-success' : 'text-status-error'}`}>
          {result.message}
        </p>
      )}

      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center" role="dialog" aria-modal="true">
          <div className="fixed inset-0 bg-black/50 backdrop-blur-sm" onClick={() => { setShowModal(false); setConfirmText(''); }} />
          <div className="relative bg-card rounded-2xl border border-border shadow-2xl p-6 w-full max-w-md mx-4">
            <h4 className="text-lg font-semibold text-foreground">{t('settings.resetConfirmTitle')}</h4>
            <p className="text-sm text-muted-foreground mt-2">{t('settings.resetConfirmHint')}</p>
            <p className="text-sm text-status-error mt-2">{t('settings.resetWarning')}</p>

            <div className="mt-4 space-y-1.5">
              <label className="block text-sm font-medium text-foreground">{t('settings.typeReset')}</label>
              <input
                type="text"
                value={confirmText}
                onChange={(e) => setConfirmText(e.target.value)}
                placeholder="RESET"
                className="w-full px-3 py-2 text-sm bg-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
                data-testid="reset-confirm-input"
              />
            </div>

            <div className="flex items-center justify-end gap-3 mt-6">
              <button
                onClick={() => {
                  setShowModal(false);
                  setConfirmText('');
                }}
                className="px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
              >
                {t('common.cancel')}
              </button>
              <button
                onClick={handleReset}
                disabled={!canConfirm || isResetting}
                className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-status-error rounded-xl hover:opacity-90 transition-opacity disabled:opacity-50"
                data-testid="reset-confirm-button"
              >
                {isResetting ? t('common.loading') : t('settings.confirmReset')}
              </button>
            </div>
          </div>
        </div>
      )}
    </section>
  );
}
