export interface Task {
  id: string;
  name: string;
  repo_root: string;
  base_ref: string;
  branch: string;
  workspace_path: string;
  status: 'active' | 'closed';
  created_at: string;
  updated_at: string;
  closed_at?: string;
  is_active?: boolean;
}

export interface LedgerStep {
  step_id: string;
  kind: 'run' | 'apply' | 'rollback';
  started_at: string;
  ended_at: string;
  duration_ms: number;
  cmd?: string[];
  cwd?: string;
  exit_code?: number;
  diff_stat?: {
    files: number;
    additions: number;
    deletions: number;
    file_list?: string[];
  };
  artifacts?: {
    patch?: string;
    output?: string;
  };
  policy_events?: Array<{
    rule: string;
    action: string;
    matched: string;
  }>;
  mode?: string;
  commit_sha?: string;
  commit_message?: string;
  target_branch?: string;
  target?: string;
  target_step?: string;
  hard?: boolean;
}

export interface Status {
  active_task_id: string;
  active_task?: Task;
}

export interface WebSocketMessage {
  type: string;
  data: unknown;
}

export interface LiveDiffData {
  task_id: string;
  files: number;
  additions: number;
  deletions: number;
  file_list: string[];
  patch: string;
}
