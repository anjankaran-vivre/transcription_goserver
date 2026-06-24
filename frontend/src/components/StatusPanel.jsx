import React, { useState } from 'react';
import { motion } from 'framer-motion';

function StatusPanel({ status, darkMode = true, wsConnected = false }) {
  if (!status) return null;

  const [hoveredIndex, setHoveredIndex] = useState(null);

  const items = [
    { 
      label: 'Status', 
      value: status.status === 'running' ? 'Online' : 'Offline', 
      icon: '🌐', 
      statusType: status.status === 'running' ? 'success' : 'error',
      glowColor: status.status === 'running' ? 'rgba(34, 197, 94, 0.4)' : 'rgba(239, 68, 68, 0.4)'
    },
    { 
      label: 'WebSocket', 
      value: wsConnected ? 'Connected' : 'Disconnected', 
      icon: '🔌', 
      statusType: wsConnected ? 'success' : 'error',
      glowColor: wsConnected ? 'rgba(34, 197, 94, 0.4)' : 'rgba(239, 68, 68, 0.4)'
    },
    { 
      label: 'Uptime', 
      value: status.uptime || '00:00:00', 
      icon: '⏱️', 
      statusType: 'info',
      glowColor: 'rgba(59, 130, 246, 0.4)'
    },
    { 
      label: 'Workers', 
      value: `${status.workers} active`, 
      icon: '👷', 
      statusType: 'info',
      glowColor: 'rgba(59, 130, 246, 0.4)'
    },
    { 
      label: 'Queue', 
      value: `${status.queue_size} pending`, 
      icon: '📥', 
      statusType: status.queue_size > 10 ? 'warning' : 'info',
      glowColor: status.queue_size > 10 ? 'rgba(234, 179, 8, 0.4)' : 'rgba(59, 130, 246, 0.4)'
    },
    { 
      label: 'Memory', 
      value: `${status.memory_mb} MB`, 
      icon: '💾', 
      statusType: status.memory_mb > 500 ? 'warning' : 'info',
      glowColor: status.memory_mb > 500 ? 'rgba(234, 179, 8, 0.4)' : 'rgba(59, 130, 246, 0.4)'
    }
  ];

  const statusStyles = {
    success: {
      bg: darkMode ? 'rgba(34, 197, 94, 0.1)' : 'rgba(34, 197, 94, 0.1)',
      border: darkMode ? 'rgba(34, 197, 94, 0.3)' : 'rgba(34, 197, 94, 0.3)',
      text: darkMode ? 'text-green-400' : 'text-green-600',
    },
    error: {
      bg: darkMode ? 'rgba(239, 68, 68, 0.1)' : 'rgba(239, 68, 68, 0.1)',
      border: darkMode ? 'rgba(239, 68, 68, 0.3)' : 'rgba(239, 68, 68, 0.3)',
      text: darkMode ? 'text-red-400' : 'text-red-600',
    },
    warning: {
      bg: darkMode ? 'rgba(234, 179, 8, 0.1)' : 'rgba(234, 179, 8, 0.1)',
      border: darkMode ? 'rgba(234, 179, 8, 0.3)' : 'rgba(234, 179, 8, 0.3)',
      text: darkMode ? 'text-yellow-400' : 'text-yellow-600',
    },
    info: {
      bg: darkMode ? 'rgba(59, 130, 246, 0.1)' : 'rgba(59, 130, 246, 0.1)',
      border: darkMode ? 'rgba(59, 130, 246, 0.3)' : 'rgba(59, 130, 246, 0.3)',
      text: darkMode ? 'text-blue-400' : 'text-blue-600',
    }
  };

  return (
    <motion.div 
      className={`rounded-2xl p-6 panel-hover border ${
        darkMode 
          ? 'bg-slate-800/50 border-slate-700/50' 
          : 'bg-white border-gray-200'
      }`}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
    >
      {/* Header */}
      <div className="flex items-center space-x-3 mb-6">
        <motion.div 
          className={`w-3 h-3 rounded-full ${status.status === 'running' ? 'bg-green-500' : 'bg-red-500'}`}
          animate={status.status === 'running' ? { 
            scale: [1, 1.3, 1],
            opacity: [1, 0.7, 1]
          } : {}}
          transition={{ duration: 1.5, repeat: Infinity }}
          style={{
            boxShadow: status.status === 'running' 
              ? '0 0 15px rgba(34, 197, 94, 0.6)' 
              : '0 0 15px rgba(239, 68, 68, 0.6)'
          }}
        />
        <h2 className={`text-lg font-semibold ${darkMode ? 'text-white' : 'text-gray-900'}`}>
          System Status
        </h2>
        <span className={`text-xs px-2 py-1 rounded-full ${
          status.status === 'running' 
            ? 'bg-green-500/20 text-green-400' 
            : 'bg-red-500/20 text-red-400'
        }`}>
          {status.status === 'running' ? 'All Systems Operational' : 'System Down'}
        </span>
      </div>

      {/* Status Grid */}
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
        {items.map((item, index) => {
          const styles = statusStyles[item.statusType];
          const isHovered = hoveredIndex === index;
          
          return (
            <motion.div
              key={index}
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: index * 0.08 }}
              onMouseEnter={() => setHoveredIndex(index)}
              onMouseLeave={() => setHoveredIndex(null)}
              whileTap={{ scale: 0.95 }}
              className={`status-card glow-${item.statusType === 'success' ? 'green' : item.statusType === 'error' ? 'red' : item.statusType === 'warning' ? 'yellow' : 'blue'} rounded-xl p-4 cursor-pointer border`}
              style={{
                background: styles.bg,
                borderColor: isHovered ? styles.border : 'transparent',
                boxShadow: isHovered 
                  ? `0 0 25px ${item.glowColor}, 0 5px 20px rgba(0,0,0,0.2)` 
                  : 'none',
                transform: isHovered ? 'translateY(-4px) scale(1.02)' : 'translateY(0) scale(1)',
                transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
              }}
            >
              <div className="flex items-center space-x-2 mb-2">
                <motion.span 
                  className="text-lg"
                  animate={isHovered ? { rotate: [0, 10, -10, 0] } : {}}
                  transition={{ duration: 0.5 }}
                >
                  {item.icon}
                </motion.span>
                <span className={`text-xs font-medium ${darkMode ? 'text-slate-400' : 'text-gray-600'}`}>
                  {item.label}
                </span>
              </div>
              <p className={`text-lg font-bold ${styles.text}`}>
                {item.value}
              </p>
            </motion.div>
          );
        })}
      </div>
    </motion.div>
  );
}

export default StatusPanel;