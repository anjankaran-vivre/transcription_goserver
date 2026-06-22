import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

function CallsTable({ calls = [], darkMode = true, compact = false }) {
  const [expandedRow, setExpandedRow] = useState(null);
  const [hoveredRow, setHoveredRow] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [sortField, setSortField] = useState('timestamp');
  const [sortDirection, setSortDirection] = useState('desc');

  const safeCalls = calls || [];

  const filteredCalls = safeCalls.filter(call => {
    if (!searchTerm) return true;
    return (
      call.call_id?.toString().includes(searchTerm) ||
      call.status?.toLowerCase().includes(searchTerm.toLowerCase()) ||
      call.transcription?.toLowerCase().includes(searchTerm.toLowerCase())
    );
  });

  const sortedCalls = [...filteredCalls].sort((a, b) => {
    let aVal = a[sortField] || '';
    let bVal = b[sortField] || '';
    
    if (sortField === 'timestamp') {
      aVal = new Date(aVal);
      bVal = new Date(bVal);
    }
    
    if (sortDirection === 'asc') {
      return aVal > bVal ? 1 : -1;
    }
    return aVal < bVal ? 1 : -1;
  });

  const getStatusConfig = (status) => {
    const configs = {
      success: {
        text: darkMode ? 'text-green-400' : 'text-green-600',
        bg: 'rgba(34, 197, 94, 0.1)',
        border: 'rgba(34, 197, 94, 0.3)',
        glow: 'rgba(34, 197, 94, 0.4)',
        icon: '✅'
      },
      error: {
        text: darkMode ? 'text-red-400' : 'text-red-600',
        bg: 'rgba(239, 68, 68, 0.1)',
        border: 'rgba(239, 68, 68, 0.3)',
        glow: 'rgba(239, 68, 68, 0.4)',
        icon: '❌'
      },
      download_failed: {
        text: darkMode ? 'text-red-400' : 'text-red-600',
        bg: 'rgba(239, 68, 68, 0.1)',
        border: 'rgba(239, 68, 68, 0.3)',
        glow: 'rgba(239, 68, 68, 0.4)',
        icon: '❌'
      },
      no_speech: {
        text: darkMode ? 'text-yellow-400' : 'text-yellow-600',
        bg: 'rgba(234, 179, 8, 0.1)',
        border: 'rgba(234, 179, 8, 0.3)',
        glow: 'rgba(234, 179, 8, 0.4)',
        icon: '🔇'
      },
      unclear_audio: {
        text: darkMode ? 'text-orange-400' : 'text-orange-600',
        bg: 'rgba(249, 115, 22, 0.1)',
        border: 'rgba(249, 115, 22, 0.3)',
        glow: 'rgba(249, 115, 22, 0.4)',
        icon: '🔊'
      },
      processing: {
        text: darkMode ? 'text-blue-400' : 'text-blue-600',
        bg: 'rgba(59, 130, 246, 0.1)',
        border: 'rgba(59, 130, 246, 0.3)',
        glow: 'rgba(59, 130, 246, 0.4)',
        icon: '⏳'
      }
    };
    return configs[status] || {
      text: darkMode ? 'text-slate-400' : 'text-gray-500',
      bg: 'rgba(100, 116, 139, 0.1)',
      border: 'rgba(100, 116, 139, 0.3)',
      glow: 'rgba(100, 116, 139, 0.4)',
      icon: '❓'
    };
  };

  const handleSort = (field) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  return (
    <div className="space-y-4">
      {/* Search Input with glow */}
      {!compact && (
        <div className="relative">
          <span className="absolute left-4 top-1/2 -translate-y-1/2 text-sm">🔍</span>
          <input
            type="text"
            placeholder="Search calls by ID, status, or transcription..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className={`w-full pl-12 pr-4 py-3 rounded-xl text-sm input-glow border ${
              darkMode 
                ? 'bg-slate-700/50 border-slate-600 text-white placeholder-slate-500' 
                : 'bg-gray-100 border-gray-200 text-gray-900 placeholder-gray-400'
            } focus:outline-none transition-all duration-300`}
            style={{
              boxShadow: searchTerm ? '0 0 20px rgba(59, 130, 246, 0.3)' : 'none'
            }}
          />
          {searchTerm && (
            <button
              onClick={() => setSearchTerm('')}
              className={`absolute right-4 top-1/2 -translate-y-1/2 text-sm ${
                darkMode ? 'text-slate-400 hover:text-white' : 'text-gray-400 hover:text-gray-600'
              }`}
            >
              ✕
            </button>
          )}
        </div>
      )}

      {/* Results count */}
      {!compact && searchTerm && (
        <p className={`text-sm ${darkMode ? 'text-slate-400' : 'text-gray-500'}`}>
          Found {sortedCalls.length} result{sortedCalls.length !== 1 ? 's' : ''}
        </p>
      )}

      {/* Table Container with glow */}
      <motion.div 
        className={`overflow-hidden rounded-2xl border ${
          darkMode 
            ? 'bg-slate-800/30 border-slate-700/50' 
            : 'bg-white border-gray-200'
        }`}
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <div className="overflow-x-auto">
          <table className="w-full">
            {/* Table Header */}
            <thead>
              <tr className={`${darkMode ? 'bg-slate-800/80' : 'bg-gray-50'}`}>
                <th 
                  className={`text-left py-4 px-5 text-sm font-semibold cursor-pointer transition-all duration-200 ${
                    darkMode 
                      ? 'text-slate-300 hover:text-blue-400 hover:bg-blue-500/10' 
                      : 'text-gray-600 hover:text-blue-600 hover:bg-blue-50'
                  }`}
                  onClick={() => handleSort('timestamp')}
                >
                  <span className="flex items-center gap-2">
                    <span>🕐</span>
                    Timestamp 
                    {sortField === 'timestamp' && (
                      <motion.span 
                        className="text-blue-400"
                        initial={{ scale: 0 }}
                        animate={{ scale: 1 }}
                      >
                        {sortDirection === 'asc' ? '↑' : '↓'}
                      </motion.span>
                    )}
                  </span>
                </th>
                <th 
                  className={`text-left py-4 px-5 text-sm font-semibold cursor-pointer transition-all duration-200 ${
                    darkMode 
                      ? 'text-slate-300 hover:text-blue-400 hover:bg-blue-500/10' 
                      : 'text-gray-600 hover:text-blue-600 hover:bg-blue-50'
                  }`}
                  onClick={() => handleSort('call_id')}
                >
                  <span className="flex items-center gap-2">
                    <span>📞</span>
                    Call ID
                    {sortField === 'call_id' && (
                      <motion.span 
                        className="text-blue-400"
                        initial={{ scale: 0 }}
                        animate={{ scale: 1 }}
                      >
                        {sortDirection === 'asc' ? '↑' : '↓'}
                      </motion.span>
                    )}
                  </span>
                </th>
                <th 
                  className={`text-left py-4 px-5 text-sm font-semibold cursor-pointer transition-all duration-200 ${
                    darkMode 
                      ? 'text-slate-300 hover:text-blue-400 hover:bg-blue-500/10' 
                      : 'text-gray-600 hover:text-blue-600 hover:bg-blue-50'
                  }`}
                  onClick={() => handleSort('status')}
                >
                  <span className="flex items-center gap-2">
                    <span>📊</span>
                    Status
                    {sortField === 'status' && (
                      <motion.span 
                        className="text-blue-400"
                        initial={{ scale: 0 }}
                        animate={{ scale: 1 }}
                      >
                        {sortDirection === 'asc' ? '↑' : '↓'}
                      </motion.span>
                    )}
                  </span>
                </th>
                {!compact && (
                  <>
                    <th className={`text-left py-4 px-5 text-sm font-semibold ${
                      darkMode ? 'text-slate-300' : 'text-gray-600'
                    }`}>
                      <span className="flex items-center gap-2">
                        <span>⏱️</span>
                        Duration
                      </span>
                    </th>
                    <th className={`text-left py-4 px-5 text-sm font-semibold ${
                      darkMode ? 'text-slate-300' : 'text-gray-600'
                    }`}>
                      <span className="flex items-center gap-2">
                        <span>📝</span>
                        Words
                      </span>
                    </th>
                    <th className={`text-left py-4 px-5 text-sm font-semibold ${
                      darkMode ? 'text-slate-300' : 'text-gray-600'
                    }`}>
                      <span className="flex items-center gap-2">
                        <span>💬</span>
                        Preview
                      </span>
                    </th>
                  </>
                )}
              </tr>
            </thead>

            {/* Table Body */}
            <tbody>
              {sortedCalls.length === 0 ? (
                <tr>
                  <td 
                    colSpan={compact ? 3 : 6} 
                    className={`text-center py-16 ${darkMode ? 'text-slate-500' : 'text-gray-400'}`}
                  >
                    <motion.div 
                      className="flex flex-col items-center gap-3"
                      initial={{ opacity: 0, scale: 0.9 }}
                      animate={{ opacity: 1, scale: 1 }}
                    >
                      <span className="text-5xl">📭</span>
                      <span className="text-lg font-medium">No calls found</span>
                      <span className="text-sm opacity-70">
                        {searchTerm ? 'Try adjusting your search' : 'Calls will appear here'}
                      </span>
                    </motion.div>
                  </td>
                </tr>
              ) : (
                sortedCalls.map((call, index) => {
                  const statusConfig = getStatusConfig(call.status);
                  const isHovered = hoveredRow === index;
                  const isExpanded = expandedRow === index;

                  return (
                    <React.Fragment key={call.call_id || index}>
                      {/* Main Row with Glow Effect */}
                      <motion.tr
                        initial={{ opacity: 0, x: -20 }}
                        animate={{ opacity: 1, x: 0 }}
                        transition={{ delay: index * 0.03 }}
                        onMouseEnter={() => setHoveredRow(index)}
                        onMouseLeave={() => setHoveredRow(null)}
                        onClick={() => !compact && setExpandedRow(isExpanded ? null : index)}
                        className={`cursor-pointer relative ${
                          darkMode ? 'border-b border-slate-700/30' : 'border-b border-gray-100'
                        }`}
                        style={{
                          background: isHovered 
                            ? `linear-gradient(90deg, ${statusConfig.bg}, transparent)` 
                            : 'transparent',
                          boxShadow: isHovered 
                            ? `inset 0 0 30px ${statusConfig.glow}, 0 0 20px ${statusConfig.glow}` 
                            : 'none',
                          transform: isHovered ? 'scale(1.01)' : 'scale(1)',
                          transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
                          zIndex: isHovered ? 10 : 1,
                          position: 'relative'
                        }}
                      >
                        {/* Left border indicator */}
                        <td 
                          className="relative py-4 px-5"
                          style={{
                            borderLeft: isHovered ? `3px solid ${statusConfig.border}` : '3px solid transparent'
                          }}
                        >
                          <span className={`text-sm ${darkMode ? 'text-slate-300' : 'text-gray-700'}`}>
                            {call.timestamp || '--'}
                          </span>
                        </td>
                        
                        <td className="py-4 px-5">
                          <span 
                            className={`text-sm font-mono px-3 py-1.5 rounded-lg ${
                              darkMode ? 'bg-slate-700/50 text-slate-300' : 'bg-gray-100 text-gray-700'
                            }`}
                            style={{
                              boxShadow: isHovered ? `0 0 10px ${statusConfig.glow}` : 'none'
                            }}
                          >
                            {call.call_id || '--'}
                          </span>
                        </td>
                        
                        <td className="py-4 px-5">
                          <motion.span 
                            className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-sm font-medium ${statusConfig.text}`}
                            style={{
                              background: statusConfig.bg,
                              border: `1px solid ${statusConfig.border}`,
                              boxShadow: isHovered ? `0 0 15px ${statusConfig.glow}` : 'none'
                            }}
                            animate={isHovered ? { scale: [1, 1.05, 1] } : {}}
                            transition={{ duration: 0.3 }}
                          >
                            <span>{statusConfig.icon}</span>
                            <span>{call.status || '--'}</span>
                          </motion.span>
                        </td>
                        
                        {!compact && (
                          <>
                            <td className={`py-4 px-5 text-sm ${darkMode ? 'text-slate-300' : 'text-gray-700'}`}>
                              <span className="flex items-center gap-1">
                                {call.duration_sec || '0'}s
                              </span>
                            </td>
                            <td className={`py-4 px-5 text-sm ${darkMode ? 'text-slate-300' : 'text-gray-700'}`}>
                              {call.word_count || '0'}
                            </td>
                            <td className={`py-4 px-5 text-sm max-w-xs ${darkMode ? 'text-slate-400' : 'text-gray-500'}`}>
                              <span className="truncate block">
                                {(call.transcription || '').substring(0, 50)}
                                {(call.transcription || '').length > 50 && '...'}
                              </span>
                            </td>
                          </>
                        )}

                        {/* Expand indicator */}
                        {!compact && (
                          <td className="py-4 px-2">
                            <motion.span
                              animate={{ rotate: isExpanded ? 180 : 0 }}
                              className={`text-sm ${darkMode ? 'text-slate-500' : 'text-gray-400'}`}
                            >
                              ▼
                            </motion.span>
                          </td>
                        )}
                      </motion.tr>
                      
                      {/* Expanded Row */}
                      <AnimatePresence>
                        {isExpanded && !compact && (
                          <motion.tr
                            initial={{ opacity: 0, height: 0 }}
                            animate={{ opacity: 1, height: 'auto' }}
                            exit={{ opacity: 0, height: 0 }}
                            transition={{ duration: 0.3 }}
                            style={{
                              background: `linear-gradient(135deg, ${statusConfig.bg}, transparent)`
                            }}
                          >
                            <td colSpan={7} className="px-6 py-6">
                              <motion.div 
                                className="space-y-4"
                                initial={{ opacity: 0, y: -10 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ delay: 0.1 }}
                              >
                                {/* Transcription Card */}
                                <div 
                                  className={`rounded-xl p-5 border ${
                                    darkMode 
                                      ? 'bg-slate-800/50 border-slate-700/50' 
                                      : 'bg-white border-gray-200'
                                  }`}
                                  style={{
                                    boxShadow: `0 0 20px ${statusConfig.glow}`
                                  }}
                                >
                                  <h4 className={`text-sm font-semibold mb-3 flex items-center gap-2 ${
                                    darkMode ? 'text-slate-200' : 'text-gray-800'
                                  }`}>
                                    <span>📝</span> Transcription
                                  </h4>
                                  <p className={`text-sm leading-relaxed ${
                                    darkMode ? 'text-slate-400' : 'text-gray-600'
                                  }`}>
                                    {call.transcription || 'No transcription available'}
                                  </p>
                                </div>
                                
                                {/* Summary Card */}
                                <div 
                                  className={`rounded-xl p-5 border ${
                                    darkMode 
                                      ? 'bg-slate-800/50 border-slate-700/50' 
                                      : 'bg-white border-gray-200'
                                  }`}
                                  style={{
                                    boxShadow: `0 0 20px ${statusConfig.glow}`
                                  }}
                                >
                                  <h4 className={`text-sm font-semibold mb-3 flex items-center gap-2 ${
                                    darkMode ? 'text-slate-200' : 'text-gray-800'
                                  }`}>
                                    <span>📋</span> Summary
                                  </h4>
                                  <p className={`text-sm leading-relaxed whitespace-pre-wrap ${
                                    darkMode ? 'text-slate-400' : 'text-gray-600'
                                  }`}>
                                    {call.summary || 'No summary available'}
                                  </p>
                                </div>
                                
                                {/* Error Message */}
                                {call.error_message && (
                                  <motion.div 
                                    className={`rounded-xl p-5 border ${
                                      darkMode 
                                        ? 'bg-red-500/10 border-red-500/30' 
                                        : 'bg-red-50 border-red-200'
                                    }`}
                                    initial={{ opacity: 0, scale: 0.95 }}
                                    animate={{ opacity: 1, scale: 1 }}
                                    style={{
                                      boxShadow: '0 0 20px rgba(239, 68, 68, 0.3)'
                                    }}
                                  >
                                    <h4 className="text-sm font-semibold mb-3 flex items-center gap-2 text-red-400">
                                      <span>⚠️</span> Error Details
                                    </h4>
                                    <p className={`text-sm ${darkMode ? 'text-red-300' : 'text-red-600'}`}>
                                      {call.error_message}
                                    </p>
                                  </motion.div>
                                )}

                                {/* Meta Info */}
                                <div className="flex flex-wrap gap-4 pt-2">
                                  <span className={`text-xs px-3 py-1.5 rounded-full ${
                                    darkMode ? 'bg-slate-700/50 text-slate-400' : 'bg-gray-100 text-gray-500'
                                  }`}>
                                    Duration: {call.duration_sec || 0}s
                                  </span>
                                  <span className={`text-xs px-3 py-1.5 rounded-full ${
                                    darkMode ? 'bg-slate-700/50 text-slate-400' : 'bg-gray-100 text-gray-500'
                                  }`}>
                                    Words: {call.word_count || 0}
                                  </span>
                                  <span className={`text-xs px-3 py-1.5 rounded-full ${
                                    darkMode ? 'bg-slate-700/50 text-slate-400' : 'bg-gray-100 text-gray-500'
                                  }`}>
                                    ID: {call.call_id}
                                  </span>
                                </div>
                              </motion.div>
                            </td>
                          </motion.tr>
                        )}
                      </AnimatePresence>
                    </React.Fragment>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </motion.div>
    </div>
  );
}

export default CallsTable;