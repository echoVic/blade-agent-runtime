import type { Task, LedgerStep, Status } from '@/types';

const API_BASE = '/api';

async function fetchJSON<T>(url: string): Promise<T> {
  const response = await fetch(`${API_BASE}${url}`);
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
  }
  return response.json();
}

export const api = {
  getHealth: () => fetchJSON<{ status: string; time: string }>('/health'),
  
  getTasks: () => fetchJSON<Task[]>('/tasks'),
  
  getTask: (id: string) => fetchJSON<Task>(`/tasks/${id}`),
  
  getLedger: (taskId: string) => fetchJSON<LedgerStep[]>(`/ledger/${taskId}`),
  
  getDiff: async (taskId: string, stepId: string): Promise<string> => {
    const response = await fetch(`${API_BASE}/diff/${taskId}/${stepId}`);
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    return response.text();
  },
  
  getStatus: () => fetchJSON<Status>('/status'),
};
