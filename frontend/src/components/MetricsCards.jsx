import React, { useEffect, useState } from 'react';
import { motion } from 'framer-motion';

function useAnimatedCounter(value, duration = 1000) {
  const [count, setCount] = useState(0);
  
  useEffect(() => {
    let startTime;
    let animationFrame;
    const startValue = count;
    
    const animate = (timestamp) => {
      if (!startTime) startTime = timestamp;
      const progress = Math.min((timestamp - startTime) / duration, 1);
      setCount(Math.floor(startValue + (value - startValue) * progress));
      
      if (progress < 1) {
        animationFrame = requestAnimationFrame(animate);
      }
    };
    
    animationFrame = requestAnimationFrame(animate);
    return () => cancelAnimationFrame(animationFrame);
  }, [value, duration]);
  
  return count;
}

function AnimatedNumber({ value }) {
  const animatedValue = useAnimatedCounter(value);
  return <>{animatedValue.toLocaleString()}</>;
}

function MetricsCards({ metrics, darkMode = true }) {
  if (!metrics) return null;

  const { realtime, historical, today } = metrics;

  const cards = [
    {
      title: 'Total Calls',
      value: historical?.total_calls || 0,
      subValue: `${today?.api_calls_today || 0} today`,
      icon: '📞',
      color: 'blue',
      glowColor: 'rgba(59, 130, 246, 0.4)'
    },
    {
      title: 'Successful',
      value: historical?.success || 0,
      subValue: `${historical?.today_success || 0} today`,
      icon: '✅',
      color: 'green',
      glowColor: 'rgba(34, 197, 94, 0.4)'
    },
    {
      title: 'Failed',
      value: historical?.failed || 0,
      subValue: `${today?.total_errors || 0} errors`,
      icon: '❌',
      color: 'red',
      glowColor: 'rgba(239, 68, 68, 0.4)'
    },
    {
      title: 'No Speech',
      value: historical?.no_speech || 0,
      subValue: 'Silent audio',
      icon: '🔇',
      color: 'yellow',
      glowColor: 'rgba(234, 179, 8, 0.4)'
    },
    {
      title: 'Unclear',
      value: historical?.unclear_audio || 0,
      subValue: 'Low quality',
      icon: '🔊',
      color: 'orange',
      glowColor: 'rgba(249, 115, 22, 0.4)'
    },
    {
      title: 'Processing',
      value: realtime?.processing || 0,
      subValue: 'In progress',
      icon: '⏳',
      color: 'purple',
      glowColor: 'rgba(168, 85, 247, 0.4)',
      pulse: (realtime?.processing || 0) > 0
    },
    {
      title: 'API Calls',
      value: today?.api_calls_today || 0,
      subValue: `${today?.total_transcriptions || 0} transcriptions`,
      icon: '⚡',
      color: 'cyan',
      glowColor: 'rgba(6, 182, 212, 0.4)'
    },
    {
      title: 'Summaries',
      value: today?.total_summaries || 0,
      subValue: 'Generated',
      icon: '📝',
      color: 'indigo',
      glowColor: 'rgba(99, 102, 241, 0.4)'
    }
  ];

  const textColors = {
    blue: darkMode ? 'text-blue-400' : 'text-blue-600',
    green: darkMode ? 'text-green-400' : 'text-green-600',
    red: darkMode ? 'text-red-400' : 'text-red-600',
    yellow: darkMode ? 'text-yellow-400' : 'text-yellow-600',
    orange: darkMode ? 'text-orange-400' : 'text-orange-600',
    purple: darkMode ? 'text-purple-400' : 'text-purple-600',
    cyan: darkMode ? 'text-cyan-400' : 'text-cyan-600',
    indigo: darkMode ? 'text-indigo-400' : 'text-indigo-600'
  };

  const bgColors = {
    blue: 'rgba(59, 130, 246, 0.1)',
    green: 'rgba(34, 197, 94, 0.1)',
    red: 'rgba(239, 68, 68, 0.1)',
    yellow: 'rgba(234, 179, 8, 0.1)',
    orange: 'rgba(249, 115, 22, 0.1)',
    purple: 'rgba(168, 85, 247, 0.1)',
    cyan: 'rgba(6, 182, 212, 0.1)',
    indigo: 'rgba(99, 102, 241, 0.1)'
  };

  const borderColors = {
    blue: 'rgba(59, 130, 246, 0.3)',
    green: 'rgba(34, 197, 94, 0.3)',
    red: 'rgba(239, 68, 68, 0.3)',
    yellow: 'rgba(234, 179, 8, 0.3)',
    orange: 'rgba(249, 115, 22, 0.3)',
    purple: 'rgba(168, 85, 247, 0.3)',
    cyan: 'rgba(6, 182, 212, 0.3)',
    indigo: 'rgba(99, 102, 241, 0.3)'
  };

  const [hoveredIndex, setHoveredIndex] = useState(null);

  return (
    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
      {cards.map((card, index) => (
        <motion.div
          key={index}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: index * 0.05 }}
          onMouseEnter={() => setHoveredIndex(index)}
          onMouseLeave={() => setHoveredIndex(null)}
          whileTap={{ scale: 0.98 }}
          className={`metric-card glow-${card.color} rounded-2xl p-5 cursor-pointer border ${
            darkMode 
              ? 'bg-slate-800/50 border-slate-700/50' 
              : 'bg-white border-gray-200'
          }`}
          style={{
            background: hoveredIndex === index 
              ? `linear-gradient(135deg, ${bgColors[card.color]}, transparent)` 
              : darkMode ? 'rgba(30, 41, 59, 0.5)' : 'white',
            borderColor: hoveredIndex === index 
              ? borderColors[card.color] 
              : darkMode ? 'rgba(51, 65, 85, 0.5)' : 'rgba(229, 231, 235, 1)',
            boxShadow: hoveredIndex === index 
              ? `0 0 30px ${card.glowColor}, 0 0 60px ${card.glowColor.replace('0.4', '0.2')}, 0 10px 40px rgba(0,0,0,0.3)` 
              : 'none',
            transform: hoveredIndex === index ? 'translateY(-5px)' : 'translateY(0)'
          }}
        >
          <div className="relative z-10">
            <div className="flex items-start justify-between">
              <div>
                <p className={`text-sm font-medium ${darkMode ? 'text-slate-400' : 'text-gray-500'}`}>
                  {card.title}
                </p>
                <p className={`text-3xl font-bold mt-2 ${textColors[card.color]}`}>
                  <AnimatedNumber value={card.value} />
                </p>
                <p className={`text-xs mt-2 ${darkMode ? 'text-slate-500' : 'text-gray-400'}`}>
                  {card.subValue}
                </p>
              </div>
              <motion.div
                className={`text-2xl p-2 rounded-xl ${
                  darkMode ? 'bg-slate-700/50' : 'bg-gray-100'
                }`}
                animate={card.pulse ? { 
                  scale: [1, 1.2, 1],
                  rotate: [0, 5, -5, 0]
                } : {}}
                transition={{ duration: 1.5, repeat: card.pulse ? Infinity : 0 }}
                style={{
                  boxShadow: hoveredIndex === index ? `0 0 15px ${card.glowColor}` : 'none'
                }}
              >
                {card.icon}
              </motion.div>
            </div>
          </div>
        </motion.div>
      ))}
    </div>
  );
}

export default MetricsCards;