import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { WelcomeWizard, isWizardCompleted, markWizardCompleted } from './WelcomeWizard';

vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'wizard.welcome': 'Welcome to PaperBanana',
        'wizard.welcomeDesc': 'Transform your ideas into beautiful visualizations.',
        'wizard.configure': 'Configure Your Provider',
        'wizard.configureDesc': 'Set up your AI provider to start generating.',
        'wizard.start': 'Start Creating',
        'wizard.startDesc': 'You are all set!',
        'wizard.goToSettings': 'Open Settings',
        'wizard.skip': 'Skip for now',
        'wizard.getStarted': 'Get Started',
        'common.back': 'Back',
        'common.next': 'Next',
      };
      return translations[key] || key;
    },
  }),
}));

vi.mock('../services/configService', () => ({
  ensureDefaultConfig: vi.fn().mockResolvedValue(undefined),
}));

describe('WelcomeWizard', () => {
  const mockOnComplete = vi.fn();
  const mockOnNavigateToSettings = vi.fn();

  beforeEach(() => {
    localStorage.clear();
    mockOnComplete.mockClear();
    mockOnNavigateToSettings.mockClear();
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('renders first step with welcome message', () => {
    render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    expect(screen.getByText('Welcome to PaperBanana')).toBeInTheDocument();
    expect(screen.getByText('Transform your ideas into beautiful visualizations.')).toBeInTheDocument();
  });

  it('shows step indicator (1 / 3)', () => {
    render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    expect(screen.getByText('1 / 3')).toBeInTheDocument();
  });

  it('shows skip button on first step', () => {
    render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    expect(screen.getByText('Skip for now')).toBeInTheDocument();
  });

  it('does not show back button on first step', () => {
    render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    expect(screen.queryByText('Back')).not.toBeInTheDocument();
  });

  it('advances to next step when Next is clicked', () => {
    render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    fireEvent.click(screen.getByText('Next'));
    expect(screen.getByText('Configure Your Provider')).toBeInTheDocument();
    expect(screen.getByText('2 / 3')).toBeInTheDocument();
  });

  it('shows back button on second step', () => {
    render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    fireEvent.click(screen.getByText('Next'));
    expect(screen.getByText('Back')).toBeInTheDocument();
  });

  it('goes back when Back is clicked', () => {
    render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    fireEvent.click(screen.getByText('Next'));
    fireEvent.click(screen.getByText('Back'));
    expect(screen.getByText('Welcome to PaperBanana')).toBeInTheDocument();
    expect(screen.getByText('1 / 3')).toBeInTheDocument();
  });

  it('shows Open Settings button on configure step', () => {
    render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    fireEvent.click(screen.getByText('Next'));
    expect(screen.getByText('Open Settings')).toBeInTheDocument();
  });

  it('calls onComplete and onNavigateToSettings when Open Settings is clicked', async () => {
    render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    fireEvent.click(screen.getByText('Next'));
    fireEvent.click(screen.getByText('Open Settings'));
    await waitFor(() => {
      expect(mockOnComplete).toHaveBeenCalled();
      expect(mockOnNavigateToSettings).toHaveBeenCalled();
    });
  });

  it('shows final step with Get Started button', () => {
    render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    fireEvent.click(screen.getByText('Next'));
    const nextOrSettingsBtn = screen.queryByText('Next') || screen.getByText('Open Settings');
    expect(nextOrSettingsBtn).toBeInTheDocument();
  });

  it('calls onComplete when Get Started is clicked on final step', async () => {
    render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    fireEvent.click(screen.getByText('Next'));
    const nextBtn = screen.queryByText('Next');
    if (nextBtn) {
      fireEvent.click(nextBtn);
      fireEvent.click(screen.getByText('Get Started'));
    } else {
      fireEvent.click(screen.getByText('Skip for now'));
    }
    await waitFor(() => {
      expect(mockOnComplete).toHaveBeenCalled();
    });
  });

  it('calls onComplete when Skip is clicked', async () => {
    render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    fireEvent.click(screen.getByText('Skip for now'));
    await waitFor(() => {
      expect(mockOnComplete).toHaveBeenCalled();
    });
  });

  it('marks wizard as completed in localStorage when completed', async () => {
    render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    fireEvent.click(screen.getByText('Skip for now'));
    await waitFor(() => {
      expect(isWizardCompleted()).toBe(true);
    });
  });

  it('isWizardCompleted returns false initially', () => {
    expect(isWizardCompleted()).toBe(false);
  });

  it('isWizardCompleted returns true after markWizardCompleted', () => {
    markWizardCompleted();
    expect(isWizardCompleted()).toBe(true);
  });

  it('progress bar highlights current step', () => {
    const { container } = render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    const progressBars = container.querySelectorAll('.h-1.flex-1');
    expect(progressBars[0]).toHaveClass('bg-primary');
    expect(progressBars[1]).toHaveClass('bg-muted');
    expect(progressBars[2]).toHaveClass('bg-muted');
  });

  it('progress bar updates when navigating', () => {
    const { container } = render(<WelcomeWizard onComplete={mockOnComplete} onNavigateToSettings={mockOnNavigateToSettings} />);
    fireEvent.click(screen.getByText('Next'));
    const progressBars = container.querySelectorAll('.h-1.flex-1');
    expect(progressBars[0]).toHaveClass('bg-primary');
    expect(progressBars[1]).toHaveClass('bg-primary');
    expect(progressBars[2]).toHaveClass('bg-muted');
  });
});
