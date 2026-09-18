import { render, screen, fireEvent, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Level } from '@/api';
import { LogFilterBar } from './logs-filter-bar';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => `t:${key}` }),
}));

const makeProps = () => ({
  text: 'foo',
  onTextChange: vi.fn(),
  activeLevels: new Set([Level.DEBUG]),
  onLevelsChange: vi.fn(),
});

describe('LogFilterBar', () => {
  it('renders the text input with the current value', () => {
    render(<LogFilterBar {...makeProps()} />);
    expect(screen.queryByDisplayValue('foo')).not.toBeNull();
  });

  it('notifies onTextChange when the text input changes', () => {
    const props = makeProps();
    render(<LogFilterBar {...props} />);
    const input = screen.getByPlaceholderText('t:LOGS.FILTER_LOGS');
    fireEvent.change(input, { target: { value: 'bar' } });
    expect(props.onTextChange).toHaveBeenCalledWith('bar');
  });

  it('renders a toggle for each level', () => {
    render(<LogFilterBar {...makeProps()} />);
    const toggles = within(screen.getByRole('toolbar')).getAllByRole('button');
    expect(toggles).toHaveLength(4);
    expect(toggles.map((b) => b.textContent?.trim())).toEqual([
      Level.DEBUG,
      Level.INFO,
      Level.WARN,
      Level.ERROR,
    ]);
  });

  it('notifies onLevelsChange with a Set when a toggle is clicked', async () => {
    const user = userEvent.setup();
    const props = makeProps(); // activeLevels = { DEBUG }
    render(<LogFilterBar {...props} />);
    const [, infoToggle] = within(screen.getByRole('toolbar')).getAllByRole('button');
    await user.click(infoToggle);
    expect(props.onLevelsChange).toHaveBeenCalledTimes(1);
    const next = props.onLevelsChange.mock.calls[0][0] as Set<Level>;
    expect(next).toBeInstanceOf(Set);
    expect([...next]).toEqual([Level.DEBUG, Level.INFO]);
  });
});
