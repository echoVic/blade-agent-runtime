import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { 
  ArrowLeft, Terminal, RotateCcw, FileDiff, 
  GitBranch
} from 'lucide-react';
import { motion } from 'framer-motion';
import Editor from '@monaco-editor/react';
import { api } from '@/services/api';
import type { Task, LedgerStep } from '@/types';

const StepIcon = ({ kind }: { kind: string }) => {
  if (kind === 'rollback') return <RotateCcw className="w-4 h-4 text-rose-400" />;
  if (kind === 'apply') return <FileDiff className="w-4 h-4 text-purple-400" />;
  return <Terminal className="w-4 h-4 text-blue-400" />;
};

const TaskDetailPage = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [task, setTask] = useState<Task | null>(null);
  const [ledger, setLedger] = useState<LedgerStep[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  const [selectedStepId, setSelectedStepId] = useState<string | null>(null);
  const [diffContent, setDiffContent] = useState('');
  const [diffLoading, setDiffLoading] = useState(false);

  useEffect(() => {
    if (id) {
      loadTaskData();
    }
  }, [id]);

  // Auto-select the last step with patch when ledger is loaded
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

  const handleSelectStep = async (step: LedgerStep) => {
    if (!step.artifacts?.patch) return;
    
    setSelectedStepId(step.step_id);
    setDiffLoading(true);
    try {
      if (!id) return;
      const content = await api.getDiff(id, step.step_id);
      setDiffContent(content);
    } catch (err) {
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
      {/* Left Sidebar: Timeline */}
      <div className="flex flex-col w-1/3 min-w-[350px] max-w-[450px] border-r border-border bg-zinc-950/30">
        {/* Task Header */}
        <div className="flex-shrink-0 p-6 border-b border-border bg-background/50 backdrop-blur z-10">
          <button 
            onClick={() => navigate('/')} 
            className="flex items-center gap-2 mb-4 text-zinc-500 hover:text-white transition-colors text-sm group w-fit"
          >
            <ArrowLeft className="w-4 h-4 group-hover:-translate-x-1 transition-transform" />
            Back to Tasks
          </button>
          
          <h1 className="text-2xl font-bold text-white tracking-tight truncate" title={task.name}>{task.name}</h1>
          <div className="flex flex-col gap-2 mt-3 text-sm text-zinc-400">
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
              <span className="font-mono">{task.branch}</span>
            </div>
          </div>
        </div>

        {/* Timeline List */}
        <div className="flex-1 overflow-y-auto p-6 scroll-smooth">
          <div className="relative pl-4 border-l border-zinc-800 space-y-6">
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
                    {/* Timeline Dot */}
                    <div className={`absolute -left-[21px] top-4 w-2.5 h-2.5 rounded-full border-2 border-background ring-4 ring-background
                      ${status === 'success' ? 'bg-emerald-500' : status === 'failure' ? 'bg-rose-500' : 'bg-zinc-500'}
                      ${isSelected ? 'scale-125' : ''} transition-all duration-200
                    `} />

                    {/* Card */}
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
                      {/* Active Indicator Bar */}
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
      </div>

      {/* Right Main: Diff Editor */}
      <div className="flex-1 bg-[#1e1e1e] flex flex-col min-w-0">
        {selectedStepId ? (
          <>
            <div className="h-12 flex items-center px-4 border-b border-[#333] bg-[#252526] text-sm text-zinc-300 select-none">
              <FileDiff className="w-4 h-4 mr-2 text-zinc-500" />
              <span>Step {selectedStepId} Patch</span>
              {diffLoading && <span className="ml-2 text-xs text-zinc-500">(Loading...)</span>}
            </div>
            <div className="flex-1 relative">
              {diffLoading ? (
                <div className="absolute inset-0 flex items-center justify-center bg-zinc-900/50 z-10">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                </div>
              ) : null}
              <Editor
                height="100%"
                defaultLanguage="diff"
                theme="vs-dark"
                value={diffContent}
                options={{
                  readOnly: true,
                  minimap: { enabled: true },
                  scrollBeyondLastLine: false,
                  fontSize: 13,
                  fontFamily: 'JetBrains Mono, Menlo, monospace',
                  renderLineHighlight: 'all',
                  padding: { top: 16, bottom: 16 },
                }}
              />
            </div>
          </>
        ) : (
          <div className="flex-1 flex flex-col items-center justify-center text-zinc-500 gap-4">
            <div className="w-16 h-16 rounded-2xl bg-zinc-900 flex items-center justify-center">
              <FileDiff className="w-8 h-8 opacity-50" />
            </div>
            <p>Select a step with changes to view diff</p>
          </div>
        )}
      </div>
    </div>
  );
};

export default TaskDetailPage;
