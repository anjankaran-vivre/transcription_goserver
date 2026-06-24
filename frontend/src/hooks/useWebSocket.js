import { useEffect, useRef, useState, useCallback } from 'react'

export const useWebSocket = (apiUrl = '') => {
  const socketRef = useRef(null)
  const [isConnected, setIsConnected] = useState(false)
  const [logs, setLogs] = useState([])

  const connect = useCallback(() => {
    if (socketRef.current?.connected) return

    // Determine WebSocket URL (same port as API, /ws path)
    let wsUrl
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
      // Development: use localhost:5050 with /ws path
      wsUrl = 'ws://localhost:5050/ws'
    } else {
      // Production: use same domain and protocol on same port
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      wsUrl = `${protocol}//${window.location.hostname}${window.location.port ? ':' + window.location.port : ''}/ws`
    }

    console.log('🔗 Connecting to WebSocket:', wsUrl)

    const socket = new WebSocket(wsUrl)
    socketRef.current = socket

    socket.onopen = () => {
      console.log('✓ WebSocket connected')
      setIsConnected(true)
      socket.send(JSON.stringify({ type: 'request_logs' }))
    }

    socket.onclose = () => {
      console.log('⚠ WebSocket disconnected, reconnecting in 3s...')
      setIsConnected(false)
      setTimeout(connect, 3000)
    }

    socket.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'log_history') {
          setLogs(Array.isArray(msg.data) ? msg.data : [])
        } else if (msg.type === 'log_message') {
          setLogs(prev => [...prev.slice(-499), msg.data])
        }
      } catch (e) {
        console.error('✗ WebSocket message error:', e)
      }
    }

    socket.onerror = (err) => {
      console.error('✗ WebSocket error:', err)
      setIsConnected(false)
    }
  }, [])

  const disconnect = useCallback(() => {
    if (socketRef.current) {
      socketRef.current.close()
      socketRef.current = null
    }
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
