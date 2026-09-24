import { render, screen, fireEvent } from '@testing-library/react';
import { RegisterForm } from './register-form';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `t:${key}`,
  }),
}));

vi.mock('@/components/ui/spinner', () => ({
  Spinner: () => <div data-slot="spinner" className="animate-spin" />,
}));

describe('RegisterForm', () => {
  const mockOnSubmit = vi.fn();

  beforeEach(() => {
    mockOnSubmit.mockClear();
  });

  it('renders username and password inputs', () => {
    const { container } = render(<RegisterForm onSubmit={mockOnSubmit} loading={false} />);
    const username = container.querySelector<HTMLInputElement>('input[name="username"]');
    const password = container.querySelector('input[name="password"]');
    expect(username).not.toBeNull();
    expect(password).not.toBeNull();
    expect(username?.type).toBe('text');
    expect(password?.getAttribute('type')).toBe('password');
  });

  it('renders submit button', () => {
    render(<RegisterForm onSubmit={mockOnSubmit} loading={false} />);
    expect(screen.getByRole('button', { name: 't:REGISTER.FORM.SUBMIT' })).toBeTruthy();
  });

  it('renders spinner when loading and disables submit', () => {
    const { container } = render(<RegisterForm onSubmit={mockOnSubmit} loading={true} />);
    expect(container.querySelector('[data-slot="spinner"]')).not.toBeNull();
    const button = screen.getByRole('button', { name: 't:REGISTER.FORM.SUBMIT' });
    expect(button.hasAttribute('disabled')).toBe(true);
  });

  it('hides spinner when not loading', () => {
    const { container } = render(<RegisterForm onSubmit={mockOnSubmit} loading={false} />);
    expect(container.querySelector('[data-slot="spinner"]')).toBeNull();
  });

  it('toggles password visibility', () => {
    const { container } = render(<RegisterForm onSubmit={mockOnSubmit} loading={false} />);
    const password = container.querySelector('input[name="password"]') as HTMLInputElement;
    expect(password?.getAttribute('type')).toBe('password');
    const toggle = screen.getByRole('button', { name: 'Show password' });
    fireEvent.click(toggle);
    expect(password?.getAttribute('type')).toBe('text');
    fireEvent.click(toggle);
    expect(password?.getAttribute('type')).toBe('password');
  });
});
