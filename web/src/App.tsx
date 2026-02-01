import { Routes, Route, Link, useLocation } from 'react-router-dom';
import { LayoutDashboard, Activity } from 'lucide-react';
import { TasksPage } from './pages/TasksPage';
import { TaskDetailPage } from './pages/TaskDetailPage';
import { useWebSocket } from './hooks/useWebSocket';

function Layout({ children }: { children: React.ReactNode }) {
  const location = useLocation();
  const { connected } = useWebSocket();

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border">
        <div className="max-w-6xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold">BAR</h1>
              <span className="text-sm text-muted-foreground">Blade Agent Runtime</span>
            </div>
            <div className="flex items-center gap-4">
              <div className={`flex items-center gap-1 text-sm ${connected ? 'text-green-600' : 'text-red-600'}`}>
                <Activity className="w-4 h-4" />
                {connected ? 'Connected' : 'Disconnected'}
              </div>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-6xl mx-auto px-4 py-6">
        <div className="flex gap-6">
          <aside className="w-48 shrink-0">
            <nav className="space-y-1">
              <Link
                to="/"
                className={`flex items-center gap-2 px-3 py-2 rounded-lg transition-colors ${
                  location.pathname === '/' ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'
                }`}
              >
                <LayoutDashboard className="w-4 h-4" />
                Tasks
              </Link>
            </nav>
          </aside>

          <main className="flex-1 min-w-0">
            {children}
          </main>
        </div>
      </div>
    </div>
  );
}

function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<TasksPage />} />
        <Route path="/tasks/:id" element={<TaskDetailPage />} />
      </Routes>
    </Layout>
  );
}

export default App;
