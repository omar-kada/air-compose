import { render, fireEvent } from '@testing-library/react';
import { useForm } from 'react-hook-form';
import type { FormValues } from './config-form-schema';
import { ConfigForm } from './config-form';

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('lucide-react', () => ({
  Plus: () => <svg data-slot="plus-icon" />,
}));

vi.mock('@/lib', () => ({
  cn: (...args: unknown[]) => args.filter(Boolean).join(' '),
}));

vi.mock('./service-card', () => ({
  ServiceCard: ({ service, onRemove }: { service: { name: string }; onRemove: () => void }) => (
    <div data-slot="service-card" data-name={service.name}>
      <button data-slot="remove-service" onClick={onRemove}>
        Remove
      </button>
    </div>
  ),
}));

vi.mock('./env-vars-array-form', () => ({
  EnvVarArrayForm: () => <div data-slot="env-vars-array-form" />,
}));

vi.mock('@/components/ui/button', () => ({
  Button: ({
    onClick,
    children,
    disabled,
  }: {
    onClick?: () => void;
    children: React.ReactNode;
    disabled?: boolean;
  }) => (
    <button onClick={onClick} disabled={disabled} data-slot="button">
      {children}
    </button>
  ),
}));

vi.mock('@/components/ui/card', () => ({
  Card: ({ children }: { children: React.ReactNode }) => <div data-slot="card">{children}</div>,
  CardContent: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="card-content">{children}</div>
  ),
}));

function TestWrapper({ disabled }: { disabled?: boolean }) {
  const form = useForm<FormValues>({
    defaultValues: {
      services: [{ name: 'web', envVars: [{ key: '', value: '' }] }],
      globalEnvVars: [],
    },
  });
  return <ConfigForm form={form} disabled={disabled} />;
}

describe('ConfigForm', () => {
  it('renders service cards for each service', () => {
    const { container } = render(<TestWrapper />);
    const cards = container.querySelectorAll('[data-slot="service-card"]');
    expect(cards).toHaveLength(1);
    expect(cards[0]?.getAttribute('data-name')).toBe('web');
  });

  it('renders global env vars form', () => {
    const { container } = render(<TestWrapper />);
    expect(container.querySelector('[data-slot="env-vars-array-form"]')).not.toBeNull();
  });

  it('renders add service button when not disabled', () => {
    const { container } = render(<TestWrapper />);
    expect(container.querySelector('[data-slot="button"]')).not.toBeNull();
    expect(container.textContent).toContain('translated:CONFIGURATION.FORM.ADD_SERVICE');
  });

  it('hides add service button when disabled', () => {
    const { container } = render(<TestWrapper disabled={true} />);
    expect(container.querySelector('[data-slot="button"]')).toBeNull();
  });

  it('renders multiple service cards', () => {
    function TestWrapperMulti() {
      const form = useForm<FormValues>({
        defaultValues: {
          services: [
            { name: 'web', envVars: [{ key: '', value: '' }] },
            { name: 'db', envVars: [{ key: '', value: '' }] },
          ],
          globalEnvVars: [],
        },
      });
      return <ConfigForm form={form} />;
    }
    const { container } = render(<TestWrapperMulti />);
    expect(container.querySelectorAll('[data-slot="service-card"]')).toHaveLength(2);
  });

  it('adds a new service when Add Service button is clicked', () => {
    const { container } = render(<TestWrapper />);
    fireEvent.click(container.querySelector('[data-slot="button"]') as HTMLElement);
    expect(container.querySelectorAll('[data-slot="service-card"]')).toHaveLength(2);
  });

  it('removes a service when Remove button is clicked', () => {
    const { container } = render(<TestWrapper />);
    fireEvent.click(container.querySelector('[data-slot="remove-service"]') as HTMLElement);
    expect(container.querySelectorAll('[data-slot="service-card"]')).toHaveLength(0);
  });
});
