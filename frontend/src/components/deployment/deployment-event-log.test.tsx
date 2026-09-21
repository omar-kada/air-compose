import { render, screen } from '@testing-library/react';
import type { Event } from '@/api/api';
import { EventType } from '@/api/api';
import { DeploymentEventLog } from './deployment-event-log';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => `translated:${key}`,
    i18n: { language: 'en' },
  }),
}));

const mockEvent = (over: Partial<Event> = {}): Event => ({
  ID: 1,
  time: '2025-01-01T00:00:00Z',
  msg: 'event message',
  type: EventType.DEPLOYMENT_STARTED,
  objectId: 123,
  objectName: 'deploy-1',
  ...over,
});

describe('DeploymentEventLog', () => {
  it('renders the events log heading', () => {
    render(<DeploymentEventLog events={[]} />);
    expect(screen.queryByText('translated:DEPLOYMENTS.EVENTS_LOG')).not.toBeNull();
  });

  it('renders a translated label for each event type', () => {
    render(
      <DeploymentEventLog
        events={[
          mockEvent({ ID: 1, type: EventType.DEPLOYMENT_STARTED }),
          mockEvent({ ID: 2, type: EventType.ERROR, msg: 'failed' }),
        ]}
      />,
    );
    expect(screen.queryByText('translated:EVENT_TYPE.DEPLOYMENT_STARTED')).not.toBeNull();
    expect(screen.queryByText('translated:EVENT_TYPE.ERROR')).not.toBeNull();
  });

  it('renders the event message when present', () => {
    const { container } = render(
      <DeploymentEventLog events={[mockEvent({ msg: 'Something happened' })]} />,
    );
    const messageEl = container.querySelector('p.font-light');
    expect(messageEl).not.toBeNull();
    expect(messageEl?.textContent).toBe('Something happened');
  });

  it('does not render a message paragraph when the message is empty', () => {
    const { container } = render(<DeploymentEventLog events={[mockEvent({ msg: '' })]} />);
    // The event type label is still rendered
    expect(screen.queryByText('translated:EVENT_TYPE.DEPLOYMENT_STARTED')).not.toBeNull();
    // No <p class="font-light"> for the message
    expect(container.querySelector('p.font-light')).toBeNull();
  });

  it('renders the correct number of timeline items', () => {
    render(
      <DeploymentEventLog
        events={[mockEvent({ ID: 1 }), mockEvent({ ID: 2 }), mockEvent({ ID: 3 })]}
      />,
    );
    expect(screen.queryAllByText(/translated:EVENT_TYPE\./)).toHaveLength(3);
  });

  it('does not render timeline items when events list is empty', () => {
    render(<DeploymentEventLog events={[]} />);
    expect(screen.queryAllByText(/translated:EVENT_TYPE\./)).toHaveLength(0);
  });
});
