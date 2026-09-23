import { render, screen, fireEvent } from '@testing-library/react';
import { ThemeToggle } from './theme-toggle';

const mockSetTheme = vi.hoisted(() => vi.fn());
const mockUseTheme = vi.hoisted(() => vi.fn());

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('@/hooks/theme-provider', () => ({
  useTheme: mockUseTheme,
}));

describe('ThemeToggle', () => {
  beforeEach(() => {
    mockSetTheme.mockClear();
  });

  it('renders button with sr-only translated label', () => {
    mockUseTheme.mockReturnValue({ theme: 'light', setTheme: mockSetTheme });
    render(<ThemeToggle />);
    expect(screen.getByRole('button', { name: 'translated:TOGGLE_THEME' })).toBeTruthy();
  });

  it('toggles from light to dark on click', () => {
    mockUseTheme.mockReturnValue({ theme: 'light', setTheme: mockSetTheme });
    render(<ThemeToggle />);
    fireEvent.click(screen.getByRole('button'));
    expect(mockSetTheme).toHaveBeenCalledWith('dark');
  });

  it('toggles from dark to light on click', () => {
    mockUseTheme.mockReturnValue({ theme: 'dark', setTheme: mockSetTheme });
    render(<ThemeToggle />);
    fireEvent.click(screen.getByRole('button'));
    expect(mockSetTheme).toHaveBeenCalledWith('light');
  });
});
