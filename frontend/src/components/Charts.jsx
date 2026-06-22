import React, { useMemo } from 'react';
import { motion } from 'framer-motion';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  BarChart,
  Bar,
  Legend
} from 'recharts';

function Charts({ calls = [], darkMode = true }) {
  const chartData = useMemo(() => {
    if (!calls || calls.length === 0) return null;

    const statusCounts = calls.reduce((acc, call) => {
      const status = call.status || 'unknown';
      acc[status] = (acc[status] || 0) + 1;
      return acc;
    }, {});

    const statusData = Object.entries(statusCounts).map(([name, value]) => ({
      name: name.charAt(0).toUpperCase() + name.slice(1).replace('_', ' '),
      value
    }));

    const hourlyData = [];
    for (let i = 23; i >= 0; i--) {
      const hour = new Date();
      hour.setHours(hour.getHours() - i);
      const hourStr = hour.getHours().toString().padStart(2, '0') + ':00';
      
      const count = calls.filter(call => {
        if (!call.timestamp) return false;
        const callHour = call.timestamp.split(' ')[1]?.split(':')[0];
        return callHour === hour.getHours().toString().padStart(2, '0');
      }).length;
      
      hourlyData.push({ time: hourStr, calls: count });
    }

    const durationRanges = [
      { range: '0-10s', min: 0, max: 10 },
      { range: '10-30s', min: 10, max: 30 },
      { range: '30-60s', min: 30, max: 60 },
      { range: '60s+', min: 60, max: Infinity }
    ];

    const durationData = durationRanges.map(({ range, min, max }) => ({
      range,
      count: calls.filter(call => {
        const duration = parseFloat(call.duration_sec) || 0;
        return duration >= min && duration < max;
      }).length
    }));

    return { statusData, hourlyData, durationData };
  }, [calls]);

  if (!chartData) {
    return (
      <motion.div 
        className={`rounded-2xl p-8 text-center chart-hover ${
          darkMode ? 'bg-slate-800/50 border border-slate-700' : 'bg-white border border-gray-200'
        }`}
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
      >
        <span className="text-4xl">📊</span>
        <p className={`mt-3 ${darkMode ? 'text-slate-500' : 'text-gray-400'}`}>
          No data available for charts
        </p>
      </motion.div>
    );
  }

  const COLORS = ['#22c55e', '#ef4444', '#eab308', '#f97316', '#3b82f6', '#8b5cf6'];

  const CustomTooltip = ({ active, payload, label }) => {
    if (active && payload && payload.length) {
      return (
        <div className={`px-4 py-3 rounded-xl shadow-2xl backdrop-blur-sm ${
          darkMode 
            ? 'bg-slate-800/90 border border-slate-600' 
            : 'bg-white/90 border border-gray-200'
        }`}>
          <p className={`text-sm font-semibold mb-1 ${darkMode ? 'text-white' : 'text-gray-900'}`}>
            {label}
          </p>
          {payload.map((entry, index) => (
            <p key={index} className="text-sm flex items-center gap-2" style={{ color: entry.color }}>
              <span className="w-2 h-2 rounded-full" style={{ backgroundColor: entry.color }}></span>
              {entry.name}: <span className="font-semibold">{entry.value}</span>
            </p>
          ))}
        </div>
      );
    }
    return null;
  };

  const axisStyle = {
    fontSize: 11,
    fill: darkMode ? '#64748b' : '#9ca3af'
  };

  const gridColor = darkMode ? '#334155' : '#e5e7eb';

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Calls Over Time - with chart-hover */}
      <motion.div 
        className={`rounded-2xl p-6 chart-hover border ${
          darkMode 
            ? 'bg-slate-800/50 border-slate-700/50' 
            : 'bg-white border-gray-200'
        }`}
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <div className="flex items-center space-x-2 mb-6">
          <motion.span 
            className="text-lg"
            animate={{ y: [0, -3, 0] }}
            transition={{ duration: 2, repeat: Infinity }}
          >
            📈
          </motion.span>
          <h3 className={`text-lg font-semibold ${darkMode ? 'text-white' : 'text-gray-900'}`}>
            Calls Over Time (24h)
          </h3>
        </div>
        
        <ResponsiveContainer width="100%" height={250}>
          <AreaChart data={chartData.hourlyData}>
            <defs>
              <linearGradient id="colorCalls" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.4}/>
                <stop offset="95%" stopColor="#3b82f6" stopOpacity={0}/>
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke={gridColor} />
            <XAxis dataKey="time" tick={axisStyle} tickLine={false} axisLine={{ stroke: gridColor }} />
            <YAxis tick={axisStyle} tickLine={false} axisLine={{ stroke: gridColor }} />
            <Tooltip content={<CustomTooltip />} />
            <Area 
              type="monotone" 
              dataKey="calls" 
              stroke="#3b82f6" 
              strokeWidth={2}
              fillOpacity={1} 
              fill="url(#colorCalls)" 
              name="Calls"
            />
          </AreaChart>
        </ResponsiveContainer>
      </motion.div>

      {/* Status Distribution - with chart-hover */}
      <motion.div 
        className={`rounded-2xl p-6 chart-hover border ${
          darkMode 
            ? 'bg-slate-800/50 border-slate-700/50' 
            : 'bg-white border-gray-200'
        }`}
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1 }}
      >
        <div className="flex items-center space-x-2 mb-6">
          <motion.span 
            className="text-lg"
            animate={{ rotate: [0, 10, -10, 0] }}
            transition={{ duration: 3, repeat: Infinity }}
          >
            🥧
          </motion.span>
          <h3 className={`text-lg font-semibold ${darkMode ? 'text-white' : 'text-gray-900'}`}>
            Status Distribution
          </h3>
        </div>
        
        <ResponsiveContainer width="100%" height={250}>
          <PieChart>
            <Pie
              data={chartData.statusData}
              cx="50%"
              cy="50%"
              innerRadius={60}
              outerRadius={90}
              paddingAngle={5}
              dataKey="value"
            >
              {chartData.statusData.map((entry, index) => (
                <Cell 
                  key={`cell-${index}`} 
                  fill={COLORS[index % COLORS.length]}
                  className="transition-all duration-300 hover:opacity-80"
                />
              ))}
            </Pie>
            <Tooltip content={<CustomTooltip />} />
            <Legend 
              formatter={(value) => (
                <span className={darkMode ? 'text-slate-300' : 'text-gray-700'}>{value}</span>
              )}
            />
          </PieChart>
        </ResponsiveContainer>
      </motion.div>

      {/* Duration Distribution - with chart-hover */}
      <motion.div 
        className={`rounded-2xl p-6 lg:col-span-2 chart-hover border ${
          darkMode 
            ? 'bg-slate-800/50 border-slate-700/50' 
            : 'bg-white border-gray-200'
        }`}
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.2 }}
      >
        <div className="flex items-center space-x-2 mb-6">
          <motion.span 
            className="text-lg"
            animate={{ scale: [1, 1.1, 1] }}
            transition={{ duration: 2, repeat: Infinity }}
          >
            📊
          </motion.span>
          <h3 className={`text-lg font-semibold ${darkMode ? 'text-white' : 'text-gray-900'}`}>
            Processing Duration Distribution
          </h3>
        </div>
        
        <ResponsiveContainer width="100%" height={200}>
          <BarChart data={chartData.durationData}>
            <CartesianGrid strokeDasharray="3 3" stroke={gridColor} />
            <XAxis dataKey="range" tick={axisStyle} tickLine={false} axisLine={{ stroke: gridColor }} />
            <YAxis tick={axisStyle} tickLine={false} axisLine={{ stroke: gridColor }} />
            <Tooltip content={<CustomTooltip />} />
            <Bar 
              dataKey="count" 
              fill="#22c55e" 
              radius={[6, 6, 0, 0]}
              name="Calls"
              className="transition-all duration-300"
            />
          </BarChart>
        </ResponsiveContainer>
      </motion.div>
    </div>
  );
}

export default Charts;