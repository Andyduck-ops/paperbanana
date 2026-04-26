import { useState } from 'react';
import { useLanguage } from '../hooks';
import { ensureDefaultConfig } from '../services/configService';

export interface WelcomeWizardProps {
  onComplete: () => void;
  onNavigateToSettings: () => void;
}

const WIZARD_COMPLETED_KEY = 'paperbanana-wizard-completed';

export function isWizardCompleted(): boolean {
  if (typeof window === 'undefined') return false;
  return localStorage.getItem(WIZARD_COMPLETED_KEY) === 'true';
}

export function markWizardCompleted(): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(WIZARD_COMPLETED_KEY, 'true');
}

export function WelcomeWizard({ onComplete, onNavigateToSettings }: WelcomeWizardProps) {
  const { t } = useLanguage();
  const [step, setStep] = useState(0);

  const steps = [
    {
      title: t('wizard.welcome') || 'Welcome to PaperBanana',
      description: t('wizard.welcomeDesc') || 'Transform your ideas into beautiful visualizations with AI-powered generation.',
      icon: (
        <svg className="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
        </svg>
      ),
    },
    {
      title: t('wizard.configure') || 'Configure Your Provider',
      description: t('wizard.configureDesc') || 'Set up your AI provider to start generating. You can always change this later in Settings.',
      icon: (
        <svg className="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
      ),
      action: {
        label: t('wizard.goToSettings') || 'Open Settings',
        onClick: async () => {
          await ensureDefaultConfig();
          markWizardCompleted();
          onNavigateToSettings();
          onComplete();
        },
      },
    },
    {
      title: t('wizard.start') || 'Start Creating',
      description: t('wizard.startDesc') || 'You are all set! Start by typing a prompt or try one of our examples.',
      icon: (
        <svg className="w-16 h-16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
      ),
    },
  ];

  const currentStep = steps[step];
  const isLastStep = step === steps.length - 1;
  const isFirstStep = step === 0;

  const handleNext = async () => {
    if (isLastStep) {
      await ensureDefaultConfig();
      markWizardCompleted();
      onComplete();
    } else {
      setStep(step + 1);
    }
  };

  const handlePrev = () => {
    if (!isFirstStep) {
      setStep(step - 1);
    }
  };

  const handleSkip = async () => {
    await ensureDefaultConfig();
    markWizardCompleted();
    onComplete();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-background/95 backdrop-blur-sm">
      <div className="w-full max-w-lg mx-4">
        <div className="bg-card border border-border rounded-2xl shadow-2xl overflow-hidden">
          <div className="flex gap-1 p-4 bg-muted/30">
            {steps.map((_, index) => (
              <div
                key={index}
                className={"h-1 flex-1 rounded-full transition-colors " + (index <= step ? 'bg-primary' : 'bg-muted')}
              />
            ))}
          </div>

          <div className="p-8 text-center">
            <div className="flex justify-center mb-6 text-primary">
              {currentStep.icon}
            </div>
            <h2 className="text-2xl font-heading font-semibold text-foreground mb-3">
              {currentStep.title}
            </h2>
            <p className="text-muted-foreground leading-relaxed mb-6">
              {currentStep.description}
            </p>

            {currentStep.action && (
              <button
                onClick={currentStep.action.onClick}
                className="w-full px-6 py-3 mb-4 text-sm font-medium text-primary-foreground bg-primary hover:bg-primary/90 rounded-lg transition-colors"
              >
                {currentStep.action.label}
              </button>
            )}
          </div>

          <div className="flex items-center justify-between px-6 py-4 border-t border-border bg-muted/20">
            <button
              onClick={handleSkip}
              className="text-sm text-muted-foreground hover:text-foreground transition-colors"
            >
              {t('wizard.skip') || 'Skip for now'}
            </button>
            <div className="flex gap-3">
              {!isFirstStep && (
                <button
                  onClick={handlePrev}
                  className="px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
                >
                  {t('common.back') || 'Back'}
                </button>
              )}
              {!currentStep.action && (
                <button
                  onClick={handleNext}
                  className="px-6 py-2 text-sm font-medium text-primary-foreground bg-primary hover:bg-primary/90 rounded-lg transition-colors"
                >
                  {isLastStep ? (t('wizard.getStarted') || 'Get Started') : (t('common.next') || 'Next')}
                </button>
              )}
            </div>
          </div>
        </div>

        <div className="text-center mt-4 text-sm text-muted-foreground">
          {step + 1} / {steps.length}
        </div>
      </div>
    </div>
  );
}

export default WelcomeWizard;
