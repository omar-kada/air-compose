import { render } from '@testing-library/react';
import { ContainerHealth, EventType, type Event } from '@/api/api';
import { NotificationBadge } from './notification-badge';

const mockEvent = (over: Partial<Event> = {}): Event => ({
  ID: 1,
  time: '2025-01-01T00:00:00Z',
  msg: '',
  type: EventType.DEPLOYMENT_STARTED,
  objectId: 123,
  objectName: 'deploy-1',
  ...over,
});

describe('NotificationBadge', () => {
  it('renders an SVG icon', () => {
    const { container } = render(<NotificationBadge notification={mockEvent()} />);
    expect(container.querySelector('svg')).not.toBeNull();
  });

  it.each<[EventType, string]>([
    [EventType.ERROR, 'bg-destructive'],
    [EventType.DEPLOYMENT_ERROR, 'bg-destructive'],
    [EventType.MISC, 'bg-blue-500'],
    [EventType.DEPLOYMENT_STARTED, 'bg-blue-500'],
    [EventType.DEPLOYMENT_SUCCESS, 'bg-green-500'],
    [EventType.PASSWORD_UPDATED, 'bg-yellow-500'],
    [EventType.CONFIGURATION_UPDATED, 'bg-yellow-500'],
    [EventType.SESSION_REUSED, 'bg-purple-500'],
  ])('applies %s color class', (type, colorClass) => {
    const { container } = render(<NotificationBadge notification={mockEvent({ type, msg: '' })} />);
    expect((container.firstChild as HTMLElement).getAttribute('class')).toContain(colorClass);
  });

  it('applies bg-destructive for unhealthy health change', () => {
    const { container } = render(
      <NotificationBadge
        notification={mockEvent({
          type: EventType.HEALTH_CHANGE,
          msg: `Status is ${ContainerHealth.unhealthy}`,
        })}
      />,
    );
    expect((container.firstChild as HTMLElement).getAttribute('class')).toContain('bg-destructive');
  });

  it('applies bg-green-500 for healthy health change', () => {
    const { container } = render(
      <NotificationBadge
        notification={mockEvent({
          type: EventType.HEALTH_CHANGE,
          msg: `Status is ${ContainerHealth.healthy}`,
        })}
      />,
    );
    expect((container.firstChild as HTMLElement).getAttribute('class')).toContain('bg-green-500');
  });

  it('applies bg-yellow-500 for other health change message', () => {
    const { container } = render(
      <NotificationBadge
        notification={mockEvent({
          type: EventType.HEALTH_CHANGE,
          msg: 'some other message',
        })}
      />,
    );
    expect((container.firstChild as HTMLElement).getAttribute('class')).toContain('bg-yellow-500');
  });

  it('applies bg-gray-500 for unknown event type', () => {
    const { container } = render(
      <NotificationBadge notification={mockEvent({ type: 'UNKNOWN' as EventType, msg: '' })} />,
    );
    expect((container.firstChild as HTMLElement).getAttribute('class')).toContain('bg-gray-500');
  });
});
