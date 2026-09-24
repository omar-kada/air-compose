import { render, screen } from '@testing-library/react';
import { LoginPage } from './login-page';

const { mockUseLogin, mockUseRegisteration } = vi.hoisted(() => ({
  mockUseLogin: vi.fn(),
  mockUseRegisteration: vi.fn(),
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `t:${key}`,
    i18n: { language: 'en' },
  }),
}));

vi.mock('@/hooks', () => ({
  useLogin: mockUseLogin,
  useRegisteration: mockUseRegisteration,
}));

vi.mock('./login', () => ({
  LoginForm: () => <div data-slot="login-form" />,
}));

vi.mock('./ui/button', () => ({
  Button: ({ onClick, children }: { onClick?: () => void; children: React.ReactNode }) => (
    <button onClick={onClick} data-slot="button">
      {children}
    </button>
  ),
}));

vi.mock('./ui/card', () => ({
  Card: ({ children }: { children: React.ReactNode }) => <div data-slot="card">{children}</div>,
  CardContent: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="card-content">{children}</div>
  ),
}));

vi.mock('./view', () => ({
  TextSeparator: ({ text }: { text: string }) => <div data-slot="text-separator">{text}</div>,
}));

describe('LoginPage', () => {
  beforeEach(() => {
    mockUseLogin.mockReturnValue({ login: vi.fn(), isPending: false });
    mockUseRegisteration.mockReturnValue({ data: undefined, isPending: false });
    vi.stubEnv('VITE_SERVER_URL', 'http://localhost:3000');
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it('renders the page title', () => {
    render(<LoginPage />);
    expect(screen.getByText('t:LOGIN.FORM.TITLE')).toBeTruthy();
  });

  it('renders the login form', () => {
    render(<LoginPage />);
    expect(document.querySelector('[data-slot="login-form"]')).not.toBeNull();
  });

  it('does not render OIDC button when registration has no oidc', () => {
    render(<LoginPage />);
    expect(screen.queryByText('t:LOGIN.FORM.LOGIN_WITH_OIDC')).toBeNull();
  });

  it('renders OIDC button when registration has oidc', () => {
    mockUseRegisteration.mockReturnValue({ data: { oidc: true }, isPending: false });
    render(<LoginPage />);
    expect(screen.getByText('t:LOGIN.FORM.LOGIN_WITH_OIDC')).toBeTruthy();
  });

  it('renders text separator when OIDC is available', () => {
    mockUseRegisteration.mockReturnValue({ data: { oidc: true }, isPending: false });
    render(<LoginPage />);
    expect(document.querySelector('[data-slot="text-separator"]')).not.toBeNull();
  });

  it('passes loading state to login form when pending', () => {
    mockUseLogin.mockReturnValue({ login: vi.fn(), isPending: true });
    render(<LoginPage />);
    expect(document.querySelector('[data-slot="login-form"]')).not.toBeNull();
  });
});
