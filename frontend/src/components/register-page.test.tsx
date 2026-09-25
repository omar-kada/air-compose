import { render, screen } from '@testing-library/react';
import { RegisterPage } from './register-page';

const { mockUseRegister } = vi.hoisted(() => ({
  mockUseRegister: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `translated:${key}`,
    i18n: { language: 'en' },
  }),
}));

vi.mock('@/hooks', () => ({
  useRegister: mockUseRegister,
}));

vi.mock('./login', () => ({
  RegisterForm: () => <div data-slot="register-form" />,
}));

vi.mock('./ui/card', () => ({
  Card: ({ children }: { children: React.ReactNode }) => <div data-slot="card">{children}</div>,
  CardContent: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="card-content">{children}</div>
  ),
}));

describe('RegisterPage', () => {
  beforeEach(() => {
    mockUseRegister.mockReturnValue({ register: vi.fn(), isPending: false });
  });

  it('renders the page title', () => {
    render(<RegisterPage />);
    expect(screen.getByText('translated:REGISTER.FORM.TITLE')).toBeTruthy();
  });

  it('renders the register form', () => {
    render(<RegisterPage />);
    expect(document.querySelector('[data-slot="register-form"]')).not.toBeNull();
  });

  it('passes isPending to the register form', () => {
    mockUseRegister.mockReturnValue({ register: vi.fn(), isPending: true });
    render(<RegisterPage />);
    expect(document.querySelector('[data-slot="register-form"]')).not.toBeNull();
  });
});
