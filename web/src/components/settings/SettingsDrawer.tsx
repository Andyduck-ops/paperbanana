import { useLanguage } from '../../hooks';
import { SettingsPage } from '../../pages/SettingsPage';

export interface SettingsDrawerProps {
  isOpen: boolean;
  onClose: () => void;
}

export function SettingsDrawer({ isOpen, onClose }: SettingsDrawerProps) {
  const { t } = useLanguage();

  if (!isOpen) {
    return null;
  }

  return (
    <>
      <div
        className={`settings-drawer__backdrop ${isOpen ? 'settings-drawer__backdrop--open' : ''}`}
        onClick={onClose}
        aria-hidden={!isOpen}
      />

      <div
        className={`settings-drawer ${isOpen ? 'settings-drawer--open' : ''}`}
        role="dialog"
        aria-modal="true"
        aria-hidden={!isOpen}
        aria-label={t('settings.title') || 'Settings'}
      >
        <SettingsPage onBack={onClose} variant="drawer" />
      </div>
    </>
  );
}
