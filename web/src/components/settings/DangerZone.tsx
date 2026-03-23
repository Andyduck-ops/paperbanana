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
    <section className="danger-zone-card">
      <div className="danger-zone-card__content">
        <div>
          <p className="danger-zone-card__eyebrow">System safety</p>
          <h3 className="danger-zone-card__title">{t('settings.dangerZone')}</h3>
          <p className="danger-zone-card__copy">{t('settings.dangerZoneHint')}</p>
        </div>

        <button onClick={() => setShowModal(true)} className="danger-zone-card__button">
          {t('settings.resetProviders')}
        </button>
      </div>

      {result && (
        <p className={`danger-zone-card__result ${result.success ? 'is-success' : 'is-error'}`}>
          {result.message}
        </p>
      )}

      {showModal && (
        <div className="danger-zone-modal" role="dialog" aria-modal="true">
          <div className="danger-zone-modal__backdrop" onClick={() => { setShowModal(false); setConfirmText(''); }} />
          <div className="danger-zone-modal__panel">
            <h4 className="danger-zone-modal__title">{t('settings.resetConfirmTitle')}</h4>
            <p className="danger-zone-modal__copy">{t('settings.resetConfirmHint')}</p>
            <p className="danger-zone-modal__warning">{t('settings.resetWarning')}</p>

            <div className="danger-zone-modal__field">
              <label className="danger-zone-modal__label">{t('settings.typeReset')}</label>
              <input
                type="text"
                value={confirmText}
                onChange={(e) => setConfirmText(e.target.value)}
                placeholder="RESET"
                className="danger-zone-modal__input"
                data-testid="reset-confirm-input"
              />
            </div>

            <div className="danger-zone-modal__actions">
              <button
                onClick={() => {
                  setShowModal(false);
                  setConfirmText('');
                }}
                className="danger-zone-modal__cancel"
              >
                {t('common.cancel')}
              </button>
              <button
                onClick={handleReset}
                disabled={!canConfirm || isResetting}
                className="danger-zone-modal__confirm"
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
