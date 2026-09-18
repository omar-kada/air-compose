import { render, screen, fireEvent } from '@testing-library/react';
import { AutoScrollArea } from './auto-scroll-area';

// AutoScrollArea wraps Radix <ScrollArea>; render it as a plain div that
// forwards the ref so the wheel/keydown listeners attach to a node we can
// target, and so the bottom anchor lands in the DOM.
vi.mock('@/components/ui/scroll-area', async () => {
  const React = await import('react');
  return {
    ScrollArea: React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
      ({ children }, ref) => (
        <div ref={ref} data-testid="scroll-area">
          {children}
        </div>
      ),
    ),
  };
});

// jsdom ships neither IntersectionObserver nor Element#scrollIntoView.
let ioCallback: IntersectionObserverCallback | null = null;
const scrollIntoView = vi.fn();

class MockIntersectionObserver {
  constructor(cb: IntersectionObserverCallback) {
    ioCallback = cb;
  }
  observe() {}
  unobserve() {}
  disconnect() {}
  takeRecords() {
    return [];
  }
}

beforeAll(() => {
  (
    globalThis as unknown as { IntersectionObserver: typeof MockIntersectionObserver }
  ).IntersectionObserver = MockIntersectionObserver;
  Element.prototype.scrollIntoView = scrollIntoView;
});

afterEach(() => {
  vi.clearAllMocks();
  ioCallback = null;
});

describe('AutoScrollArea', () => {
  it('renders children and the bottom anchor', () => {
    render(
      <AutoScrollArea>
        <span data-testid="child">log line</span>
      </AutoScrollArea>,
    );
    expect(screen.getByTestId('child')).toBeTruthy();
    expect(screen.getByTestId('scroll-area').querySelector('.h-1')).not.toBeNull();
  });

  it('scrolls to the bottom on mount', () => {
    render(<AutoScrollArea>child</AutoScrollArea>);
    expect(scrollIntoView).toHaveBeenCalled();
    expect(scrollIntoView).toHaveBeenCalledWith({ behavior: 'instant' });
  });

  it('shows a reset button when scrolling away from the bottom', () => {
    const { queryByRole } = render(<AutoScrollArea>child</AutoScrollArea>);
    const viewport = screen.getByTestId('scroll-area');
    expect(queryByRole('button')).toBeNull();
    ioCallback?.([{ isIntersecting: false } as IntersectionObserverEntry], undefined as never);
    fireEvent.wheel(viewport, { deltaY: 1 });
    expect(queryByRole('button')).not.toBeNull();
  });

  it('hides the reset button when scrolling back to the bottom', () => {
    const { queryByRole } = render(<AutoScrollArea>child</AutoScrollArea>);
    const viewport = screen.getByTestId('scroll-area');
    ioCallback?.([{ isIntersecting: false } as IntersectionObserverEntry], undefined as never);
    fireEvent.wheel(viewport);
    expect(queryByRole('button')).not.toBeNull();
    ioCallback?.([{ isIntersecting: true } as IntersectionObserverEntry], undefined as never);
    fireEvent.wheel(viewport);
    expect(queryByRole('button')).toBeNull();
  });

  it('scrolls smoothly to the bottom when the reset button is clicked', () => {
    const { getByRole } = render(<AutoScrollArea>child</AutoScrollArea>);
    const viewport = screen.getByTestId('scroll-area');
    ioCallback?.([{ isIntersecting: false } as IntersectionObserverEntry], undefined as never);
    fireEvent.wheel(viewport);
    scrollIntoView.mockClear();
    fireEvent.click(getByRole('button'));
    expect(scrollIntoView).toHaveBeenCalledWith({ behavior: 'smooth' });
  });

  it('scrolls to the bottom when the `watch` prop changes', () => {
    const { rerender } = render(<AutoScrollArea watch={1}>child</AutoScrollArea>);
    scrollIntoView.mockClear();
    rerender(<AutoScrollArea watch={2}>child</AutoScrollArea>);
    expect(scrollIntoView).toHaveBeenCalledWith({ behavior: 'instant' });
  });

  it('treats keyboard navigation (ArrowDown) as a scroll intent', () => {
    const { queryByRole } = render(<AutoScrollArea>child</AutoScrollArea>);
    const viewport = screen.getByTestId('scroll-area');
    ioCallback?.([{ isIntersecting: false } as IntersectionObserverEntry], undefined as never);
    fireEvent.keyDown(viewport, { key: 'ArrowDown' });
    expect(queryByRole('button')).not.toBeNull();
  });
});
