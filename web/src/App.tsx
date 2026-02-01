import { useEffect, useState } from 'react';
import { Routes, Route, NavLink, useLocation, useNavigate } from 'react-router-dom';
import { List, Terminal, PanelLeftClose, PanelLeft } from 'lucide-react';
import TasksPage from './pages/TasksPage';
import TaskDetailPage from './pages/TaskDetailPage';
import { useWebSocket } from './hooks/useWebSocket';
import { api } from '@/services/api';

const NavItem = ({ to, icon: Icon, label, collapsed }: { to: string; icon: React.ElementType; label: string; collapsed?: boolean }) => {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        `flex items-center gap-3 px-3 py-2 rounded-md transition-all duration-200 group ${
          isActive 
            ? 'bg-zinc-800 text-white shadow-sm ring-1 ring-zinc-700' 
            : 'text-zinc-400 hover:text-white hover:bg-zinc-800/50'
        } ${collapsed ? 'justify-center' : ''}`
      }
      title={collapsed ? label : undefined}
    >
      <Icon className="w-4 h-4 flex-shrink-0" />
      {!collapsed && <span className="text-sm font-medium">{label}</span>}
    </NavLink>
  );
};

const Layout = ({ children }: { children: React.ReactNode }) => {
  const location = useLocation();
  const { connected } = useWebSocket();
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  
  const getPageTitle = () => {
    if (location.pathname.startsWith('/tasks/')) return 'Task Details';
    if (location.pathname === '/tasks' || location.pathname === '/') return 'Tasks';
    return 'Dashboard';
  };

  const isTaskDetail = location.pathname.startsWith('/tasks/');

  return (
    <div className="flex overflow-hidden w-full h-screen bg-background text-primary">
      {/* Sidebar */}
      <aside 
        className={`flex flex-col border-r border-border bg-zinc-950/50 transition-all duration-200 ${
          sidebarCollapsed ? 'w-16' : 'w-64'
        }`}
      >
        <div className={`flex items-center p-4 border-b border-border/50 ${sidebarCollapsed ? 'justify-center' : 'gap-2 px-6'}`}>
          {sidebarCollapsed ? (
            <button
              onClick={() => setSidebarCollapsed(false)}
              className="p-1.5 rounded hover:bg-zinc-800 text-zinc-500 hover:text-white transition-colors"
              title="Expand sidebar"
            >
              <PanelLeft className="w-5 h-5" />
            </button>
          ) : (
            <>
              <div className="flex justify-center items-center w-8 h-8 bg-white rounded-lg flex-shrink-0">
                <Terminal className="w-5 h-5 text-black" />
              </div>
              <span className="text-lg font-bold tracking-tight text-white flex-1">BAR <span className="font-normal text-zinc-500">UI</span></span>
              <button
                onClick={() => setSidebarCollapsed(true)}
                className="p-1.5 rounded hover:bg-zinc-800 text-zinc-500 hover:text-white transition-colors"
                title="Collapse sidebar"
              >
                <PanelLeftClose className="w-4 h-4" />
              </button>
            </>
          )}
        </div>

        <nav className={`flex-1 py-6 space-y-1 ${sidebarCollapsed ? 'px-2' : 'px-4'}`}>
          <NavItem to="/tasks" icon={List} label="Tasks" collapsed={sidebarCollapsed} />
        </nav>

        <div className={`p-4 border-t border-border/50 ${sidebarCollapsed ? 'px-2' : ''}`}>
          <div className={`flex items-center py-2 ${sidebarCollapsed ? 'justify-center' : 'gap-3 px-2'}`}>
            <div className={`w-2 h-2 rounded-full flex-shrink-0 ${connected ? 'bg-emerald-500 animate-pulse' : 'bg-red-500'}`}></div>
            {!sidebarCollapsed && (
              <span className="font-mono text-xs text-zinc-500">
                {connected ? 'SYSTEM ONLINE' : 'DISCONNECTED'}
              </span>
            )}
          </div>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex overflow-hidden flex-col flex-1 min-w-0">
        {/* Topbar */}
        <header className="flex z-10 justify-between items-center px-8 h-16 border-b backdrop-blur border-border bg-background/80">
          <h1 className="text-xl font-semibold text-zinc-100">{getPageTitle()}</h1>
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
