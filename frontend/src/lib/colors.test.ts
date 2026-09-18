import { EventType } from '@/api/api';
import type { ContainerHealth } from '@/api/api';
import { colorForStatus, borderForStatus, textColorForStatus, logColor } from './colors';

describe('colorForStatus', () => {
  it('maps healthy/success to a green class', () => {
    expect(colorForStatus('healthy')).toMatch(/green/);
    expect(colorForStatus('success')).toMatch(/green/);
  });

  it('maps unhealthy/error to a red class', () => {
    expect(colorForStatus('unhealthy')).toMatch(/red/);
    expect(colorForStatus('error')).toMatch(/red/);
  });

  it('maps starting/planned to a slate class', () => {
    expect(colorForStatus('starting')).toMatch(/slate/);
    expect(colorForStatus('planned')).toMatch(/slate/);
  });

  it('maps running to a blue class', () => {
    expect(colorForStatus('running')).toMatch(/blue/);
  });

  it('returns empty string for unknown statuses', () => {
    expect(colorForStatus('unknown' as ContainerHealth)).toBe('');
  });
});

describe('borderForStatus', () => {
  it('maps running/healthy to a green class', () => {
    expect(borderForStatus('running')).toMatch(/green/);
    expect(borderForStatus('healthy')).toMatch(/green/);
  });

  it('maps dead/removing/unhealthy to a red class', () => {
    expect(borderForStatus('dead')).toMatch(/red/);
    expect(borderForStatus('removing')).toMatch(/red/);
    expect(borderForStatus('unhealthy')).toMatch(/red/);
  });

  it('maps exited/paused/none to a slate class', () => {
    expect(borderForStatus('exited')).toMatch(/slate/);
    expect(borderForStatus('paused')).toMatch(/slate/);
    expect(borderForStatus('none')).toMatch(/slate/);
  });

  it('maps created/restarting/starting to a blue class', () => {
    expect(borderForStatus('created')).toMatch(/blue/);
    expect(borderForStatus('restarting')).toMatch(/blue/);
    expect(borderForStatus('starting')).toMatch(/blue/);
  });

  it('returns empty string when called with no status', () => {
    expect(borderForStatus()).toBe('');
  });
});

describe('textColorForStatus', () => {
  it('maps running/healthy to a green class', () => {
    expect(textColorForStatus('running')).toMatch(/green/);
    expect(textColorForStatus('healthy')).toMatch(/green/);
  });

  it('maps dead/removing/unhealthy to a red class', () => {
    expect(textColorForStatus('dead')).toMatch(/red/);
    expect(textColorForStatus('removing')).toMatch(/red/);
    expect(textColorForStatus('unhealthy')).toMatch(/red/);
  });

  it('maps exited/paused/none to a slate class', () => {
    expect(textColorForStatus('exited')).toMatch(/slate/);
    expect(textColorForStatus('paused')).toMatch(/slate/);
    expect(textColorForStatus('none')).toMatch(/slate/);
  });

  it('maps created/restarting/starting to a blue class', () => {
    expect(textColorForStatus('created')).toMatch(/blue/);
    expect(textColorForStatus('restarting')).toMatch(/blue/);
    expect(textColorForStatus('starting')).toMatch(/blue/);
  });

  it('returns empty string when called with no status', () => {
    expect(textColorForStatus()).toBe('');
  });
});

describe('logColor', () => {
  it('maps ERROR and DEPLOYMENT_ERROR to a red class', () => {
    expect(logColor(EventType.ERROR)).toMatch(/red/);
    expect(logColor(EventType.DEPLOYMENT_ERROR)).toMatch(/red/);
  });

  it('maps MISC to a gray class', () => {
    expect(logColor(EventType.MISC)).toMatch(/gray/);
  });

  it('returns empty string for unhandled event types', () => {
    expect(logColor(EventType.DEPLOYMENT_SUCCESS)).toBe('');
    expect(logColor(EventType.DEPLOYMENT_STARTED)).toBe('');
    expect(logColor(EventType.HEALTH_CHANGE)).toBe('');
    expect(logColor(EventType.CONFIGURATION_UPDATED)).toBe('');
    expect(logColor(EventType.PASSWORD_UPDATED)).toBe('');
    expect(logColor(EventType.SESSION_REUSED)).toBe('');
  });
});
