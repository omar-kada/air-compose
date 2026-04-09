import { ContainerHealth, DeploymentStatus, type StackStatus, type State } from '@/api/api';
import { http } from 'msw';

const mockStatus: StackStatus = {
  name: 'homepage',
  stackId: 'stack-1',
  services: [
    {
      containerId: '1',
      state: 'running',
      name: 'web-container',
      health: 'healthy',
      startedAt: `${new Date()}`,
    },
  ],
};

const mockState: State = {
  nextDeploy: new Date().toString(),
  status: DeploymentStatus.success,
  health: ContainerHealth.healthy,
  initialized: true,
};

export const handlers = [
  http.get('/api/status', () => {
    return new Response(JSON.stringify(mockStatus), {
      status: 200,
    });
  }),
  http.get('/api/state/:days', () => {
    return new Response(JSON.stringify(mockState), {
      status: 200,
    });
  }),
];
