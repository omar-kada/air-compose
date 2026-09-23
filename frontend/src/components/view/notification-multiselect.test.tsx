import { render, screen } from '@testing-library/react';
import { useForm } from 'react-hook-form';
import { NotificationMultiSelect } from './notification-multiselect';

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

interface FormValues {
  eventType: string[];
}

function FormWrapper({ label }: { label: string }) {
  const form = useForm<FormValues>({ defaultValues: { eventType: [] } });
  return <NotificationMultiSelect name="eventType" form={form} label={label} />;
}

describe('NotificationMultiSelect', () => {
  it('renders field title with translated label', () => {
    render(<FormWrapper label="NOTIFY.LABEL" />);
    expect(screen.getByText('translated:NOTIFY.LABEL')).toBeTruthy();
  });

  it('renders checkbox groups with translated titles and options', () => {
    render(<FormWrapper label="NOTIFY.LABEL" />);
    expect(screen.getByText('translated:EVENT_TYPE.DEPLOYMENT')).toBeTruthy();
    expect(screen.getByText('translated:EVENT_TYPE.ERROR')).toBeTruthy();
    expect(screen.getByText('translated:EVENT_TYPE.DEPLOYMENT_STARTED')).toBeTruthy();
  });
});
