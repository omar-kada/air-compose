import type { LogLine } from '@/api';
import { useLogs } from '@/hooks';

export function LogsPage() {
  const { data: logs = [] } = useLogs(50); // startLogs on mount, endLogs on unmount

  return (
    <div>
      <ul>
        {logs.map((line: LogLine, i: number) => (
          <li key={i}>
            [{line.level}] {new Date(line.time).toLocaleTimeString()} {line.msg}
          </li>
        ))}
      </ul>
    </div>
  );
}
