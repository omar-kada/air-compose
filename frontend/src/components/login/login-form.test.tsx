import { render, screen, fireEvent } from '@testing-library/react';
import { LoginForm } from './login-form';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `t:${key}`,
  }),
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
    render(<LoginForm onSubmit={mockOnSubmit} loading={false} />);
    expect(screen.getByRole('button', { name: 't:LOGIN.FORM.SUBMIT' })).toBeTruthy();
  });

  it('renders spinner when loading and disables submit', () => {
    render(<LoginForm onSubmit={mockOnSubmit} loading />);
    expect(screen.getByRole('status')).toBeTruthy();
    const button = screen.getByRole('button', { name: 't:LOGIN.FORM.SUBMIT' });
    expect(button.hasAttribute('disabled')).toBe(true);
  });

  it('hides spinner when not loading', () => {
    render(<LoginForm onSubmit={mockOnSubmit} loading={false} />);
    expect(screen.queryByRole('status')).toBeNull();
  });

  it('toggles password visibility', () => {
    const { container } = render(<LoginForm onSubmit={mockOnSubmit} loading={false} />);
    const password = container.querySelector('input[name="password"]') as HTMLInputElement;
    expect(password?.getAttribute('type')).toBe('password');
    const toggle = screen.getByRole('button', { name: 'Show password' });
    fireEvent.click(toggle);
    expect(password?.getAttribute('type')).toBe('text');
    fireEvent.click(toggle);
    expect(password?.getAttribute('type')).toBe('password');
  });

  it('renders form fields with icons', () => {
    render(<LoginForm onSubmit={mockOnSubmit} loading={false} />);
    expect(screen.getByText('t:LOGIN.FORM.username')).toBeTruthy();
    expect(screen.getByText('t:LOGIN.FORM.password')).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Show password' })).toBeTruthy();
  });
});
