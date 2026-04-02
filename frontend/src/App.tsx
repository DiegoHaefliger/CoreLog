import React, { useEffect } from 'react';
import { useLogStore } from './store/logStore';
import { LogTable } from './components/LogTable';
import { Search, Shield, Pause, Play } from 'lucide-react';
import clsx from 'clsx';
import axios from 'axios';

// Basic Login Overlay for MVP
const LoginOverlay = ({ onLogin }: { onLogin: () => void }) => {
  const [username, setUsername] = React.useState('admin');
  const [password, setPassword] = React.useState('admin');
  const [mustChangePassword, setMustChangePassword] = React.useState(false);
  const [newPassword, setNewPassword] = React.useState('');
  const [confirmPassword, setConfirmPassword] = React.useState('');

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await axios.post('/api/login', { username, password }, { withCredentials: true });
      if (res.data.requires_password_change) {
        setMustChangePassword(true);
      } else {
        onLogin(); // Success
      }
    } catch (err) {
      alert("Login falhou");
    }
  };

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (newPassword !== confirmPassword) {
      alert("Senhas não batem");
      return;
    }
    if (newPassword.length < 5) {
      alert("A nova senha deve ter pelo menos 5 caracteres");
      return;
    }
    try {
      await axios.post('/api/change-password', { new_password: newPassword }, { withCredentials: true });
      onLogin(); // Proceed to dashboard!
    } catch (err) {
      alert("Erro ao alterar senha");
    }
  };

  if (mustChangePassword) {
    return (
      <div className="fixed inset-0 bg-[#0d1117] flex items-center justify-center z-50">
        <form onSubmit={handleChangePassword} className="bg-[#161b22] p-8 rounded-lg border border-[#30363d] w-96 font-sans">
          <div className="flex items-center gap-3 mb-6 justify-center">
            <Shield className="w-8 h-8 text-[#d29922]" />
            <h1 className="text-2xl font-bold text-white tracking-tight">Segurança</h1>
          </div>
          <p className="text-[#8b949e] mb-6 text-center text-sm">O administrator requer que você mude a senha padrão para usar o painel.</p>
          <input
            type="password"
            value={newPassword} onChange={e => setNewPassword(e.target.value)}
            className="w-full bg-[#0d1117] border border-[#30363d] rounded px-3 py-2 text-white mb-4 outline-none focus:border-[#d29922] transition-colors"
            placeholder="Nova Senha"
          />
          <input
            type="password"
            value={confirmPassword} onChange={e => setConfirmPassword(e.target.value)}
            className="w-full bg-[#0d1117] border border-[#30363d] rounded px-3 py-2 text-white mb-6 outline-none focus:border-[#d29922] transition-colors"
            placeholder="Confirme Nova Senha"
          />
          <button type="submit" className="w-full bg-[#d29922] hover:bg-[#b07d1a] text-black font-semibold py-2 px-4 rounded transition-colors">
            Salvar e Continuar
          </button>
        </form>
      </div>
    );
  }

  return (
    <div className="fixed inset-0 bg-[#0d1117] flex items-center justify-center z-50">
      <form onSubmit={handleLogin} className="bg-[#161b22] p-8 rounded-lg border border-[#30363d] w-96 font-sans">
        <div className="flex items-center gap-3 mb-6 justify-center">
          <Shield className="w-8 h-8 text-[var(--color-core-accent)]" />
          <h1 className="text-2xl font-bold text-white tracking-tight">CoreLog</h1>
        </div>
        <p className="text-[#8b949e] mb-6 text-center text-sm">Autenticação necessária para acesso aos logs sensitivos.</p>
        <input
          type="text"
          value={username} onChange={e => setUsername(e.target.value)}
          className="w-full bg-[#0d1117] border border-[#30363d] rounded px-3 py-2 text-white mb-4 outline-none focus:border-[var(--color-core-accent)] transition-colors"
          placeholder="Usuário"
        />
        <input
          type="password"
          value={password} onChange={e => setPassword(e.target.value)}
          className="w-full bg-[#0d1117] border border-[#30363d] rounded px-3 py-2 text-white mb-6 outline-none focus:border-[var(--color-core-accent)] transition-colors"
          placeholder="Senha"
        />
        <button type="submit" className="w-full bg-[var(--color-core-success)] hover:bg-[#2ea043] text-white font-medium py-2 px-4 rounded transition-colors">
          Acessar Painel
        </button>
      </form>
    </div>
  );
};

export default function App() {
  const [isAuthenticated, setIsAuthenticated] = React.useState(false);
  const { availableServices, setAvailableServices, isLive, toggleLive, addLiveLog, setLogs, filterLevel, setFilterLevel, searchQueries, addSearchQuery, removeSearchQuery, selectedService, setSelectedService, dateFrom, setDateFrom, dateTo, setDateTo } = useLogStore();

  const fetchServices = async () => {
    try {
      const res = await axios.get('/api/services', { withCredentials: true });
      setAvailableServices(res.data || []);
    } catch (e) {
      console.error("Error fetching services", e);
    }
  };

  // Load initial logs
  // Converte string ISO (ex: "2026-04-02T07:31:58Z") para "YYYY-MM-DD HH:MM:SS" UTC para comparação no backend
  const toUTC = (isoStr: string): string => {
    const d = new Date(isoStr);
    if (isNaN(d.getTime())) return '';
    return d.toISOString().slice(0, 19).replace('T', ' ');
  };

  const fetchLogs = async () => {
    try {
      const formattedFrom = dateFrom ? toUTC(dateFrom) : '';
      const formattedTo = dateTo ? toUTC(dateTo) : '';

      let finalQ = "";
      if (searchQueries.length > 0) {
        finalQ = searchQueries.map(s => `"${s}"`).join(' OR ');
      }

      const res = await axios.get(`/api/logs?q=${encodeURIComponent(finalQ)}&service=${encodeURIComponent(selectedService || '')}&level=${encodeURIComponent(filterLevel || '')}&from=${encodeURIComponent(formattedFrom)}&to=${encodeURIComponent(formattedTo)}`, { withCredentials: true });
      setLogs(res.data || []);
    } catch (e) {
      console.error(e);
      if (axios.isAxiosError(e) && e.response?.status === 401) {
        setIsAuthenticated(false);
      }
    }
  };

  useEffect(() => {
    if (!isAuthenticated) return;
    fetchServices();
  }, [isAuthenticated]);

  useEffect(() => {
    if (!isAuthenticated) return;
    fetchLogs();
  }, [isAuthenticated, searchQueries, selectedService, filterLevel, dateFrom, dateTo]);

  // WebSocket Connection
  useEffect(() => {
    if (!isAuthenticated || !isLive) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsHost = window.location.host || 'localhost:8080';
    const ws = new WebSocket(`${protocol}//${wsHost}/api/ws`);

    ws.onmessage = (event) => {
      try {
        const log = JSON.parse(event.data);
        addLiveLog(log);
      } catch (e) {
        console.error("WS Parse error", e);
      }
    };

    return () => ws.close();
  }, [isAuthenticated, isLive]);

  if (!isAuthenticated) {
    return <LoginOverlay onLogin={() => setIsAuthenticated(true)} />;
  }

  const levels = ['ERROR', 'WARN', 'INFO', 'DEBUG'];

  return (
    <div className="flex h-screen bg-[var(--color-core-bg)] text-[var(--color-core-text)] font-sans antialiased overflow-hidden">

      {/* Sidebar minimalista expandida para mostrar aplicações */}
      <div className="w-48 flex flex-col items-start py-4 border-r border-[var(--color-core-border)] z-10 shrink-0 bg-[#010409]">
        <div className="w-full flex items-center gap-2 px-4 mb-8">
          <Shield className="w-6 h-6 text-[var(--color-core-accent)] shrink-0" />
          <div className="flex flex-col">
            <span className="font-bold text-white tracking-tight">CoreLog</span>
            <span className="text-[10px] text-[#484f58]">v1.0.1</span>
          </div>
        </div>

        <div className="flex flex-col w-full text-sm">
          <div className="uppercase text-xs font-semibold text-[#8b949e] px-4 mb-2 tracking-wider">Aplicações</div>

          <button
            onClick={() => setSelectedService(null)}
            className={clsx(
              "w-full text-left px-4 py-2 hover:bg-[#161b22] hover:text-[var(--color-core-text)] transition-colors border-l-2",
              selectedService === null ? "text-[var(--color-core-text)] border-[var(--color-core-accent)] bg-[#161b22]" : "text-[#8b949e] border-transparent"
            )}
          >
            Todas
          </button>

          {availableServices.map(service => (
            <button
              key={service}
              onClick={() => setSelectedService(service)}
              className={clsx(
                "w-full text-left px-4 py-2 hover:bg-[#161b22] hover:text-[var(--color-core-text)] transition-colors border-l-2 truncate",
                selectedService === service ? "text-[var(--color-core-accent)] border-[var(--color-core-accent)] bg-[#161b22]" : "text-[#8b949e] border-transparent"
              )}
            >
              {service}
            </button>
          ))}
        </div>
      </div>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0">

        {/* Top Navbar Header */}
        <header className="h-16 border-b border-[var(--color-core-border)] bg-[var(--color-core-bg)] flex items-center px-6 gap-6 shrink-0 z-10 shadow-sm">
          <div className="text-lg font-semibold tracking-tight text-white mr-4">Logs em Tempo Real</div>

          {/* Main Search Bar (Tag Based) */}
          <div className="flex-1 max-w-2xl flex items-center bg-[#010409] border border-[var(--color-core-border)] rounded-md px-3 py-1.5 focus-within:border-[var(--color-core-accent)] focus-within:ring-1 focus-within:ring-[var(--color-core-accent)] transition-all overflow-x-hidden">
            <Search className="w-4 h-4 text-[#8b949e] mr-2 shrink-0" />

            <div className="flex flex-1 items-center gap-2 overflow-x-auto no-scrollbar">
              {searchQueries.map(q => (
                <div key={q} className="flex items-center gap-1 bg-[#1f2428] text-white px-2 rounded-full text-xs whitespace-nowrap">
                  {q}
                  <button onClick={() => removeSearchQuery(q)} className="text-[#8b949e] hover:text-white flex-shrink-0 ml-1 font-bold">×</button>
                </div>
              ))}
              <input
                type="text"
                placeholder={searchQueries.length === 0 ? "Buscar via SQLite FTS5 (Pressione Enter para adicionar filtro)..." : "+ Adicionar"}
                className="bg-transparent border-none outline-none text-sm min-w-[200px] flex-1 text-white placeholder-[#484f58]"
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    const val = e.currentTarget.value.trim();
                    if (val) {
                      addSearchQuery(val);
                      e.currentTarget.value = '';
                    }
                  } else if (e.key === 'Backspace' && e.currentTarget.value === '' && searchQueries.length > 0) {
                    e.preventDefault();
                    removeSearchQuery(searchQueries[searchQueries.length - 1]);
                  }
                }}
              />
            </div>
          </div>

          {/* Level Filter Toggles */}
          <div className="flex items-center gap-2 bg-[#010409] border border-[var(--color-core-border)] rounded-md p-1 shrink-0">
            {levels.map(level => (
              <button
                key={level}
                onClick={() => setFilterLevel(filterLevel === level ? null : level)}
                className={clsx(
                  "px-3 py-1 text-xs font-semibold rounded cursor-pointer transition-colors",
                  filterLevel === level
                    ? "bg-[#1f2428] text-white shadow-sm"
                    : "text-[#8b949e] hover:bg-[#161b22]"
                )}
              >
                {level}
              </button>
            ))}
          </div>

          <div className="flex items-center gap-2 text-sm shrink-0">
            <input
              type="text"
              value={dateFrom}
              onChange={e => setDateFrom(e.target.value)}
              placeholder="2026-04-02T07:31:58Z"
              className="bg-[#010409] text-[#8b949e] border border-[var(--color-core-border)] rounded-md px-2 py-1 outline-none focus:border-[#58a6ff] w-44 font-mono text-xs"
              title="Data Inicial (ex: 2026-04-02T07:31:58Z)"
            />
            <span className="text-[#8b949e]">-</span>
            <input
              type="text"
              value={dateTo}
              onChange={e => setDateTo(e.target.value)}
              placeholder="2026-04-02T07:31:58Z"
              className="bg-[#010409] text-[#8b949e] border border-[var(--color-core-border)] rounded-md px-2 py-1 outline-none focus:border-[#58a6ff] w-44 font-mono text-xs"
              title="Data Final (ex: 2026-04-02T07:31:58Z)"
            />
          </div>

          {/* Live Tail Toggle */}
          <button
            onClick={toggleLive}
            className={clsx(
              "ml-auto flex items-center gap-2 px-4 py-1.5 rounded-md text-sm font-medium transition-all border",
              isLive
                ? "bg-[var(--color-core-success)] hover:bg-[#2ea043] border-transparent text-white shadow-[0_0_10px_rgba(63,185,80,0.3)]"
                : "bg-transparent border-[var(--color-core-border)] text-[#8b949e] hover:text-white hover:border-[#8b949e]"
            )}
          >
            {isLive ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
            {isLive ? 'Pausar Live Tail' : 'Live Tail Parado'}
          </button>
        </header>

        {/* The Grid */}
        <LogTable />

      </div>
    </div>
  );
}
