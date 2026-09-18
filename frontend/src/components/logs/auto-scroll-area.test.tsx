import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import type { LogLine, LogMessages } from '@/api';
import { Level } from '@/api';
import { LogEntries } from './logs-entries';

// Integration test: exercise AutoScrollArea *through* LogEntries using the
// project's real ScrollArea (no mock of AutoScrollArea / ScrollArea). jsdom
// lacks IntersectionObserver / Element#scrollIntoView, so stub only those and
// feed intersection entries to drive the reset-button show/hide behaviour.
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => `translated:${key}` }),
}));

let ioCallback: IntersectionObserverCallback | null = null;
const scrollIntoView = vi.fn();

class MockIntersectionObserver {
  constructor(cb: IntersectionObserverCallback) {
    ioCallback = cb;
  }
  observe = vi.fn();
  unobserve = vi.fn();
  disconnect = vi.fn();
  takeRecords = vi.fn((): IntersectionObserverEntry[] => []);
}

class MockResizeObserver {
  observe = vi.fn();
  unobserve = vi.fn();
  disconnect = vi.fn();
}

beforeAll(() => {
  vi.stubGlobal('IntersectionObserver', MockIntersectionObserver);
  vi.stubGlobal('ResizeObserver', MockResizeObserver);
  Element.prototype.scrollIntoView = scrollIntoView;
});

afterAll(() => {
  vi.unstubAllGlobals();
});

afterEach(() => {
  vi.clearAllMocks();
  ioCallback = null;
});

const line = (over: Partial<LogLine> = {}): LogLine => ({
  level: Level.INFO,
  msg: 'hello',
  time: '2025-01-01T00:00:00Z',
  meta: {},
  ...over,
});
const list = (...lines: LogLine[]): LogMessages => lines;

const viewport = (container: HTMLElement): HTMLElement => {
  const el = container.querySelector('[data-radix-scroll-area-viewport]');
  expect(el).not.toBeNull();
  return el as HTMLElement;
};

const resetButton = (container: HTMLElement): HTMLButtonElement | null =>
  container.querySelector<HTMLButtonElement>('button.fixed.bottom-16.right-6');

const triggerIo = (intersecting: boolean) =>
  ioCallback?.([{ isIntersecting: intersecting } as IntersectionObserverEntry], undefined as never);

describe('LogEntries + AutoScrollArea integration', () => {
  it('renders log lines through the real ScrollArea viewport', () => {
    const { container } = render(
      <LogEntries logs={list(line({ msg: 'first' }), line({ msg: 'second' }))} />,
    );
    expect(screen.queryByText('first')).not.toBeNull();
    expect(screen.queryByText('second')).not.toBeNull();
    expect(viewport(container)).not.toBeNull();
  });

  it('renders the bottom anchor used to observe intersection', () => {
    const { container } = render(<LogEntries logs={list(line())} />);
    expect(container.querySelector('.h-1')).not.toBeNull();
  });

  it('scrolls to the bottom on mount (scrollIntoView called)', () => {
    render(<LogEntries logs={list(line())} />);
    expect(scrollIntoView).toHaveBeenCalledWith({ behavior: 'instant' });
  });

  it('shows the reset button when scrolled away from the bottom', async () => {
    const { container } = render(<LogEntries logs={list(line())} />);
    triggerIo(false);
    fireEvent.wheel(viewport(container), { deltaY: 1 });
    await waitFor(() => expect(resetButton(container)).not.toBeNull());
  });

  it('hides the reset button when scrolled back to the bottom', async () => {
    const { container } = render(<LogEntries logs={list(line())} />);
    triggerIo(false);
    fireEvent.wheel(viewport(container));
    await waitFor(() => expect(resetButton(container)).not.toBeNull());
    triggerIo(true);
    fireEvent.wheel(viewport(container));
    await waitFor(() => expect(resetButton(container)).toBeNull());
  });

  it('scrolls smoothly when the reset button is clicked', async () => {
    const { container } = render(<LogEntries logs={list(line())} />);
    triggerIo(false);
    fireEvent.wheel(viewport(container));
    await waitFor(() => expect(resetButton(container)).not.toBeNull());
    scrollIntoView.mockClear();
    const btn = resetButton(container);
    expect(btn).not.toBeNull();
    if (btn) fireEvent.click(btn);
    expect(scrollIntoView).toHaveBeenCalledWith({ behavior: 'smooth' });
  });

  it('re-scrolls when the number of logs changes (watch prop)', () => {
    const { rerender } = render(<LogEntries logs={list(line({ msg: 'a' }))} />);
    scrollIntoView.mockClear();
    rerender(<LogEntries logs={list(line({ msg: 'a' }), line({ msg: 'b' }))} />);
    expect(scrollIntoView).toHaveBeenCalledWith({ behavior: 'instant' });
  });
});
