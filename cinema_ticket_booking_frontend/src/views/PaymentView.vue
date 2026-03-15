<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/services/api'
import type { Movie, Booking } from '@/types'

const route = useRoute()
const router = useRouter()

const bookingId = route.query.bookingId as string
const movieId = route.query.movieId as string
const seatsParam = route.query.seats as string
const totalPriceParam = route.query.price as string

const movie = ref<Movie | null>(null)
const loading = ref(true)
const processing = ref(false)
const error = ref('')
const success = ref(false)
const booking = ref<Booking | null>(null)

const selectedSeats = computed(() => (seatsParam ? seatsParam.split(',') : []))
const totalPrice = computed(() => parseFloat(totalPriceParam) || 0)

const timer = ref(0)
const timerInterval = ref<number | null>(null)

const formattedTime = computed(() => {
  const minutes = Math.floor(timer.value / 60)
  const seconds = timer.value % 60
  return `${minutes}:${seconds.toString().padStart(2, '0')}`
})

onMounted(async () => {
  if (!bookingId || !movieId) {
    router.push('/')
    return
  }

  try {
    const [movies, bookings] = await Promise.all([api.getMovies(), api.getMyBookings()])

    movie.value = movies.find((m) => m.id === movieId) || null
    booking.value = bookings.find((b) => (b.id || (b as any)._id) === bookingId) || null

    if (booking.value?.paymentDeadline) {
      startTimer(booking.value.paymentDeadline)
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to load'
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  if (timerInterval.value) {
    clearInterval(timerInterval.value)
  }
})

const startTimer = (deadline: string) => {
  const updateTimer = () => {
    const now = new Date().getTime()
    const deadlineTime = new Date(deadline).getTime()
    const remaining = Math.floor((deadlineTime - now) / 1000)

    if (remaining <= 0) {
      timer.value = 0
      if (timerInterval.value) {
        clearInterval(timerInterval.value)
      }
      error.value = 'Payment time has expired. Please try again.'
      return
    }

    timer.value = remaining
  }

  updateTimer()
  timerInterval.value = window.setInterval(updateTimer, 1000)
}

const handlePayment = async () => {
  processing.value = true
  error.value = ''

  try {
    console.log('Creating payment for booking:', bookingId)

    const result = await api.createPayment({
      bookingId,
    })

    console.log('Payment result:', result)
    booking.value = result.booking
    success.value = true

    if (timerInterval.value) {
      clearInterval(timerInterval.value)
    }
  } catch (e) {
    console.error('Payment error:', e)
    error.value = e instanceof Error ? e.message : 'Payment failed'
  } finally {
    processing.value = false
  }
}

const goHome = () => {
  router.push('/')
}

const goToBookings = () => {
  router.push('/bookings')
}
</script>

<template>
  <div class="payment-page">
    <div v-if="loading" class="loading">Loading...</div>

    <div v-else-if="success" class="success-container">
      <div class="success-icon">✓</div>
      <h1>Payment Successful!</h1>
      <p v-if="booking">Booking Code: {{ booking.bookingCode }}</p>
      <div class="booking-details">
        <p>Movie: {{ movie?.title }}</p>
        <p>Seats: {{ selectedSeats.join(', ') }}</p>
        <p>Total: ฿{{ totalPrice }}</p>
      </div>
      <div class="success-buttons">
        <button class="btn-home" @click="goHome">Back to Movies</button>
        <button class="btn-bookings" @click="goToBookings">View My Bookings</button>
      </div>
    </div>

    <div v-else class="payment-container">
      <h1>Payment</h1>

      <div v-if="timer > 0" class="timer-warning">
        <p>
          Time remaining to complete payment: <strong>{{ formattedTime }}</strong>
        </p>
      </div>

      <div class="order-summary">
        <h2>Order Summary</h2>
        <div class="summary-item">
          <span>Movie:</span>
          <span>{{ movie?.title }}</span>
        </div>
        <div class="summary-item">
          <span>Seats:</span>
          <span>{{ selectedSeats.join(', ') }}</span>
        </div>
        <div class="summary-item total">
          <span>Total:</span>
          <span>฿{{ totalPrice }}</span>
        </div>
      </div>

      <div class="payment-form">
        <h2>Payment Details</h2>
        <div class="form-group">
          <label>Card Number</label>
          <input type="text" placeholder="1234 5678 9012 3456" />
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>Expiry Date</label>
            <input type="text" placeholder="MM/YY" />
          </div>
          <div class="form-group">
            <label>CVV</label>
            <input type="text" placeholder="123" />
          </div>
        </div>

        <div v-if="error" class="error">{{ error }}</div>

        <button class="btn-pay" :disabled="processing" @click="handlePayment">
          {{ processing ? 'Processing...' : `Pay ฿${totalPrice}` }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.payment-page {
  padding: 2rem;
  max-width: 600px;
  margin: 0 auto;
}

.loading {
  text-align: center;
  padding: 2rem;
}

.success-container {
  text-align: center;
  background: white;
  padding: 3rem;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.success-icon {
  width: 80px;
  height: 80px;
  background: #42b883;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 3rem;
  margin: 0 auto 1rem;
}

.success-container h1 {
  color: #42b883;
  margin-bottom: 1rem;
}

.booking-details {
  background: #f8f9fa;
  padding: 1rem;
  border-radius: 4px;
  margin: 1rem 0;
}

.booking-details p {
  margin: 0.5rem 0;
  color: #333;
}

.btn-home {
  padding: 0.75rem 2rem;
  background: #42b883;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
}

.btn-bookings {
  padding: 0.75rem 2rem;
  background: #3498db;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
  margin-left: 1rem;
}

.success-buttons {
  margin-top: 1.5rem;
}

.payment-container {
  background: white;
  padding: 2rem;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.payment-container h1 {
  color: #2c3e50;
  margin: 0 0 1.5rem;
}

.timer-warning {
  background: #fff3cd;
  border: 1px solid #ffc107;
  padding: 1rem;
  border-radius: 4px;
  margin-bottom: 1.5rem;
  text-align: center;
}

.timer-warning p {
  margin: 0;
  color: #856404;
}

.timer-warning strong {
  color: #d32f2f;
}

.order-summary {
  background: #f8f9fa;
  padding: 1.5rem;
  border-radius: 8px;
  margin-bottom: 2rem;
}

.order-summary h2 {
  margin: 0 0 1rem;
  color: #2c3e50;
  font-size: 1.2rem;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem 0;
  color: #666;
}

.summary-item.total {
  border-top: 1px solid #ddd;
  margin-top: 0.5rem;
  padding-top: 1rem;
  font-weight: bold;
  color: #2c3e50;
  font-size: 1.2rem;
}

.payment-form h2 {
  margin: 0 0 1rem;
  color: #2c3e50;
  font-size: 1.2rem;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: #333;
}

.form-group input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 1rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.error {
  color: #e74c3c;
  margin: 1rem 0;
}

.btn-pay {
  width: 100%;
  padding: 1rem;
  background: #42b883;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 1.1rem;
  cursor: pointer;
  margin-top: 1rem;
}

.btn-pay:hover:not(:disabled) {
  background: #3aa876;
}

.btn-pay:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
