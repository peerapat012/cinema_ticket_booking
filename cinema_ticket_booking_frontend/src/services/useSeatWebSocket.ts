import { ref, onMounted, onUnmounted } from 'vue'
import type { Seat } from '@/types'

interface WebSocketMessage {
  type: string
  movieId: string
  seatNo: string
  status: string
  userId?: string
}

export function useSeatWebSocket(movieId: string, userId: string) {
  const socket = ref<WebSocket | null>(null)
  const seats = ref<Map<string, Seat>>(new Map())
  const connected = ref(false)
  const error = ref<string | null>(null)

  const connect = () => {
    const wsUrl = `ws://localhost:8080/ws?movieId=${movieId}&userId=${userId}`
    socket.value = new WebSocket(wsUrl)

    socket.value.onopen = () => {
      connected.value = true
      error.value = null
    }

    socket.value.onmessage = (event) => {
      try {
        const data: Seat = JSON.parse(event.data)
        seats.value.set(data.seatNo, data)
      } catch (e) {
        console.error('Failed to parse seat update:', e)
      }
    }

    socket.value.onerror = () => {
      error.value = 'WebSocket connection error'
    }

    socket.value.onclose = () => {
      connected.value = false
    }
  }

  const sendMessage = (type: string, seatNo: string) => {
    if (socket.value && connected.value) {
      const message: WebSocketMessage = {
        type,
        movieId,
        seatNo,
        status: '',
        userId,
      }
      socket.value.send(JSON.stringify(message))
    }
  }

  const lockSeat = (seatNo: string) => sendMessage('lock_seat', seatNo)
  const bookSeat = (seatNo: string) => sendMessage('book_seat', seatNo)
  const releaseSeat = (seatNo: string) => sendMessage('release_seat', seatNo)

  const disconnect = () => {
    if (socket.value) {
      socket.value.close()
      socket.value = null
    }
  }

  onMounted(() => {
    if (movieId && userId) {
      connect()
    }
  })

  onUnmounted(() => {
    disconnect()
  })

  return {
    seats,
    connected,
    error,
    lockSeat,
    bookSeat,
    releaseSeat,
    connect,
    disconnect,
  }
}
