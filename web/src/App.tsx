import { useEffect } from 'react';
import { Routes, Route, NavLink, useLocation, useNavigate } from 'react-router-dom';
import { List, Terminal } from 'lucide-react';
import TasksPage from './pages/TasksPage';
import TaskDetailPage from './pages/TaskDetailPage';
import { useWebSocket } from './hooks/useWebSocket';
import { api } from '@/services/api';

// Sidebar Nav Item
const NavItem = ({ to, icon: Icon, label }: { to: string; icon: React.ElementType; label: string }) => {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        `flex items-center gap-3 px-3 py-2 rounded-md transition-all duration-200 group ${
          isActive 
            ? 'bg-zinc-800 text-white shadow-sm ring-1 ring-zinc-700' 
            : 'text-zinc-400 hover:text-white hover:bg-zinc-800/50'
        }`
      }
    >
      <Icon className="w-4 h-4" />
      <span className="text-sm font-medium">{label}</span>
    </NavLink>
  );
};

const Layout = ({ children }: { children: React.ReactNode }) => {
  const location = useLocation();
  const { connected } = useWebSocket();
  
  const getPageTitle = () => {
    if (location.pathname.startsWith('/tasks/')) return 'Task Details';
    if (location.pathname === '/tasks' || location.pathname === '/') return 'Tasks';
    return 'Dashboard';
  };

  const isTaskDetail = location.pathname.startsWith('/tasks/');

  return (
    <div className="flex overflow-hidden w-full h-screen bg-background text-primary">
      {/* Sidebar */}
      <aside className="flex flex-col w-64 border-r border-border bg-zinc-950/50">
        <div className="flex gap-2 items-center p-6 border-b border-border/50">
          <div className="flex justify-center items-center w-8 h-8 bg-white rounded-lg">
            <Terminal className="w-5 h-5 text-black" />
          </div>
          <span className="text-lg font-bold tracking-tight text-white">BAR <span className="font-normal text-zinc-500">UI</span></span>
        </div>

        <nav className="flex-1 px-4 py-6 space-y-1">
          <NavItem to="/tasks" icon={List} label="Tasks" />
        </nav>

        <div className="p-4 border-t border-border/50">
          <div className="flex gap-3 items-center px-2 py-2">
            <div className={`w-2 h-2 rounded-full ${connected ? 'bg-emerald-500 animate-pulse' : 'bg-red-500'}`}></div>
            <span className="font-mono text-xs text-zinc-500">
              {connected ? 'SYSTEM ONLINE' : 'DISCONNECTED'}
            </span>
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex overflow-hidden flex-col flex-1 min-w-0">
        {/* Topbar */}
        <header className="flex z-10 justify-between items-center px-8 h-16 border-b backdrop-blur border-border bg-background/80">
          <h1 className="text-xl font-semibold text-zinc-100">{getPageTitle()}</h1>
          <div className="flex gap-4 items-center">
             <div className="w-8 h-8 bg-gradient-to-tr rounded-full ring-1 from-zinc-700 to-zinc-600 ring-white/10"></div>
          </div>
        </header>

        {/* Content Area */}
        <div className={`flex-1 ${isTaskDetail ? 'flex overflow-hidden flex-col' : 'overflow-y-auto p-8 scroll-smooth'}`}>
          <div className={`${isTaskDetail ? 'flex-1 h-full' : 'mx-auto max-w-6xl animate-fade-in'}`}>
            {children}
          </div>
        </div>
      </main>
    </div>
  );
};

const HomeRedirect = () => {
  const navigate = useNavigate();
  useEffect(() => {
    api.getStatus().then(status => {
      if (status.active_task_id) {
        navigate(`/tasks/${status.active_task_id}`, { replace: true });
      } else {
        navigate('/tasks', { replace: true });
      }
    }).catch(() => {
        navigate('/tasks', { replace: true });
    });
  }, []);
  return (
    <div className="flex justify-center items-center h-full">
      <div className="w-8 h-8 rounded-full border-b-2 animate-spin border-primary"></div>
    </div>
  );
};

const App = () => {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<HomeRedirect />} />
        <Route path="/tasks" element={<TasksPage />} />
        <Route path="/tasks/:id" element={<TaskDetailPage />} />
      </Routes>
    </Layout>
  );
};

export default App;
