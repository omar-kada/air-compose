import { render, screen } from '@testing-library/react';
import type { AxiosError } from 'axios';
import type { Error as ApiError } from '@/api';
import { ErrorAlert } from './error-alert';

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

describe('ErrorAlert', () => {
  it('renders nothing when error is null', () => {
    const { container } = render(<ErrorAlert title="ERROR.TITLE" error={null} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders translated title and message for plain Error', () => {
    const error: ApiError = { code: 'SERVER_ERROR', message: 'Something broke' };
    render(<ErrorAlert title="ERROR.TITLE" error={error} />);
    expect(screen.getByText('translated:ERROR.TITLE')).toBeTruthy();
    expect(screen.getByText('translated:Something broke')).toBeTruthy();
  });

  it('uses response.data.message for AxiosError', () => {
    const axiosError = {
      response: { data: { message: 'Server error' } },
      message: 'Network Error',
    } as unknown as AxiosError<ApiError>;
    render(<ErrorAlert title="ERROR.TITLE" error={axiosError} />);
    expect(screen.getByText('translated:Server error')).toBeTruthy();
  });

  it('falls back to error.message when response data has no message', () => {
    const axiosError = {
      response: { data: {} },
      message: 'Fallback error',
    } as unknown as AxiosError<ApiError>;
    render(<ErrorAlert title="ERROR.TITLE" error={axiosError} />);
    expect(screen.getByText('translated:Fallback error')).toBeTruthy();
  });

  it('renders only title when error message is empty', () => {
    const axiosError = {
      response: { data: { message: '' } },
      message: '',
    } as unknown as AxiosError<ApiError>;
    render(<ErrorAlert title="ERROR.TITLE" error={axiosError} />);
    expect(screen.getByText('translated:ERROR.TITLE')).toBeTruthy();
    expect(document.querySelector('[data-slot="alert-description"]')).toBeNull();
  });
});
