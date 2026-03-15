export interface Movie {
  id: string
  title: string
  description?: string
  durationMinutes: number
  posterUrl?: string
  price: number
  status: string
  createdAt: string
  updatedAt: string
}

export interface User {
  id: string
  username: string
  email: string
  role: string
  picture?: string
}

export interface TokenResponse {
  token: string
  user: User
}

export interface BookingSeat {
  seatNo: string
  price: number
}

export interface Booking {
  id: string
  bookingCode: string
  userId: string
  movieId: string
  seats: BookingSeat[]
  totalPrice: number
  status: 'pending' | 'confirmed' | 'cancelled' | 'expired'
  paymentStatus: 'pending' | 'completed' | 'failed'
  paymentDeadline?: string
  bookedAt: string
  createdAt: string
  updatedAt: string
  _id?: string
}

export interface Payment {
  id: string
  userId: string
  movieId: string
  movieTitle: string
  seats: BookingSeat[]
  amount: number
  status: 'pending' | 'completed' | 'failed'
  transactionId: string
  createdAt: string
  updatedAt: string
}

export interface CreatePaymentRequest {
  bookingId: string
}

export interface Seat {
  seatNo: string
  status: 'available' | 'locked' | 'booked'
  userId?: string
}

export interface Showtime {
  id: string
  movieId: string
  startTime: string
  endTime: string
  price: number
  status: string
}

export interface SeatLockRequest {
  movieId: string
  seats: string[]
}

export interface SeatLockResponse {
  message: string
  lockedSeats: string[]
  failedSeats?: string[]
}

export interface CreateBookingRequest {
  movieId: string
  seats: BookingSeat[]
}
