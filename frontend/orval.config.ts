import { defineConfig } from 'orval';

export default defineConfig({
  'devops-api': {
    input: {
      target: '../spec/v1-api.yaml',
    },
    output: {
      target: './src/api/client.ts',
      schemas: './src/api/model',
      client: 'axios',
      mock: {
        type: 'msw',
        output: './src/api/mocks',
      },
    },
  },
});
