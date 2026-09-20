import { render, screen, fireEvent, act } from '@testing-library/react';
import { ConfigViewer } from './config-viewer';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));
vi.mock('@/hooks/theme-provider', () => ({
  useTheme: () => ({ theme: 'light' as const, setTheme: vi.fn() }),
}));

describe('ConfigViewer', () => {
  const mockWriteText = vi.fn();
  let originalClipboard: typeof navigator.clipboard;

  beforeEach(() => {
    vi.useFakeTimers();
    mockWriteText.mockResolvedValue(undefined);
    originalClipboard = navigator.clipboard;
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: mockWriteText },
      configurable: true,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    Object.defineProperty(navigator, 'clipboard', {
      value: originalClipboard,
      configurable: true,
    });
  });

  describe('rendering', () => {
    it('renders the yaml text', () => {
      const { container } = render(<ConfigViewer text="globalVariables: {}" />);
      expect(container.textContent).toContain('globalVariables: {}');
    });

    it('renders the config.yaml title', () => {
      render(<ConfigViewer text="test" />);
      expect(screen.getByText('config.yaml')).not.toBeNull();
    });

    it('renders the copy button when onClose is not provided', () => {
      render(<ConfigViewer text="test" />);
      expect(screen.getByRole('button', { name: 'ACTION.COPY' })).not.toBeNull();
    });

    it('does not render the close button when onClose is not provided', () => {
      render(<ConfigViewer text="test" />);
      expect(screen.queryByRole('button', { name: 'ACTION.CLOSE' })).toBeNull();
    });

    it('renders both close and copy buttons when onClose is provided', () => {
      render(<ConfigViewer text="test" onClose={vi.fn()} />);
      expect(screen.getByRole('button', { name: 'ACTION.CLOSE' })).not.toBeNull();
      expect(screen.getByRole('button', { name: 'ACTION.COPY' })).not.toBeNull();
    });
  });

  describe('copy functionality', () => {
    it('calls navigator.clipboard.writeText with the text', () => {
      render(<ConfigViewer text="yaml-content" />);
      fireEvent.click(screen.getByRole('button', { name: 'ACTION.COPY' }));
      expect(mockWriteText).toHaveBeenCalledWith('yaml-content');
    });

    it('shows COPIED text immediately after clicking copy', () => {
      render(<ConfigViewer text="yaml-content" />);
      fireEvent.click(screen.getByRole('button', { name: 'ACTION.COPY' }));
      expect(screen.getByText('ALERT.COPIED')).not.toBeNull();
    });

    it('reverts to copy button after 3 seconds', () => {
      render(<ConfigViewer text="yaml-content" />);
      fireEvent.click(screen.getByRole('button', { name: 'ACTION.COPY' }));
      expect(screen.getByRole('button', { name: 'ALERT.COPIED' })).not.toBeNull();
      act(() => {
        vi.advanceTimersByTime(3000);
      });
      expect(screen.queryByRole('button', { name: 'ALERT.COPIED' })).toBeNull();
      expect(screen.getByRole('button', { name: 'ACTION.COPY' })).not.toBeNull();
    });
  });

  describe('close button', () => {
    it('calls onClose when clicked', () => {
      const onClose = vi.fn();
      render(<ConfigViewer text="test" onClose={onClose} />);
      fireEvent.click(screen.getByRole('button', { name: 'ACTION.CLOSE' }));
      expect(onClose).toHaveBeenCalledTimes(1);
    });
  });
});
