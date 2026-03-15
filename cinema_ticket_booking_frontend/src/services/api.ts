import type {
  Movie,
  TokenResponse,
  Booking,
  SeatLockResponse,
  CreateBookingRequest,
  User,
  Payment,
  CreatePaymentRequest,
} from '@/types'

const API_BASE = 'http://localhost:8080'

class ApiService {
  private token: string | null = localStorage.getItem('token')

  constructor() {
    if (this.token) {
      console.log('API initialized with token from localStorage')
    }
  }

  setToken(token: string | null) {
    this.token = token
    console.log('Token set in API:', token ? 'yes' : 'no')
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    }

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`
    }

    console.log(`API Request: ${options.method || 'GET'} ${endpoint}`, options.body)

    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers,
    })

    console.log(`API Response: ${response.status}`, response.statusText)

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: 'Request failed' }))
      console.error('API Error:', errorData)
      throw new Error(errorData.error || `Request failed: ${response.status}`)
    }

    const data = await response.json()
    console.log('API Data:', data)
    return data.data ?? data
  }

  async loginWithGoogle(idToken: string): Promise<TokenResponse> {
    const data = await this.request<TokenResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ idToken }),
    })
    this.token = data.token
    return data
  }

  async getMovies(): Promise<Movie[]> {
    const response = await this.request<{ data: Movie[] }>('/movies')
    return response.data || response
  }

  async getBookedSeats(movieId: string): Promise<string[]> {
    const response = await this.request<{ bookedSeats: string[] }>(`/movies/${movieId}/seats`)
    return response.bookedSeats || []
  }

  async getCurrentUser(): Promise<User> {
    return this.request<User>('/users/me')
  }

  async lockSeats(movieId: string, seats: string[]): Promise<SeatLockResponse> {
    return this.request<SeatLockResponse>('/seats/lock', {
      method: 'POST',
      body: JSON.stringify({ movieId, seats }),
    })
  }

  async unlockSeats(movieId: string, seats: string[]): Promise<{ message: string }> {
    return this.request<{ message: string }>('/seats/unlock', {
      method: 'POST',
      body: JSON.stringify({ movieId, seats }),
    })
  }

  async createBooking(request: CreateBookingRequest): Promise<Booking> {
    return this.request<Booking>('/bookings', {
      method: 'POST',
      body: JSON.stringify(request),
    })
  }

  async getMyBookings(): Promise<Booking[]> {
    const response = await this.request<{ data: Booking[] }>('/bookings/me')
    return response.data || response
  }

  async getAllBookings(query: string = ''): Promise<Booking[]> {
    const response = await this.request<{ data: Booking[] }>(`/bookings${query}`)
    return response.data || response
  }

  async getAllUsers(): Promise<User[]> {
    const response = await this.request<{ data: User[] }>('/users')
    return response.data || response
  }

  async createPayment(
    request: CreatePaymentRequest,
  ): Promise<{ payment: Payment; booking: Booking }> {
    return this.request<{ payment: Payment; booking: Booking }>('/payments', {
      method: 'POST',
      body: JSON.stringify(request),
    })
  }

  async getMyPayments(): Promise<Payment[]> {
    const response = await this.request<{ data: Payment[] }>('/payments/me')
    return response.data || response
  }
}

export const api = new ApiService()
export default api
