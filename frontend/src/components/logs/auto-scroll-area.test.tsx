import { render, screen, within, fireEvent, waitFor } from '@testing-library/react';
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
const mockObservers: MockIntersectionObserver[] = [];

class MockIntersectionObserver {
  constructor(cb: IntersectionObserverCallback) {
    ioCallback = cb;
    mockObservers.push(this);
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
  mockObservers.length = 0;
});

const line = (over: Partial<LogLine> = {}): LogLine => ({
  level: Level.INFO,
  msg: 'hello',
  time: '2025-01-01T00:00:00Z',
  meta: {},
  ...over,
});
const list = (...lines: LogLine[]): LogMessages => lines;

// Radix ScrollArea exposes its scrollable viewport via this data attribute.
const viewport = (container: HTMLElement): HTMLElement => {
  const el = container.querySelector('[data-radix-scroll-area-viewport]');
  expect(el).not.toBeNull();
  return el as HTMLElement;
};

// The reset button is an icon-only button distinguished by its accessible
// name; query by role+name (intent) rather than a Tailwind class or test-id.
const resetButton = (container: HTMLElement): HTMLButtonElement | null =>
  within(container).queryByRole('button', {
    name: /scroll.*bottom/i,
  }) as HTMLButtonElement | null;

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

  it('observes the bottom anchor on mount', () => {
    render(<LogEntries logs={list(line())} />);
    // intent: a single IntersectionObserver is created and given the bottom
    // anchor to observe (the anchor exists purely to be observed)
    expect(mockObservers.length).toBe(1);
    expect(mockObservers[0].observe).toHaveBeenCalled();
  });

  it('scrolls to the bottom on mount (scrollIntoView called)', () => {
    render(<LogEntries logs={list(line())} />);
    expect(scrollIntoView).toHaveBeenCalledWith({ behavior: 'instant' });
  });

  it('scrolls to the bottom when the number of logs changes (watch prop)', () => {
    const { rerender } = render(<LogEntries logs={list(line({ msg: 'a' }))} />);
    scrollIntoView.mockClear();
    rerender(<LogEntries logs={list(line({ msg: 'a' }), line({ msg: 'b' }))} />);
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

  it('treats scroll-key keydown (ArrowDown) as a scroll intent', async () => {
    const { container } = render(<LogEntries logs={list(line())} />);
    const vp = viewport(container);

    // Scrolled away from the bottom (IO not intersecting) -> no scroll intent yet
    triggerIo(false);
    expect(resetButton(container)).toBeNull();

    // ArrowDown is a scroll key -> onKeyDown -> onUserScrollIntent -> shows reset
    fireEvent.keyDown(vp, { key: 'ArrowDown' });
    await waitFor(() => expect(resetButton(container)).not.toBeNull());

    // 'Enter' is not a scroll key -> onKeyDown must not fire scroll intent
    fireEvent.keyDown(vp, { key: 'Enter' });
    expect(resetButton(container)).not.toBeNull();
  });

  it('does not show the reset button for a scroll intent already at the bottom', () => {
    const { container } = render(<LogEntries logs={list(line())} />);
    // bottom is visible on mount (scrollToBottom) and we stay at the bottom;
    // a scroll-intent event here hits neither branch of onUserScrollIntent,
    // so the reset button must stay hidden.
    triggerIo(true);
    fireEvent.wheel(viewport(container));
    expect(resetButton(container)).toBeNull();
  });

  it('does not auto-scroll on watch change when the user has already scrolled away', async () => {
    const { container, rerender } = render(<LogEntries logs={list(line({ msg: 'a' }))} />);
    // user scrolls away -> reset button appears (manual scroll intent)
    triggerIo(false);
    fireEvent.wheel(viewport(container));
    await waitFor(() => expect(resetButton(container)).not.toBeNull());
    // rerendering with more logs must NOT auto-scroll to the bottom,
    // because the user is browsing manually (userHasScrolledRef is true).
    scrollIntoView.mockClear();
    rerender(<LogEntries logs={list(line({ msg: 'a' }), line({ msg: 'b' }))} />);
    expect(scrollIntoView).not.toHaveBeenCalled();
  });
});
