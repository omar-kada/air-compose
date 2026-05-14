/**
 * @type {import('orval/dist/config').Config}
 */
const config = {
  airCompose: {
    input: '../api/tsp-output/schema/openapi.1.0.yaml',
    output: {
      target: './src/api/api.ts',
      client: 'react-query',
      httpClient: 'axios',
    },
  },
};

module.exports = config;
