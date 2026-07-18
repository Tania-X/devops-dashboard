import { setupWorker } from 'msw/browser';
import { getDevOpsDashboardAPIMock } from '../api/client';

export const worker = setupWorker(...getDevOpsDashboardAPIMock());
