import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ImageComparison } from './ImageComparison';

// Mock the hooks
vi.mock('../hooks', () => ({
  useLanguage: () => ({
    t: (key: string) => {
      const translations: Record<string, string> = {
        'comparison.before': 'Before',
        'comparison.after': 'After',
        'comparison.dragToCompare': 'Drag to compare',
      };
      return translations[key] || key;
    },
  }),
}));

describe('ImageComparison', () => {
  const mockBeforeImage = 'data:image/png;base64,before';
  const mockAfterImage = 'data:image/png;base64,after';

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders both images', () => {
    render(
      <ImageComparison
        beforeImage={mockBeforeImage}
        afterImage={mockAfterImage}
      />
    );

    const images = screen.getAllByRole('img');
    expect(images).toHaveLength(2);
    expect(images[0]).toHaveAttribute('src', mockBeforeImage);
    expect(images[1]).toHaveAttribute('src', mockAfterImage);
  });

  it('slider is initially at 50%', () => {
    const { container } = render(
      <ImageComparison
        beforeImage={mockBeforeImage}
        afterImage={mockAfterImage}
      />
    );

    // Check that the slider handle is at 50%
    const sliderHandle = container.querySelector('.absolute.top-0.bottom-0.w-1');
    expect(sliderHandle).toHaveStyle({ left: '50%' });
  });

  it('dragging slider changes reveal position', () => {
    const { container } = render(
      <ImageComparison
        beforeImage={mockBeforeImage}
        afterImage={mockAfterImage}
      />
    );

    const sliderContainer = container.querySelector('.relative.w-full.overflow-hidden');
    expect(sliderContainer).toBeTruthy();

    // Mock getBoundingClientRect
    const mockRect = { left: 0, width: 200, top: 0, height: 100, right: 200, bottom: 100, x: 0, y: 0, toJSON: () => {} };
    vi.spyOn(sliderContainer!, 'getBoundingClientRect').mockReturnValue(mockRect as DOMRect);

    // Simulate mouse down at 75% position (x=150)
    fireEvent.mouseDown(sliderContainer!, { clientX: 150 });

    // Slider should have moved
    const sliderHandle = container.querySelector('.absolute.top-0.bottom-0.w-1');
    expect(sliderHandle).toHaveStyle({ left: '75%' });
  });

  it('Before and After labels are displayed', () => {
    render(
      <ImageComparison
        beforeImage={mockBeforeImage}
        afterImage={mockAfterImage}
        beforeLabel="Original"
        afterLabel="Enhanced"
      />
    );

    expect(screen.getByText('Original')).toBeInTheDocument();
    expect(screen.getByText('Enhanced')).toBeInTheDocument();
  });

  it('works with touch events for mobile', () => {
    const { container } = render(
      <ImageComparison
        beforeImage={mockBeforeImage}
        afterImage={mockAfterImage}
      />
    );

    const sliderContainer = container.querySelector('.relative.w-full.overflow-hidden');
    expect(sliderContainer).toBeTruthy();

    // Mock getBoundingClientRect
    const mockRect = { left: 0, width: 200, top: 0, height: 100, right: 200, bottom: 100, x: 0, y: 0, toJSON: () => {} };
    vi.spyOn(sliderContainer!, 'getBoundingClientRect').mockReturnValue(mockRect as DOMRect);

    // Simulate touch start at 25% position (x=50)
    fireEvent.touchStart(sliderContainer!, {
      touches: [{ clientX: 50, clientY: 50 }],
    });

    // Slider should have moved
    const sliderHandle = container.querySelector('.absolute.top-0.bottom-0.w-1');
    expect(sliderHandle).toHaveStyle({ left: '25%' });
  });

  it('uses default labels when not provided', () => {
    render(
      <ImageComparison
        beforeImage={mockBeforeImage}
        afterImage={mockAfterImage}
      />
    );

    expect(screen.getByText('Before')).toBeInTheDocument();
    expect(screen.getByText('After')).toBeInTheDocument();
  });
});
