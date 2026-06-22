import React, { useState } from 'react';
import { motion } from 'framer-motion';

function TokenUsage({ metrics, darkMode = true }) {
  if (!metrics?.rate_limit) return null;

  const { rate_limit } = metrics;
  const [hoveredCard, setHoveredCard] = useState(null);

  const getColor = (percentage) => {
    if (percentage >= 80) return { 
      bar: 'progress-bar-red',
      bg: 'rgba(239, 68, 68, 0.1)',
      border: 'rgba(239, 68, 68, 0.3)',
      text: darkMode ? 'text-red-400' : 'text-red-600',
      glow: 'rgba(239, 68, 68, 0.4)'
    };
    if (percentage >= 50) return { 
      bar: 'progress-bar-yellow',
      bg: 'rgba(234, 179, 8, 0.1)',
      border: 'rgba(234, 179, 8, 0.3)',
      text: darkMode ? 'text-yellow-400' : 'text-yellow-600',
      glow: 'rgba(234, 179, 8, 0.4)'
    };
    return { 
      bar: 'progress-bar-green',
      bg: 'rgba(34, 197, 94, 0.1)',
      border: 'rgba(34, 197, 94, 0.3)',
      text: darkMode ? 'text-green-400' : 'text-green-600',
      glow: 'rgba(34, 197, 94, 0.4)'
    };
  };

  const dailyColors = getColor(rate_limit.daily_percentage);
  const minuteColors = getColor(rate_limit.minute_percentage);

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
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center space-x-3">
          <motion.span 
            className="text-xl"
            animate={{ rotate: 360 }}
            transition={{ duration: 3, repeat: Infinity, ease: "linear" }}
          >
            ⚡
          </motion.span>
          <h3 className={`text-lg font-semibold ${darkMode ? 'text-white' : 'text-gray-900'}`}>
            Groq API Usage
          </h3>
        </div>
        {(rate_limit.daily_percentage >= 80 || rate_limit.minute_percentage >= 80) && (
          <motion.div 
            className={`flex items-center space-x-2 px-4 py-2 rounded-full border ${
              darkMode 
                ? 'bg-yellow-500/10 border-yellow-500/30' 
                : 'bg-yellow-50 border-yellow-200'
            }`}
            animate={{ scale: [1, 1.02, 1] }}
            transition={{ duration: 1.5, repeat: Infinity }}
            style={{ boxShadow: '0 0 15px rgba(234, 179, 8, 0.3)' }}
          >
            <span className="text-sm">⚠️</span>
            <span className={`text-sm font-medium ${darkMode ? 'text-yellow-400' : 'text-yellow-600'}`}>
              High Usage
            </span>
          </motion.div>
        )}
      </div>

      {/* Usage Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Daily Usage */}
        <motion.div 
          className={`usage-card rounded-xl p-5 border ${
            darkMode 
              ? 'bg-slate-700/30 border-slate-600/50' 
              : 'bg-gray-50 border-gray-200'
          }`}
          onMouseEnter={() => setHoveredCard('daily')}
          onMouseLeave={() => setHoveredCard(null)}
          style={{
            boxShadow: hoveredCard === 'daily' ? `0 0 25px ${dailyColors.glow}` : 'none',
            borderColor: hoveredCard === 'daily' ? dailyColors.border : undefined,
            transform: hoveredCard === 'daily' ? 'scale(1.02)' : 'scale(1)',
            transition: 'all 0.3s ease'
          }}
        >
          <div className="flex items-center justify-between mb-4">
            <span className={`text-sm font-medium flex items-center gap-2 ${darkMode ? 'text-slate-300' : 'text-gray-700'}`}>
              <span>📅</span> Daily Requests
            </span>
            <motion.span 
              className={`text-xl font-bold ${dailyColors.text}`}
              animate={{ scale: [1, 1.05, 1] }}
              transition={{ duration: 2, repeat: Infinity }}
            >
              {rate_limit.daily_percentage.toFixed(1)}%
            </motion.span>
          </div>
          
          <div className={`h-4 rounded-full overflow-hidden ${darkMode ? 'bg-slate-600/50' : 'bg-gray-200'}`}>
            <motion.div 
              className={`h-full rounded-full ${dailyColors.bar}`}
              initial={{ width: 0 }}
              animate={{ width: `${Math.min(rate_limit.daily_percentage, 100)}%` }}
              transition={{ duration: 1, ease: "easeOut" }}
            />
          </div>
          
          <div className={`flex justify-between mt-3 text-xs ${darkMode ? 'text-slate-400' : 'text-gray-500'}`}>
            <span className="flex items-center gap-1">
              <span className={dailyColors.text}>●</span> {rate_limit.daily_used.toLocaleString()} used
            </span>
            <span>{rate_limit.daily_limit.toLocaleString()} limit</span>
          </div>
        </motion.div>

        {/* Per-Minute Usage */}
        <motion.div 
          className={`usage-card rounded-xl p-5 border ${
            darkMode 
              ? 'bg-slate-700/30 border-slate-600/50' 
              : 'bg-gray-50 border-gray-200'
          }`}
          onMouseEnter={() => setHoveredCard('minute')}
          onMouseLeave={() => setHoveredCard(null)}
          style={{
            boxShadow: hoveredCard === 'minute' ? `0 0 25px ${minuteColors.glow}` : 'none',
            borderColor: hoveredCard === 'minute' ? minuteColors.border : undefined,
            transform: hoveredCard === 'minute' ? 'scale(1.02)' : 'scale(1)',
            transition: 'all 0.3s ease'
          }}
        >
          <div className="flex items-center justify-between mb-4">
            <span className={`text-sm font-medium flex items-center gap-2 ${darkMode ? 'text-slate-300' : 'text-gray-700'}`}>
              <span>⏱️</span> Per Minute
            </span>
            <motion.span 
              className={`text-xl font-bold ${minuteColors.text}`}
              animate={{ scale: [1, 1.05, 1] }}
              transition={{ duration: 2, repeat: Infinity }}
            >
              {rate_limit.minute_percentage.toFixed(1)}%
            </motion.span>
          </div>
          
          <div className={`h-4 rounded-full overflow-hidden ${darkMode ? 'bg-slate-600/50' : 'bg-gray-200'}`}>
            <motion.div 
              className={`h-full rounded-full ${minuteColors.bar}`}
              initial={{ width: 0 }}
              animate={{ width: `${Math.min(rate_limit.minute_percentage, 100)}%` }}
              transition={{ duration: 1, ease: "easeOut" }}
            />
          </div>
          
          <div className={`flex justify-between mt-3 text-xs ${darkMode ? 'text-slate-400' : 'text-gray-500'}`}>
            <span className="flex items-center gap-1">
              <span className={minuteColors.text}>●</span> {rate_limit.minute_used} used
            </span>
            <span>{rate_limit.minute_limit} limit</span>
          </div>
        </motion.div>
      </div>

      {/* Critical Warning */}
      {rate_limit.daily_percentage >= 90 && (
        <motion.div 
          className={`mt-6 p-4 rounded-xl flex items-center space-x-3 border ${
            darkMode 
              ? 'bg-red-500/10 border-red-500/30' 
              : 'bg-red-50 border-red-200'
          }`}
          initial={{ opacity: 0, y: 10, scale: 0.95 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          style={{ boxShadow: '0 0 20px rgba(239, 68, 68, 0.3)' }}
        >
          <motion.span 
            className="text-2xl"
            animate={{ scale: [1, 1.2, 1] }}
            transition={{ duration: 0.5, repeat: Infinity }}
          >
            🚨
          </motion.span>
          <span className={`text-sm font-medium ${darkMode ? 'text-red-400' : 'text-red-600'}`}>
            Daily API limit nearly reached! Consider upgrading your Groq plan.
          </span>
        </motion.div>
      )}
    </motion.div>
  );
}

export default TokenUsage;