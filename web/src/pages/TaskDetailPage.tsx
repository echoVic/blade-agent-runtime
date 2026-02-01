import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, Clock, FolderGit, GitBranch, FileCode, Terminal, AlertCircle } from 'lucide-react';
import { api } from '@/services/api';
import type { Task, LedgerStep } from '@/types';

export function TaskDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [task, setTask] = useState<Task | null>(null);
  const [ledger, setLedger] = useState<LedgerStep[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedDiff, setSelectedDiff] = useState<string | null>(null);
  const [diffLoading, setDiffLoading] = useState(false);

  useEffect(() => {
    if (id) {
      loadTaskData();
    }
  }, [id]);

  const loadTaskData = async () => {
    try {
      setLoading(true);
      const [taskData, ledgerData] = await Promise.all([
        api.getTask(id!),
        api.getLedger(id!),
      ]);
      setTask(taskData);
      setLedger(ledgerData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load task');
    } finally {
      setLoading(false);
    }
  };

  const loadDiff = async (stepId: string) => {
    if (!id) return;
    try {
      setDiffLoading(true);
      const diff = await api.getDiff(id, stepId);
      setSelectedDiff(diff);
    } catch (err) {
      setSelectedDiff('Failed to load diff');
    } finally {
      setDiffLoading(false);
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString();
  };

  const getStepIcon = (kind: string) => {
    switch (kind) {
      case 'run':
        return <Terminal className="w-4 h-4" />;
      case 'apply':
        return <GitBranch className="w-4 h-4" />;
      case 'rollback':
        return <AlertCircle className="w-4 h-4" />;
      default:
        return <FileCode className="w-4 h-4" />;
    }
  };

  const getStepColor = (kind: string) => {
    switch (kind) {
      case 'run':
        return 'bg-blue-100 text-blue-700';
      case 'apply':
        return 'bg-green-100 text-green-700';
      case 'rollback':
        return 'bg-red-100 text-red-700';
      default:
        return 'bg-gray-100 text-gray-700';
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (error || !task) {
    return (
      <div className="p-4 bg-red-50 text-red-700 rounded-lg">
        Error: {error || 'Task not found'}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          to="/"
          className="p-2 hover:bg-muted rounded-lg transition-colors"
        >
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold">{task.name}</h1>
          <div className="flex items-center gap-4 text-sm text-muted-foreground">
            <span>ID: {task.id}</span>
            <span className={`px-2 py-0.5 rounded-full text-xs ${task.status === 'active' ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-700'}`}>
              {task.status}
            </span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="p-4 bg-card rounded-lg border border-border">
          <div className="flex items-center gap-2 text-muted-foreground mb-2">
            <FolderGit className="w-4 h-4" />
            <span className="text-sm">Branch</span>
          </div>
          <code className="text-sm">{task.branch}</code>
        </div>
        <div className="p-4 bg-card rounded-lg border border-border">
          <div className="flex items-center gap-2 text-muted-foreground mb-2">
            <Clock className="w-4 h-4" />
            <span className="text-sm">Created</span>
          </div>
          <span className="text-sm">{formatDate(task.created_at)}</span>
        </div>
      </div>

      <div>
        <h2 className="text-lg font-semibold mb-4">Ledger ({ledger.length} steps)</h2>
        <div className="space-y-3">
          {ledger.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              No steps recorded yet
            </div>
          ) : (
            ledger.map((step) => (
              <div
                key={step.step_id}
                className="p-4 bg-card rounded-lg border border-border"
              >
                <div className="flex items-start justify-between">
                  <div className="space-y-2">
                    <div className="flex items-center gap-2">
                      <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs ${getStepColor(step.kind)}`}>
                        {getStepIcon(step.kind)}
                        {step.kind}
                      </span>
                      <span className="text-sm text-muted-foreground">
                        {formatDate(step.started_at)}
                      </span>
                      {step.duration_ms > 0 && (
                        <span className="text-sm text-muted-foreground">
                          ({Math.round(step.duration_ms / 1000)}s)
                        </span>
                      )}
                    </div>
                    {step.cmd && (
                      <code className="block text-sm bg-muted p-2 rounded">
                        {step.cmd.join(' ')}
                      </code>
                    )}
                    {step.diff_stat && (
                      <div className="flex items-center gap-4 text-sm">
                        <span className="text-muted-foreground">
                          {step.diff_stat.files} files
                        </span>
                        <span className="text-green-600">
                          +{step.diff_stat.additions}
                        </span>
                        <span className="text-red-600">
                          -{step.diff_stat.deletions}
                        </span>
                      </div>
                    )}
                    {step.policy_events && step.policy_events.length > 0 && (
                      <div className="flex flex-wrap gap-2">
                        {step.policy_events.map((event, idx) => (
                          <span
                            key={idx}
                            className="text-xs px-2 py-1 bg-yellow-100 text-yellow-800 rounded"
                          >
                            {event.rule}: {event.action}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                  {step.artifacts?.patch && (
                    <button
                      onClick={() => loadDiff(step.step_id)}
                      className="text-sm text-primary hover:underline"
                    >
                      View Diff
                    </button>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      </div>

      {selectedDiff && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50">
          <div className="bg-card rounded-lg w-full max-w-4xl max-h-[80vh] overflow-hidden flex flex-col">
            <div className="flex items-center justify-between p-4 border-b border-border">
              <h3 className="font-semibold">Diff</h3>
              <button
                onClick={() => setSelectedDiff(null)}
                className="p-2 hover:bg-muted rounded-lg"
              >
                ✕
              </button>
            </div>
            <div className="p-4 overflow-auto flex-1">
              {diffLoading ? (
                <div className="flex items-center justify-center py-8">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                </div>
              ) : (
                <pre className="text-sm font-mono whitespace-pre-wrap">{selectedDiff}</pre>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
