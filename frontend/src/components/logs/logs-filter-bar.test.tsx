import { render, screen, fireEvent } from '@testing-library/react';
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
    const toggles = screen.getAllByRole('button');
    expect(toggles).toHaveLength(4);
    expect(toggles.map((b) => b.textContent?.trim())).toEqual([
      Level.DEBUG,
      Level.INFO,
      Level.WARN,
      Level.ERROR,
    ]);
  });
});
