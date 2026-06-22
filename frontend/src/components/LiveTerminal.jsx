import React, { useRef, useEffect, useState } from 'react';
import { Terminal, Trash2, Download, Pause, Play } from 'lucide-react';

function LiveTerminal({ logs, clearLogs, darkMode, isConnected }) {
  const terminalRef = useRef(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const [filter, setFilter] = useState('all');
  const [hoveredIndex, setHoveredIndex] = useState(null);

  // Auto-scroll to bottom
  useEffect(() => {
    if (autoScroll && terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  // Filter logs
  const filteredLogs = logs.filter(log => {
    if (filter === 'all') return true;
    return log.level?.toLowerCase() === filter;
  });

  // Get level color and glow
  const getLevelStyles = (level) => {
    switch (level?.toUpperCase()) {
      case 'ERROR':
        return {
          textClass: 'terminal-text-red',
          glowColor: '#f87171',
          glowLight: 'rgba(248, 113, 113, 0.15)'
        };
      case 'WARNING':
        return {
          textClass: 'terminal-text-yellow',
          glowColor: '#facc15',
          glowLight: 'rgba(250, 204, 21, 0.15)'
        };
      case 'INFO':
        return {
          textClass: 'terminal-text-green',
          glowColor: '#4ade80',
          glowLight: 'rgba(74, 222, 128, 0.15)'
        };
      case 'DEBUG':
        return {
          textClass: 'terminal-text-gray',
          glowColor: '#94a3b8',
          glowLight: 'rgba(148, 163, 184, 0.1)'
        };
      default:
        return {
          textClass: 'text-slate-400',
          glowColor: '#94a3b8',
          glowLight: 'rgba(148, 163, 184, 0.1)'
        };
    }
  };

  // Download logs
  const handleDownload = () => {
    const content = filteredLogs
      .map(log => `[${log.timestamp}] [${log.level}] [${log.component}] ${log.message}`)
      .join('\n');
    
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `logs_${new Date().toISOString().split('T')[0]}.txt`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <div className={`${darkMode ? 'bg-slate-800 border-slate-700' : 'bg-white border-gray-200'} border rounded-xl overflow-hidden`}>
      {/* Header */}
      <div className={`${darkMode ? 'bg-slate-900 border-slate-700' : 'bg-gray-100 border-gray-200'} border-b px-4 py-3`}>
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <Terminal className={`w-5 h-5 ${darkMode ? 'text-green-400' : 'text-green-600'}`} />
            <span className={`font-medium ${darkMode ? 'text-white' : 'text-gray-900'}`}>
              Live Terminal
            </span>
            <span className={`text-xs px-2 py-1 rounded-full ${
              isConnected 
                ? 'bg-green-500/20 text-green-400' 
                : 'bg-red-500/20 text-red-400'
            }`}>
              {isConnected ? 'Connected' : 'Disconnected'}
            </span>
          </div>

          <div className="flex items-center space-x-2">
            {/* Filter */}
            <select
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              className={`text-xs px-2 py-1 rounded ${
                darkMode 
                  ? 'bg-slate-700 text-slate-300 border-slate-600' 
                  : 'bg-gray-200 text-gray-700 border-gray-300'
              } border`}
            >
              <option value="all">All Levels</option>
              <option value="error">Errors</option>
              <option value="warning">Warnings</option>
              <option value="info">Info</option>
              <option value="debug">Debug</option>
            </select>

            {/* Auto-scroll toggle */}
            <button
              onClick={() => setAutoScroll(!autoScroll)}
              className={`p-2 rounded ${
                autoScroll 
                  ? 'bg-green-500/20 text-green-400' 
                  : darkMode ? 'bg-slate-700 text-slate-400' : 'bg-gray-200 text-gray-600'
              }`}
              title={autoScroll ? 'Pause auto-scroll' : 'Resume auto-scroll'}
            >
              {autoScroll ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
            </button>

            {/* Download */}
            <button
              onClick={handleDownload}
              className={`p-2 rounded ${darkMode ? 'bg-slate-700 hover:bg-slate-600 text-slate-400' : 'bg-gray-200 hover:bg-gray-300 text-gray-600'}`}
              title="Download logs"
            >
              <Download className="w-4 h-4" />
            </button>

            {/* Clear */}
            <button
              onClick={clearLogs}
              className={`p-2 rounded ${darkMode ? 'bg-slate-700 hover:bg-slate-600 text-slate-400' : 'bg-gray-200 hover:bg-gray-300 text-gray-600'}`}
              title="Clear logs"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      {/* Terminal Content */}
      <div
        ref={terminalRef}
        className={`${darkMode ? 'bg-slate-900' : 'bg-gray-900'} p-4 h-96 overflow-y-auto terminal-scrollbar`}
        style={{ fontFamily: "'Consolas', 'Monaco', 'Courier New', monospace" }}
      >
        {filteredLogs.length === 0 ? (
          <div className="text-slate-500 text-center py-8">
            {isConnected ? 'Waiting for logs...' : 'Connecting to server...'}
          </div>
        ) : (
          filteredLogs.map((log, index) => {
            const styles = getLevelStyles(log.level);
            const isHovered = hoveredIndex === index;

            return (
              <div 
                key={index} 
                className="flex space-x-2 px-2 py-1 rounded cursor-pointer text-xs"
                onMouseEnter={() => setHoveredIndex(index)}
                onMouseLeave={() => setHoveredIndex(null)}
                style={{
                  background: isHovered 
                    ? `linear-gradient(90deg, ${styles.glowLight}, transparent)` 
                    : 'transparent',
                  borderLeft: isHovered ? `2px solid ${styles.glowColor}` : '2px solid transparent',
                  boxShadow: isHovered 
                    ? `0 0 20px ${styles.glowColor}40, inset 0 0 25px ${styles.glowLight}`
                    : 'none',
                  transition: 'all 0.25s cubic-bezier(0.4, 0, 0.2, 1)'
                }}
              >
                <span className="terminal-text-timestamp whitespace-nowrap">
                  [{log.timestamp}]
                </span>
                <span className={`font-bold whitespace-nowrap ${styles.textClass}`}>
                  [{log.level?.padEnd(7)}]
                </span>
                <span className="text-cyan-400 whitespace-nowrap">
                  [{log.component}]
                </span>
                <span 
                  className={`break-all ${isHovered ? 'text-white' : styles.textClass}`}
                  style={{
                    textShadow: isHovered ? `0 0 8px ${styles.glowColor}` : 'none',
                    transition: 'all 0.25s ease'
                  }}
                >
                  {log.message}
                </span>
              </div>
            );
          })
        )}
      </div>

      {/* Footer */}
      <div className={`${darkMode ? 'bg-slate-900 border-slate-700' : 'bg-gray-100 border-gray-200'} border-t px-4 py-2`}>
        <div className="flex items-center justify-between text-xs">
          <span className={darkMode ? 'text-slate-500' : 'text-gray-500'}>
            {filteredLogs.length} log entries
          </span>
          <span className={darkMode ? 'text-slate-500' : 'text-gray-500'}>
            Auto-scroll: {autoScroll ? 'ON' : 'OFF'}
          </span>
        </div>
      </div>
    </div>
  );
}

export default LiveTerminal;