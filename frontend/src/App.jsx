import React, { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { getStatus, getMetrics, getCalls, restartServer, stopServer, clearQueue, testEmail } from './services/api';
import { useWebSocket } from './hooks/useWebSocket';
import MetricsCards from './components/MetricsCards';
import StatusPanel from './components/StatusPanel';
import LiveTerminal from './components/LiveTerminal';
import CallsTable from './components/CallsTable';
import TokenUsage from './components/TokenUsage';
import Charts from './components/Charts';
import ManualCallEditor from './components/ManualCallEditor';

const pageVariants = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -20 }
};

const containerVariants = {
  animate: {
    transition: { staggerChildren: 0.1 }
  }
};

function App() {
  const [status, setStatus] = useState(null);
  const [metrics, setMetrics] = useState(null);
  const [calls, setCalls] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [darkMode, setDarkMode] = useState(true);
  const [activeTab, setActiveTab] = useState('dashboard');
  const [actionLoading, setActionLoading] = useState(null);
  const [notification, setNotification] = useState(null);

  const { isConnected, logs, clearLogs } = useWebSocket('http://localhost:5050');

  const fetchData = useCallback(async () => {
    try {
      const [statusData, metricsData, callsData] = await Promise.all([
        getStatus(),
        getMetrics(),
        getCalls(100)
      ]);
      setStatus(statusData);
      setMetrics(metricsData);
      setCalls(callsData.calls || []);
      setError(null);
    } catch (err) {
      console.error('Fetch error:', err);
      setError('Failed to connect to server');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, [fetchData]);

  const showNotification = (message, type = 'success') => {
    setNotification({ message, type });
    setTimeout(() => setNotification(null), 3000);
  };

  const handleRestart = async () => {
    if (!confirm('Restart server?')) return;
    setActionLoading('restart');
    try {
      await restartServer();
      showNotification('Server restarting...');
    } catch {
      showNotification('Failed to restart', 'error');
    } finally {
      setActionLoading(null);
    }
  };

  const handleStop = async () => {
    if (!confirm('Stop server?')) return;
    setActionLoading('stop');
    try {
      await stopServer();
      showNotification('Server stopping...');
    } catch {
      showNotification('Failed to stop', 'error');
    } finally {
      setActionLoading(null);
    }
  };

  const handleClearQueue = async () => {
    if (!confirm('Clear queue?')) return;
    setActionLoading('clear');
    try {
      const result = await clearQueue();
      showNotification(`Cleared ${result.cleared} items`);
      fetchData();
    } catch {
      showNotification('Failed to clear', 'error');
    } finally {
      setActionLoading(null);
    }
  };

  const handleTestEmail = async () => {
    setActionLoading('email');
    try {
      await testEmail();
      showNotification('Test email sent!');
    } catch {
      showNotification('Failed to send email', 'error');
    } finally {
      setActionLoading(null);
    }
  };

  const tabs = [
    { id: 'dashboard', label: 'Dashboard', icon: '📊' },
    { id: 'terminal', label: 'Terminal', icon: '💻' },
    { id: 'calls', label: 'Calls', icon: '📋' },
    { id: 'editor', label: 'Edit Call', icon: '🔧' },
    { id: 'settings', label: 'Settings', icon: '⚙️' }
  ];

  if (loading) {
    return (
      <div className={`min-h-screen flex items-center justify-center ${darkMode ? 'bg-slate-900' : 'bg-gray-50'}`}>
        <motion.div 
          className="text-center"
          initial={{ opacity: 0, scale: 0.8 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.5 }}
        >
          <motion.div 
            className={`w-16 h-16 border-4 ${darkMode ? 'border-blue-500' : 'border-blue-600'} border-t-transparent rounded-full mx-auto mb-4`}
            animate={{ rotate: 360 }}
            transition={{ duration: 1, repeat: Infinity, ease: "linear" }}
          />
          <motion.p 
            className={`text-lg ${darkMode ? 'text-slate-400' : 'text-gray-600'}`}
            animate={{ opacity: [0.5, 1, 0.5] }}
            transition={{ duration: 1.5, repeat: Infinity }}
          >
            Loading Dashboard...
          </motion.p>
        </motion.div>
      </div>
    );
  }

  return (
    <div className={`min-h-screen transition-colors duration-500 ${darkMode ? 'bg-slate-900 text-slate-100' : 'bg-gray-50 text-gray-900'}`}>
      {/* Animated Notification */}
      <AnimatePresence>
        {notification && (
          <motion.div
            initial={{ opacity: 0, y: -50, x: '-50%' }}
            animate={{ opacity: 1, y: 0, x: '-50%' }}
            exit={{ opacity: 0, y: -50, x: '-50%' }}
            className={`fixed top-4 left-1/2 z-50 px-6 py-3 rounded-xl shadow-2xl font-medium ${
              notification.type === 'success' 
                ? 'bg-green-500 text-white' 
                : 'bg-red-500 text-white'
            }`}
          >
            {notification.type === 'success' ? '✅' : '❌'} {notification.message}
          </motion.div>
        )}
      </AnimatePresence>

      {/* Header */}
      <motion.header 
        className={`sticky top-0 z-40 backdrop-blur-md border-b ${darkMode ? 'bg-slate-900/90 border-slate-800' : 'bg-white/90 border-gray-200'}`}
        initial={{ y: -100 }}
        animate={{ y: 0 }}
        transition={{ type: "spring", stiffness: 100 }}
      >
        <div className="max-w-7xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            {/* Logo & Status */}
            <motion.div 
              className="flex items-center space-x-3"
              whileHover={{ scale: 1.02 }}
            >
              <motion.span 
                className="text-2xl"
                whileHover={{ rotate: [0, -10, 10, 0] }}
                transition={{ duration: 0.5 }}
              >
                📞
              </motion.span>
              <div>
                <h1 className={`text-xl font-bold ${darkMode ? 'text-white' : 'text-gray-900'}`}>
                  Transcription Server
                </h1>
                <div className="flex items-center space-x-2">
                  <motion.span 
                    className={`w-2 h-2 rounded-full ${status?.status === 'running' ? 'bg-green-500' : 'bg-red-500'}`}
                    animate={status?.status === 'running' ? { scale: [1, 1.2, 1] } : {}}
                    transition={{ duration: 1, repeat: Infinity }}
                  />
                  <span className={`text-sm ${darkMode ? 'text-slate-400' : 'text-gray-500'}`}>
                    {status?.status === 'running' ? 'Online' : 'Offline'}
                    {status?.uptime && ` · ${status.uptime}`}
                  </span>
                </div>
              </div>
            </motion.div>

            {/* Right Side Controls */}
            <div className="flex items-center space-x-4">
              {/* Live Indicator - Green Dot with Pulse */}
              <div className="flex items-center space-x-2">
                <div className="relative">
                  <motion.span 
                    className={`block w-2.5 h-2.5 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`}
                    animate={isConnected ? { 
                      scale: [1, 1.3, 1],
                      opacity: [1, 0.8, 1]
                    } : {}}
                    transition={{ 
                      duration: 1.5, 
                      repeat: Infinity,
                      ease: "easeInOut"
                    }}
                  />
                  {isConnected && (
                    <motion.span 
                      className="absolute inset-0 w-2.5 h-2.5 rounded-full bg-green-500"
                      animate={{ 
                        scale: [1, 2, 2.5],
                        opacity: [0.5, 0.2, 0]
                      }}
                      transition={{ 
                        duration: 1.5, 
                        repeat: Infinity,
                        ease: "easeOut"
                      }}
                    />
                  )}
                </div>
                <span className={`text-sm font-medium ${isConnected ? 'text-green-500' : 'text-red-500'}`}>
                  {isConnected ? 'Live' : 'Offline'}
                </span>
              </div>

              {/* Refresh Button */}
              <motion.button
                onClick={fetchData}
                className={`p-2 rounded-lg ${darkMode ? 'hover:bg-slate-800' : 'hover:bg-gray-100'}`}
                whileHover={{ scale: 1.1 }}
                whileTap={{ scale: 0.9 }}
              >
                <motion.span
                  className="text-lg"
                  whileHover={{ rotate: 180 }}
                  transition={{ duration: 0.3 }}
                >
                  🔄
                </motion.span>
              </motion.button>

              {/* Dark/Light Mode Toggle */}
              <motion.button
                onClick={() => setDarkMode(!darkMode)}
                className={`p-2 rounded-lg ${darkMode ? 'hover:bg-slate-800' : 'hover:bg-gray-100'}`}
                whileHover={{ scale: 1.1 }}
                whileTap={{ scale: 0.9 }}
              >
                <motion.span
                  className="text-lg"
                  animate={{ rotate: darkMode ? 0 : 180 }}
                  transition={{ duration: 0.5 }}
                >
                  {darkMode ? '☀️' : '🌙'}
                </motion.span>
              </motion.button>
            </div>
          </div>

          {/* Animated Tabs */}
          <div className="flex space-x-1 mt-4">
            {tabs.map((tab) => (
              <motion.button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`relative px-4 py-2 rounded-lg font-medium flex items-center space-x-2 transition-colors ${
                  activeTab === tab.id
                    ? 'text-white'
                    : darkMode 
                      ? 'text-slate-400 hover:text-white hover:bg-slate-800' 
                      : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
                }`}
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
              >
                {activeTab === tab.id && (
                  <motion.div
                    className="absolute inset-0 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg"
                    layoutId="activeTab"
                    transition={{ type: "spring", stiffness: 300, damping: 30 }}
                  />
                )}
                <span className="relative z-10 text-sm">{tab.icon}</span>
                <span className="relative z-10">{tab.label}</span>
              </motion.button>
            ))}
          </div>
        </div>
      </motion.header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 py-6">
        <AnimatePresence>
          {error && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              className={`mb-6 p-4 rounded-xl flex items-center space-x-3 ${
                darkMode ? 'bg-red-500/20 border border-red-500/50 text-red-400' : 'bg-red-50 border border-red-200 text-red-600'
              }`}
            >
              <span className="text-xl">⚠️</span>
              <span>{error}</span>
            </motion.div>
          )}
        </AnimatePresence>

        <AnimatePresence mode="wait">
          {/* Dashboard Tab */}
          {activeTab === 'dashboard' && (
            <motion.div
              key="dashboard"
              variants={containerVariants}
              initial="initial"
              animate="animate"
              exit="exit"
              className="space-y-6"
            >
              <motion.div variants={pageVariants}>
                <StatusPanel status={status} darkMode={darkMode} />
              </motion.div>
              
              <motion.div variants={pageVariants}>
                <MetricsCards metrics={metrics} darkMode={darkMode} />
              </motion.div>
              
              <motion.div variants={pageVariants}>
                <TokenUsage metrics={metrics} darkMode={darkMode} />
              </motion.div>
              
              <motion.div variants={pageVariants}>
                <Charts calls={calls} darkMode={darkMode} />
              </motion.div>
              
              <motion.div variants={pageVariants}>
                <div className={`rounded-2xl p-6 ${darkMode ? 'bg-slate-800/50' : 'bg-white border border-gray-200'}`}>
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-semibold flex items-center space-x-2">
                      <span className="text-sm">📋</span>
                      <span>Recent Calls</span>
                    </h3>
                    <motion.button
                      onClick={() => setActiveTab('calls')}
                      className="text-blue-500 text-sm font-medium hover:text-blue-400"
                      whileHover={{ x: 5 }}
                    >
                      View All →
                    </motion.button>
                  </div>
                  <CallsTable calls={calls.slice(0, 5)} darkMode={darkMode} compact />
                </div>
              </motion.div>
            </motion.div>
          )}

          {/* Terminal Tab */}
          {activeTab === 'terminal' && (
            <motion.div
              key="terminal"
              variants={pageVariants}
              initial="initial"
              animate="animate"
              exit="exit"
            >
              <LiveTerminal logs={logs} clearLogs={clearLogs} darkMode={darkMode} isConnected={isConnected} />
            </motion.div>
          )}

          {/* Calls Tab */}
          {activeTab === 'calls' && (
            <motion.div
              key="calls"
              variants={pageVariants}
              initial="initial"
              animate="animate"
              exit="exit"
            >
              <div className={`rounded-2xl p-6 ${darkMode ? 'bg-slate-800/50' : 'bg-white border border-gray-200'}`}>
                <div className="flex items-center justify-between mb-6">
                  <h3 className="text-lg font-semibold flex items-center space-x-2">
                    <span className="text-sm">📋</span>
                    <span>Call History</span>
                  </h3>
                  <span className={`text-sm px-3 py-1 rounded-full ${darkMode ? 'bg-slate-700 text-slate-300' : 'bg-gray-100 text-gray-600'}`}>
                    {calls.length} calls
                  </span>
                </div>
                <CallsTable calls={calls} darkMode={darkMode} />
              </div>
            </motion.div>
          )}

          {/* Manual Call Editor Tab */}
          {activeTab === 'editor' && (
            <motion.div
              key="editor"
              variants={pageVariants}
              initial="initial"
              animate="animate"
              exit="exit"
            >
              <ManualCallEditor darkMode={darkMode} />
            </motion.div>
          )}

          {/* Settings Tab */}
          {activeTab === 'settings' && (
            <motion.div
              key="settings"
              variants={containerVariants}
              initial="initial"
              animate="animate"
              exit="exit"
              className="space-y-6"
            >
              <motion.div variants={pageVariants} className={`rounded-2xl p-6 ${darkMode ? 'bg-slate-800/50' : 'bg-white border border-gray-200'}`}>
                <h3 className="text-lg font-semibold mb-6 flex items-center space-x-2">
                  <span className="text-sm">🎛️</span>
                  <span>Server Controls</span>
                </h3>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  <ControlButton onClick={handleRestart} loading={actionLoading === 'restart'} icon="🔄" label="Restart" color="blue" darkMode={darkMode} />
                  <ControlButton onClick={handleStop} loading={actionLoading === 'stop'} icon="⏹️" label="Stop" color="red" darkMode={darkMode} />
                  <ControlButton onClick={handleClearQueue} loading={actionLoading === 'clear'} icon="🗑️" label="Clear Queue" color="orange" darkMode={darkMode} />
                  <ControlButton onClick={handleTestEmail} loading={actionLoading === 'email'} icon="📧" label="Test Email" color="purple" darkMode={darkMode} />
                </div>
              </motion.div>

              <motion.div variants={pageVariants} className={`rounded-2xl p-6 ${darkMode ? 'bg-slate-800/50' : 'bg-white border border-gray-200'}`}>
                <h3 className="text-lg font-semibold mb-6 flex items-center space-x-2">
                  <span className="text-sm">ℹ️</span>
                  <span>Server Info</span>
                </h3>
                <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                  <InfoCard label="Process ID" value={status?.pid || 'N/A'} darkMode={darkMode} />
                  <InfoCard label="Workers" value={status?.workers || 0} darkMode={darkMode} />
                  <InfoCard label="Queue Size" value={status?.queue_size || 0} darkMode={darkMode} />
                  <InfoCard label="Memory" value={`${status?.memory_mb || 0} MB`} darkMode={darkMode} />
                  <InfoCard label="Uptime" value={status?.uptime || '00:00:00'} darkMode={darkMode} />
                  <InfoCard label="Status" value={status?.status || 'unknown'} darkMode={darkMode} />
                </div>
              </motion.div>

              <motion.div variants={pageVariants} className={`rounded-2xl p-6 ${darkMode ? 'bg-slate-800/50' : 'bg-white border border-gray-200'}`}>
                <h3 className="text-lg font-semibold mb-6 flex items-center space-x-2">
                  <span className="text-sm">🔗</span>
                  <span>API Endpoints</span>
                </h3>
                <div className="space-y-3">
                  <EndpointCard method="POST" path="/server_transcription" description="Transcription webhook" darkMode={darkMode} />
                  <EndpointCard method="GET" path="/health" description="Health check" darkMode={darkMode} />
                  <EndpointCard method="GET" path="/api/status" description="Server status" darkMode={darkMode} />
                  <EndpointCard method="GET" path="/api/metrics" description="Metrics data" darkMode={darkMode} />
                </div>
              </motion.div>
            </motion.div>
          )}
        </AnimatePresence>
      </main>

      {/* Footer */}
      <motion.footer 
        className={`py-4 mt-8 text-center text-sm ${darkMode ? 'text-slate-500 border-t border-slate-800' : 'text-gray-400 border-t border-gray-200'}`}
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.5 }}
      >
        Transcription Server Dashboard · Last updated: {new Date().toLocaleTimeString()}
      </motion.footer>
    </div>
  );
}

// Control Button Component
function ControlButton({ onClick, loading, icon, label, color, darkMode }) {
  const colors = {
    blue: 'from-blue-600 to-blue-700 hover:from-blue-500 hover:to-blue-600',
    red: 'from-red-600 to-red-700 hover:from-red-500 hover:to-red-600',
    orange: 'from-orange-600 to-orange-700 hover:from-orange-500 hover:to-orange-600',
    purple: 'from-purple-600 to-purple-700 hover:from-purple-500 hover:to-purple-600'
  };

  return (
    <motion.button
      onClick={onClick}
      disabled={loading}
      className={`p-4 bg-gradient-to-br ${colors[color]} text-white rounded-xl flex flex-col items-center justify-center space-y-2 disabled:opacity-50 shadow-lg`}
      whileHover={{ scale: 1.05, y: -2 }}
      whileTap={{ scale: 0.95 }}
    >
      <motion.span 
        className="text-xl"
        animate={loading ? { rotate: 360 } : {}}
        transition={{ duration: 1, repeat: loading ? Infinity : 0 }}
      >
        {icon}
      </motion.span>
      <span className="font-medium text-sm">{label}</span>
    </motion.button>
  );
}

// Info Card Component
function InfoCard({ label, value, darkMode }) {
  return (
    <motion.div 
      className={`p-4 rounded-xl ${darkMode ? 'bg-slate-700/50' : 'bg-gray-50'}`}
      whileHover={{ scale: 1.02, y: -2 }}
    >
      <p className={`text-xs ${darkMode ? 'text-slate-400' : 'text-gray-500'}`}>{label}</p>
      <p className="text-lg font-semibold mt-1">{value}</p>
    </motion.div>
  );
}

// Endpoint Card Component
function EndpointCard({ method, path, description, darkMode }) {
  const methodColors = {
    GET: 'bg-green-500/20 text-green-400',
    POST: 'bg-blue-500/20 text-blue-400',
    PUT: 'bg-yellow-500/20 text-yellow-400',
    DELETE: 'bg-red-500/20 text-red-400'
  };

  return (
    <motion.div 
      className={`flex items-center space-x-4 p-3 rounded-lg ${darkMode ? 'bg-slate-700/30 hover:bg-slate-700/50' : 'bg-gray-50 hover:bg-gray-100'}`}
      whileHover={{ x: 5 }}
    >
      <span className={`px-2 py-1 rounded text-xs font-bold ${methodColors[method]}`}>
        {method}
      </span>
      <code className={`font-mono text-sm ${darkMode ? 'text-green-400' : 'text-green-600'}`}>{path}</code>
      <span className={`text-sm ${darkMode ? 'text-slate-500' : 'text-gray-400'}`}>{description}</span>
    </motion.div>
  );
}

export default App;