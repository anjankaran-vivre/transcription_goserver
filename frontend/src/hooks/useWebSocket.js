import { useEffect, useRef, useState, useCallback } from 'react'
import { io } from 'socket.io-client'

export const useWebSocket = () => {
  const socketRef = useRef(null)
  const [isConnected, setIsConnected] = useState(false)
  const [logs, setLogs] = useState([])

  const connect = useCallback(() => {
    if (socketRef.current?.connected) return

    socketRef.current = io(window.location.origin, {
      path: '/socket.io',
      transports: ['polling', 'websocket'],
      reconnection: true,
      reconnectionAttempts: 10,
      reconnectionDelay: 1000,
      timeout: 20000,
    })

    socketRef.current.on('connect', () => {
      console.log('WebSocket connected')
      setIsConnected(true)
      socketRef.current.emit('request_logs')
    })

    socketRef.current.on('disconnect', () => {
      console.log('WebSocket disconnected')
      setIsConnected(false)
    })

    socketRef.current.on('log_message', (log) => {
      setLogs(prev => [...prev.slice(-499), log])
    })

    // ✅ FIXED HERE
    socketRef.current.on('log_history', (history) => {
      setLogs(Array.isArray(history) ? history : [])
    })

    socketRef.current.on('connect_error', (error) => {
      console.error('WebSocket error:', error)
      setIsConnected(false)
    })
  }, [])

  const disconnect = useCallback(() => {
    socketRef.current?.disconnect()
    socketRef.current = null
  }, [])

  const clearLogs = useCallback(() => {
    setLogs([])
  }, [])

  useEffect(() => {
    connect()
    return () => disconnect()
  }, [connect, disconnect])

  return { isConnected, logs, clearLogs }
}
