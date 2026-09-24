import { render } from '@testing-library/react';
import { useForm, useFieldArray } from 'react-hook-form';
import type { FormValues } from './config-form-schema';
import { ServiceCard } from './service-card';

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

vi.mock('lucide-react', () => ({
  Trash2: () => <svg data-slot="trash-icon" />,
}));

vi.mock('@/lib', () => ({
  ServiceLogo: ({ service }: { service: string }) => (
    <div data-slot="service-logo" data-service={service} />
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
  CardAction: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="card-action">{children}</div>
  ),
  CardContent: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="card-content">{children}</div>
  ),
  CardHeader: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="card-header">{children}</div>
  ),
  CardTitle: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="card-title">{children}</div>
  ),
}));

vi.mock('../ui/input', () => ({
  Input: (props: Record<string, unknown>) => <input data-slot="input" {...props} />,
}));

vi.mock('../ui/field', () => ({
  Field: ({ children }: { children: React.ReactNode }) => <div data-slot="field">{children}</div>,
  FieldError: ({ errors }: { errors?: Array<{ message?: string }> }) =>
    errors?.length ? <div data-slot="field-error" /> : null,
  FieldGroup: ({ children }: { children: React.ReactNode }) => (
    <div data-slot="field-group">{children}</div>
  ),
}));

function TestWrapper({ disabled }: { disabled?: boolean }) {
  const form = useForm<FormValues>({
    defaultValues: {
      services: [{ name: 'web', envVars: [{ key: '', value: '' }] }],
      globalEnvVars: [],
    },
  });
  const { fields } = useFieldArray({ control: form.control, name: 'services' });
  return (
    <ServiceCard
      form={form}
      service={fields[0]}
      name={`services.0`}
      disabled={disabled}
      onRemove={vi.fn()}
    />
  );
}

describe('ServiceCard', () => {
  it('renders service logo with service name', () => {
    const { container } = render(<TestWrapper />);
    const logo = container.querySelector('[data-slot="service-logo"]');
    expect(logo).not.toBeNull();
    expect(logo?.getAttribute('data-service')).toBe('web');
  });

  it('renders env vars form', () => {
    const { container } = render(<TestWrapper />);
    expect(container.querySelector('[data-slot="env-vars-array-form"]')).not.toBeNull();
  });

  it('renders remove button when not disabled', () => {
    const { container } = render(<TestWrapper />);
    expect(container.querySelector('[data-slot="button"]')).not.toBeNull();
    expect(container.querySelector('[data-slot="trash-icon"]')).not.toBeNull();
  });

  it('hides remove button when disabled', () => {
    const { container } = render(<TestWrapper disabled={true} />);
    expect(container.querySelector('[data-slot="button"]')).toBeNull();
  });

  it('renders service name input with placeholder', () => {
    const { container } = render(<TestWrapper />);
    const input = container.querySelector('[data-slot="input"]');
    expect(input).not.toBeNull();
    expect(input?.getAttribute('placeholder')).toBe('translated:CONFIGURATION.FORM.SERVICE_NAME');
  });
});
