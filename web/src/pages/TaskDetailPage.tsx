import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { 
  Terminal, RotateCcw, FileDiff, 
  GitBranch, PanelLeftClose, PanelLeft, Radio
} from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { api } from '@/services/api';
import { useWebSocket } from '@/hooks/useWebSocket';
import type { Task, LedgerStep, LiveDiffData } from '@/types';
import DiffViewer from '@/components/DiffViewer';

const StepIcon = ({ kind }: { kind: string }) => {
  if (kind === 'rollback') return <RotateCcw className="w-4 h-4 text-rose-400" />;
  if (kind === 'apply') return <FileDiff className="w-4 h-4 text-purple-400" />;
  return <Terminal className="w-4 h-4 text-blue-400" />;
};

const TaskDetailPage = () => {
  const { id } = useParams<{ id: string }>();
  const { lastMessage } = useWebSocket();
  const [task, setTask] = useState<Task | null>(null);
  const [ledger, setLedger] = useState<LedgerStep[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  const [selectedStepId, setSelectedStepId] = useState<string | null>(null);
  const [diffContent, setDiffContent] = useState('');
  const [diffLoading, setDiffLoading] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  
  // Live diff state
  const [liveDiff, setLiveDiff] = useState<{
    files: number;
    additions: number;
    deletions: number;
    patch: string;
  } | null>(null);
  const [showLive, setShowLive] = useState(true);

  useEffect(() => {
    if (id) {
      loadTaskData();
    }
  }, [id]);

  // Listen for live diff updates
  useEffect(() => {
    if (lastMessage?.type === 'live_diff') {
      const data = lastMessage.data as LiveDiffData;
      if (data.task_id === id) {
        setLiveDiff({
          files: data.files,
          additions: data.additions,
          deletions: data.deletions,
          patch: data.patch,
        });
        // Auto-switch to live view
        if (showLive) {
          setSelectedStepId(null);
          setDiffContent(data.patch);
        }
      }
    }
  }, [lastMessage, id, showLive]);

  useEffect(() => {
    if (ledger.length > 0 && !selectedStepId) {
      const lastStepWithPatch = [...ledger].reverse().find(s => s.artifacts?.patch);
      if (lastStepWithPatch) {
        handleSelectStep(lastStepWithPatch);
      }
    }
  }, [ledger]);

  const loadTaskData = async () => {
    try {
      setLoading(true);
      const [taskData, ledgerData] = await Promise.all([
        api.getTask(id!),
        api.getLedger(id!).catch(() => []),
      ]);
      setTask(taskData);
      setLedger(ledgerData || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load task');
    } finally {
      setLoading(false);
    }
  };

  const handleSelectStep = async (step: LedgerStep) => {
    if (!step.artifacts?.patch) return;
    
    setSelectedStepId(step.step_id);
    setDiffLoading(true);
    try {
      if (!id) return;
      const content = await api.getDiff(id, step.step_id);
      setDiffContent(content);
    } catch {
      setDiffContent('Failed to load diff content');
    } finally {
      setDiffLoading(false);
    }
  };

  const getStatus = (step: LedgerStep) => {
    if (step.exit_code === 0) return 'success';
    if (step.exit_code !== undefined && step.exit_code !== 0) return 'failure';
    return 'pending';
  };

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (error || !task) {
    return (
      <div className="p-4 bg-red-950/30 border border-red-900/50 text-red-400 rounded-lg m-8">
        Error: {error || 'Task not found'}
      </div>
    );
  }

  return (
    <div className="flex w-full h-full overflow-hidden">
      {/* Collapsible Sidebar */}
      <AnimatePresence initial={false}>
        {!sidebarCollapsed && (
          <motion.div 
            initial={{ width: 0, opacity: 0 }}
            animate={{ width: 350, opacity: 1 }}
            exit={{ width: 0, opacity: 0 }}
            transition={{ duration: 0.2 }}
            className="flex flex-col border-r border-border bg-zinc-950/30 overflow-hidden"
          >
            {/* Task Header */}
            <div className="flex-shrink-0 p-4 border-b border-border bg-background/50 backdrop-blur z-10">
              <div className="flex items-center justify-between mb-3">
                <h1 className="text-lg font-bold text-white tracking-tight truncate" title={task.name}>{task.name}</h1>
                <button
                  onClick={() => setSidebarCollapsed(true)}
                  className="p-1.5 rounded hover:bg-zinc-800 text-zinc-500 hover:text-white transition-colors"
                  title="Collapse sidebar"
                >
                  <PanelLeftClose className="w-4 h-4" />
                </button>
              </div>
              <div className="flex flex-col gap-2 text-sm text-zinc-400">
                <div className="flex items-center gap-2">
                  <span className="bg-zinc-800 px-2 py-0.5 rounded text-zinc-300 text-xs font-mono">#{task.id}</span>
                  <span className={`px-2 py-0.5 rounded-full text-xs font-medium border ${
                    task.status === 'active' 
                      ? 'bg-emerald-950/30 text-emerald-400 border-emerald-900/50' 
                      : 'bg-zinc-800/50 text-zinc-400 border-zinc-700/50'
                  }`}>
                    {task.status.toUpperCase()}
                  </span>
                </div>
                <div className="flex items-center gap-2 text-xs">
                  <GitBranch className="w-3.5 h-3.5" />
                  <span className="font-mono truncate">{task.branch}</span>
                </div>
              </div>
            </div>

            {/* Timeline */}
            <div className="flex-1 min-h-0 overflow-y-auto p-4 scroll-smooth">
              <div className="relative pl-4 border-l border-zinc-800 space-y-4">
                {ledger.length === 0 ? (
                  <div className="text-zinc-500 py-4 text-sm">No steps recorded yet.</div>
                ) : (
                  ledger.map((step, index) => {
                    const status = getStatus(step);
                    const hasPatch = !!step.artifacts?.patch;
                    const isSelected = selectedStepId === step.step_id;
                    
                    return (
                      <motion.div 
                        key={step.step_id}
                        initial={{ opacity: 0, x: -10 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: index * 0.05 }}
                        className="relative group"
                      >
                        <div className={`absolute -left-[21px] top-3 w-2.5 h-2.5 rounded-full border-2 border-background ring-4 ring-background
                          ${status === 'success' ? 'bg-emerald-500' : status === 'failure' ? 'bg-rose-500' : 'bg-zinc-500'}
                          ${isSelected ? 'scale-125' : ''} transition-all duration-200
                        `} />

                        <div 
                          onClick={() => hasPatch && handleSelectStep(step)}
                          className={`
                            relative rounded-lg border transition-all duration-200 overflow-hidden
                            ${hasPatch ? 'cursor-pointer hover:border-zinc-600' : 'opacity-80'}
                            ${isSelected 
                              ? 'bg-zinc-900 border-accent/50 shadow-[0_0_15px_-3px_rgba(37,99,235,0.2)]' 
                              : 'bg-surface border-border'
                            }
                          `}
                        >
                          {isSelected && (
                            <div className="absolute left-0 top-0 bottom-0 w-1 bg-accent" />
                          )}

                          <div className="p-3">
                            <div className="flex items-center justify-between mb-2">
                              <div className="flex items-center gap-2">
                                <StepIcon kind={step.kind} />
                                <span className={`text-sm font-medium ${isSelected ? 'text-white' : 'text-zinc-300'}`}>
                                  {step.kind}
                                </span>
                              </div>
                              <span className="text-xs font-mono text-zinc-500">{formatDate(step.started_at)}</span>
                            </div>
                            
                            {step.cmd && step.cmd.length > 0 && (
                              <div className="font-mono text-xs text-zinc-400 mb-2 truncate bg-black/30 p-1.5 rounded">
                                $ {step.cmd.join(' ')}
                              </div>
                            )}

                            {step.diff_stat && (
                              <div className="flex items-center gap-3 text-xs">
                                 <span className="text-zinc-500">{step.diff_stat.files} files</span>
                                 <div className="flex items-center gap-1">
                                   <span className="text-emerald-500">+{step.diff_stat.additions}</span>
                                   <span className="text-rose-500">-{step.diff_stat.deletions}</span>
                                 </div>
                              </div>
                            )}
                          </div>
                        </div>
                      </motion.div>
                    );
                  })
                )}
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Main Content */}
      <div className="flex-1 bg-zinc-950 flex flex-col min-w-0">
        {/* Toolbar */}
        <div className="h-12 flex items-center px-4 border-b border-zinc-800 bg-zinc-900/50 text-sm text-zinc-300 select-none gap-3">
          {sidebarCollapsed && (
            <button
              onClick={() => setSidebarCollapsed(false)}
              className="p-1.5 rounded hover:bg-zinc-800 text-zinc-500 hover:text-white transition-colors"
              title="Expand sidebar"
            >
              <PanelLeft className="w-4 h-4" />
            </button>
          )}
          <FileDiff className="w-4 h-4 text-zinc-500" />
          <span>
            {!selectedStepId && liveDiff ? 'Live Changes' : selectedStepId ? `Step ${selectedStepId}` : 'No step selected'}
          </span>
          {diffLoading && <span className="text-xs text-zinc-500">(Loading...)</span>}
          
          {/* Live indicator */}
          {liveDiff && (
            <div className="ml-auto flex items-center gap-3">
              <div className="flex items-center gap-2 text-xs">
                <span className="text-zinc-500">{liveDiff.files} files</span>
                <span className="text-emerald-500">+{liveDiff.additions}</span>
                <span className="text-rose-500">-{liveDiff.deletions}</span>
              </div>
              <button
                onClick={() => {
                  setShowLive(!showLive);
                  if (!showLive && liveDiff) {
                    setSelectedStepId(null);
                    setDiffContent(liveDiff.patch);
                  }
                }}
                className={`flex items-center gap-1.5 px-2 py-1 rounded text-xs transition-colors ${
                  showLive 
                    ? 'bg-emerald-950/50 text-emerald-400 border border-emerald-900/50' 
                    : 'bg-zinc-800 text-zinc-400 border border-zinc-700'
                }`}
              >
                <Radio className={`w-3 h-3 ${showLive ? 'animate-pulse' : ''}`} />
                LIVE
              </button>
            </div>
          )}
        </div>

        {/* Diff Content */}
        <div className="flex-1 relative overflow-hidden">
          {(selectedStepId || (showLive && liveDiff)) ? (
            diffLoading ? (
              <div className="absolute inset-0 flex items-center justify-center bg-zinc-900/50 z-10">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
              </div>
            ) : (
              <DiffViewer diffContent={diffContent} />
            )
          ) : (
            <div className="flex-1 h-full flex flex-col items-center justify-center text-zinc-500 gap-4">
              <div className="w-16 h-16 rounded-2xl bg-zinc-900 flex items-center justify-center">
                <FileDiff className="w-8 h-8 opacity-50" />
              </div>
              <p>{task?.status === 'active' ? 'Waiting for changes...' : 'Select a step to view diff'}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default TaskDetailPage;
