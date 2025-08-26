<template>
  <div id="app">
    <div class="container">
      <h1>Order Information System</h1>
      <OrderSearch @search-order="handleSearch" />
      <OrderDetails v-if="order" :order="order" />
      <ErrorMessage v-if="error" :message="error" />
      <div v-if="searchPerformed && !loading && !order && !error" class="no-data">
        Order not found or error occurred
      </div>
      <div v-if="loading" class="loading">Loading...</div>
    </div>
  </div>
</template>

<script>
import OrderSearch from './components/OrderSearch.vue'
import OrderDetails from './components/OrderDetails.vue'
import ErrorMessage from './components/ErrorMessage.vue'
import axios from 'axios'

export default {
  name: 'App',
  components: {
    OrderSearch,
    OrderDetails,
    ErrorMessage
  },
  data() {
    return {
      order: null,
      loading: false,
      searchPerformed: false,
      error: null
    }
  },
  methods: {
    async handleSearch(orderUid) {
      this.loading = true
      this.searchPerformed = true
      this.order = null
      this.error = null
      
      try {
        const response = await axios.get(`http://localhost:8080/api/v1/orders/${orderUid}`)
        if (response.data && response.data.order_uid) {
          this.order = response.data
        } else {
          this.error = "Order not found"
        }
      } catch (error) {
        console.error('Error fetching order:', error)
        if (error.response && error.response.status === 404) {
          this.error = "Order not found"
        } else {
          this.error = "Error fetching order data"
        }
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  color: #333;
  min-height: 100vh;
}

#app {
  min-height: 100vh;
  padding: 20px;
  display: flex;
  justify-content: center;
  align-items: flex-start;
}

.container {
  max-width: 1200px;
  width: 100%;
  background: white;
  padding: 30px;
  border-radius: 10px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  margin-top: 20px;
}

h1 {
  text-align: center;
  margin-bottom: 30px;
  color: #2c3e50;
}

.loading {
  text-align: center;
  padding: 20px;
  font-size: 18px;
  color: #666;
}

.no-data {
  text-align: center;
  padding: 20px;
  font-size: 16px;
  color: #e74c3c;
}
</style>
