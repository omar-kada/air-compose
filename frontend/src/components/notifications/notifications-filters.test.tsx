import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { EventType } from '@/api/api';
import { NotificationFilter } from './notifications-filters';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `translated:${key}`,
    i18n: { language: 'en' },
  }),
}));

describe('NotificationFilter', () => {
  it('renders three toggle buttons', () => {
    render(<NotificationFilter onFilterChanged={() => {}} />);
    expect(screen.getAllByRole('button')).toHaveLength(3);
  });

  it('calls onFilterChanged with all event types on mount', async () => {
    const onFilterChanged = vi.fn();
    render(<NotificationFilter onFilterChanged={onFilterChanged} />);
    await waitFor(() => {
      expect(onFilterChanged).toHaveBeenCalledWith(Object.values(EventType));
    });
  });

  it('calls onFilterChanged with filtered types when a filter is toggled', async () => {
    const onFilterChanged = vi.fn();
    render(<NotificationFilter onFilterChanged={onFilterChanged} />);
    await waitFor(() => {
      expect(onFilterChanged).toHaveBeenCalledWith(Object.values(EventType));
    });
    const toggles = screen.getAllByRole('button');
    const errorToggle = toggles.find(
      (t) => t.getAttribute('aria-describedby') === 'translated:EVENT_TYPE.ERROR',
    );
    expect(errorToggle).toBeDefined();
    fireEvent.click(errorToggle as HTMLElement);
    await waitFor(() => {
      expect(onFilterChanged).toHaveBeenLastCalledWith([
        EventType.DEPLOYMENT_ERROR,
        EventType.ERROR,
      ]);
    });
  });

  it('resets to all types when the only filter is deselected', async () => {
    const onFilterChanged = vi.fn();
    render(<NotificationFilter onFilterChanged={onFilterChanged} />);
    await waitFor(() => {
      expect(onFilterChanged).toHaveBeenCalledWith(Object.values(EventType));
    });
    const toggles = screen.getAllByRole('button');
    const errorToggle = toggles.find(
      (t) => t.getAttribute('aria-describedby') === 'translated:EVENT_TYPE.ERROR',
    );
    fireEvent.click(errorToggle as HTMLElement);
    await waitFor(() => {
      expect(onFilterChanged).toHaveBeenLastCalledWith([
        EventType.DEPLOYMENT_ERROR,
        EventType.ERROR,
      ]);
    });
    fireEvent.click(errorToggle as HTMLElement);
    await waitFor(() => {
      expect(onFilterChanged).toHaveBeenLastCalledWith(Object.values(EventType));
    });
  });
});
