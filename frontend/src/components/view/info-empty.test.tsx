import { render, screen } from '@testing-library/react';
import { InfoEmpty } from './info-empty';

vi.mock('react-i18next', async () => {
  const { createI18nMock } = await import('@/tests/mock-factories');
  return createI18nMock();
});

describe('InfoEmpty', () => {
  it('renders nothing when title is null', () => {
    const { container } = render(<InfoEmpty title={null} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders title and details when provided', () => {
    render(<InfoEmpty title="TITLE" details="DETAILS" />);
    expect(screen.getByText('translated:TITLE')).toBeTruthy();
    expect(screen.getByText('translated:DETAILS')).toBeTruthy();
  });

  it('renders children when title is provided', () => {
    render(
      <InfoEmpty title="TITLE">
        <button>Action</button>
      </InfoEmpty>,
    );
    expect(screen.getByRole('button', { name: 'Action' })).toBeTruthy();
  });

  it('does not render details section when not provided', () => {
    render(<InfoEmpty title="TITLE" />);
    expect(document.querySelector('[data-slot="empty-description"]')).toBeNull();
  });
});
