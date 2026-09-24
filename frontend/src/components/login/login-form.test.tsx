import { render, fireEvent } from '@testing-library/react';
import { LoginForm } from './login-form';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `t:${key}`,
  }),
}));

vi.mock('lucide-react', () => ({
  EyeIcon: () => <svg data-slot="eye-icon" />,
  EyeOffIcon: () => <svg data-slot="eye-off-icon" />,
  Lock: () => <svg data-slot="lock-icon" />,
  User: () => <svg data-slot="user-icon" />,
}));

vi.mock('@/components/ui/spinner', () => ({
  Spinner: () => <div data-slot="spinner" className="animate-spin" />,
}));

describe('LoginForm', () => {
  const mockOnSubmit = vi.fn();

  beforeEach(() => {
    mockOnSubmit.mockClear();
  });

  it('renders username and password inputs', () => {
    const { container } = render(<LoginForm onSubmit={mockOnSubmit} loading={false} />);
    const username = container.querySelector<HTMLInputElement>('input[name="username"]');
    const password = container.querySelector('input[name="password"]');
    expect(username).not.toBeNull();
    expect(password).not.toBeNull();
    expect(username?.type).toBe('text');
    expect(password?.getAttribute('type')).toBe('password');
  });

  it('renders submit button', () => {
    const { container } = render(<LoginForm onSubmit={mockOnSubmit} loading={false} />);
    const button = container.querySelector('button[type="submit"]');
    expect(button).not.toBeNull();
    expect(button?.textContent).toContain('t:LOGIN.FORM.SUBMIT');
  });

  it('renders spinner when loading and disables submit', () => {
    const { container } = render(<LoginForm onSubmit={mockOnSubmit} loading={true} />);
    const spinner = container.querySelector('[data-slot="spinner"]');
    expect(spinner).not.toBeNull();
    const button = container.querySelector('button[type="submit"]');
    expect(button?.hasAttribute('disabled')).toBe(true);
  });

  it('hides spinner when not loading', () => {
    const { container } = render(<LoginForm onSubmit={mockOnSubmit} loading={false} />);
    const spinner = container.querySelector('[data-slot="spinner"]');
    expect(spinner).toBeNull();
  });

  it('toggles password visibility', () => {
    const { container } = render(<LoginForm onSubmit={mockOnSubmit} loading={false} />);
    const password = container.querySelector('input[name="password"]') as HTMLInputElement;
    expect(password?.getAttribute('type')).toBe('password');
    const toggle = Array.from(container.querySelectorAll('button')).find((btn) => {
      const svg = btn.querySelector('svg[data-slot="eye-icon"]');
      return svg !== null;
    });
    expect(toggle).toBeTruthy();

    fireEvent.click(toggle as HTMLElement);
    expect(password?.getAttribute('type')).toBe('text');

    fireEvent.click(toggle as HTMLElement);
    expect(password?.getAttribute('type')).toBe('password');
  });

  it('renders username icon and lock icon', () => {
    const { container } = render(<LoginForm onSubmit={mockOnSubmit} loading={false} />);
    expect(container.querySelector('[data-slot="user-icon"]')).not.toBeNull();
    expect(container.querySelector('[data-slot="lock-icon"]')).not.toBeNull();
  });
});
