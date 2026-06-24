import { useEffect, useRef, useState, useCallback } from 'react'
import io from 'socket.io-client'

export const useWebSocket = () => {
  const socketRef = useRef(null)
  const [isConnected, setIsConnected] = useState(false)
  const [logs, setLogs] = useState([])

  const connect = useCallback(() => {
    if (socketRef.current?.readyState === WebSocket.OPEN) return

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/ws`

    const ws = new WebSocket(wsUrl)
    socketRef.current = ws

    ws.onopen = () => {
      console.log('WebSocket connected')
      setIsConnected(true)
      ws.send(JSON.stringify({ type: 'request_logs' }))
    }

    ws.onclose = () => {
      console.log('WebSocket disconnected')
      setIsConnected(false)
      setTimeout(connect, 1000)
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'log_history') {
          setLogs(Array.isArray(msg.data) ? msg.data : [])
        } else if (msg.type === 'log_message') {
          setLogs(prev => [...prev.slice(-499), msg.data])
        }
      } catch (e) {
        console.error('WebSocket message error:', e)
      }
    }

    ws.onerror = (err) => {
      console.error('WebSocket error:', err)
      setIsConnected(false)
    }
  }, [])

  const disconnect = useCallback(() => {
    socketRef.current?.close()
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
