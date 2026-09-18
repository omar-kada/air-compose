import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import type { LogLine, LogMessages } from '@/api';
import { Level } from '@/api';
import { LogEntries } from './logs-entries';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => `translated:${key}` }),
}));

vi.mock('./auto-scroll-area', () => ({
  AutoScrollArea: ({ children }: { children: ReactNode }) => children,
}));

const line = (over: Partial<LogLine> = {}): LogLine => ({
  level: Level.INFO,
  msg: 'hello',
  time: '2025-01-01T00:00:00Z',
  meta: {},
  ...over,
});

const list = (...lines: LogLine[]): LogMessages => lines;

describe('LogEntries', () => {
  it('renders the empty state when there are no logs', () => {
    render(<LogEntries logs={[]} />);
    expect(screen.queryByText('translated:LOGS.NO_LOGS_FILTERED')).not.toBeNull();
  });

  it('renders one entry per log line', () => {
    render(<LogEntries logs={list(line({ msg: 'first' }), line({ msg: 'second' }))} />);
    expect(screen.queryByText('first')).not.toBeNull();
    expect(screen.queryByText('second')).not.toBeNull();
  });

  it('shows the level badge with the level text', () => {
    render(<LogEntries logs={list(line({ level: Level.ERROR, msg: 'boom' }))} />);
    expect(screen.queryByText('ERROR')).not.toBeNull();
  });

  it('keeps meta collapsed until Expand All is toggled', async () => {
    render(<LogEntries logs={list(line({ meta: { pod: 'api-1', node: 'n1' } }))} />);
    expect(screen.queryByText('pod: api-1')).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'translated:LOGS.EXPAND_ALL' }));
    await waitFor(() => {
      expect(screen.queryByText('pod: api-1')).not.toBeNull();
      expect(screen.queryByText('node: n1')).not.toBeNull();
    });
  });
});
