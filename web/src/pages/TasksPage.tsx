import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { Clock, FolderGit, CheckCircle, Activity } from 'lucide-react';
import { api } from '@/services/api';
import type { Task } from '@/types';

export function TasksPage() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadTasks();
  }, []);

  const loadTasks = async () => {
    try {
      setLoading(true);
      const data = await api.getTasks();
      setTasks(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load tasks');
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString();
  };

  const formatDuration = (start: string, end?: string) => {
    const startDate = new Date(start);
    const endDate = end ? new Date(end) : new Date();
    const diff = Math.floor((endDate.getTime() - startDate.getTime()) / 1000);
    if (diff < 60) return `${diff}s`;
    if (diff < 3600) return `${Math.floor(diff / 60)}m`;
    return `${Math.floor(diff / 3600)}h ${Math.floor((diff % 3600) / 60)}m`;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 text-red-700 rounded-lg">
        Error: {error}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">Tasks</h1>
        <span className="text-sm text-muted-foreground">
          {tasks.filter(t => t.status === 'active').length} active, {tasks.filter(t => t.status === 'closed').length} closed
        </span>
      </div>

      <div className="grid gap-4">
        {tasks.length === 0 ? (
          <div className="text-center py-12 text-muted-foreground">
            No tasks found. Create one with <code className="bg-muted px-2 py-1 rounded">bar task start &lt;name&gt;</code>
          </div>
        ) : (
          tasks.map((task) => (
            <Link
              key={task.id}
              to={`/tasks/${task.id}`}
              className="block p-4 bg-card rounded-lg border border-border hover:border-primary/50 transition-colors"
            >
              <div className="flex items-start justify-between">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <h3 className="font-semibold">{task.name}</h3>
                    {task.is_active && (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs bg-green-100 text-green-700">
                        <Activity className="w-3 h-3" />
                        Active
                      </span>
                    )}
                    {task.status === 'closed' && (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs bg-gray-100 text-gray-700">
                        <CheckCircle className="w-3 h-3" />
                        Closed
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-4 text-sm text-muted-foreground">
                    <span className="flex items-center gap-1">
                      <FolderGit className="w-4 h-4" />
                      {task.branch}
                    </span>
                    <span className="flex items-center gap-1">
                      <Clock className="w-4 h-4" />
                      {formatDate(task.created_at)}
                    </span>
                  </div>
                </div>
                <div className="text-right text-sm text-muted-foreground">
                  <div>ID: {task.id}</div>
                  {task.closed_at && (
                    <div>Duration: {formatDuration(task.created_at, task.closed_at)}</div>
                  )}
                </div>
              </div>
            </Link>
          ))
        )}
      </div>
    </div>
  );
}
