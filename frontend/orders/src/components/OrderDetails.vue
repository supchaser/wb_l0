<template>
  <div class="order-details">
    <div class="order-header">
      <h2>Order Details</h2>
      <div class="order-id">UID: {{ order.order_uid }}</div>
    </div>

    <div class="sections-grid">
      <!-- Main Info -->
      <div class="section">
        <h3>Main Information</h3>
        <div class="info-grid">
          <div class="info-item">
            <span class="label">Track Number:</span>
            <span class="value">{{ order.track_number }}</span>
          </div>
          <div class="info-item">
            <span class="label">Entry:</span>
            <span class="value">{{ order.entry }}</span>
          </div>
          <div class="info-item">
            <span class="label">Locale:</span>
            <span class="value">{{ order.locale }}</span>
          </div>
          <div class="info-item">
            <span class="label">Customer ID:</span>
            <span class="value">{{ order.customer_id }}</span>
          </div>
          <div class="info-item">
            <span class="label">Delivery Service:</span>
            <span class="value">{{ order.delivery_service }}</span>
          </div>
          <div class="info-item">
            <span class="label">Date Created:</span>
            <span class="value">{{ formatDate(order.date_created) }}</span>
          </div>
        </div>
      </div>

      <!-- Delivery -->
      <div class="section">
        <h3>Delivery Information</h3>
        <div class="info-grid">
          <div class="info-item">
            <span class="label">Name:</span>
            <span class="value">{{ order.delivery.name }}</span>
          </div>
          <div class="info-item">
            <span class="label">Phone:</span>
            <span class="value">{{ order.delivery.phone }}</span>
          </div>
          <div class="info-item">
            <span class="label">Email:</span>
            <span class="value">{{ order.delivery.email }}</span>
          </div>
          <div class="info-item">
            <span class="label">Address:</span>
            <span class="value">{{ order.delivery.address }}, {{ order.delivery.city }}</span>
          </div>
          <div class="info-item">
            <span class="label">Region:</span>
            <span class="value">{{ order.delivery.region }}</span>
          </div>
          <div class="info-item">
            <span class="label">ZIP:</span>
            <span class="value">{{ order.delivery.zip }}</span>
          </div>
        </div>
      </div>

      <!-- Payment -->
      <div class="section">
        <h3>Payment Information</h3>
        <div class="info-grid">
          <div class="info-item">
            <span class="label">Transaction:</span>
            <span class="value">{{ order.payment.transaction }}</span>
          </div>
          <div class="info-item">
            <span class="label">Amount:</span>
            <span class="value">${{ order.payment.amount }}</span>
          </div>
          <div class="info-item">
            <span class="label">Currency:</span>
            <span class="value">{{ order.payment.currency }}</span>
          </div>
          <div class="info-item">
            <span class="label">Provider:</span>
            <span class="value">{{ order.payment.provider }}</span>
          </div>
          <div class="info-item">
            <span class="label">Bank:</span>
            <span class="value">{{ order.payment.bank }}</span>
          </div>
          <div class="info-item">
            <span class="label">Payment Date:</span>
            <span class="value">{{ formatTimestamp(order.payment.payment_dt) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Items -->
    <div class="section">
      <h3>Order Items ({{ order.items.length }})</h3>
      <div class="items-grid">
        <div v-for="(item, index) in order.items" :key="index" class="item-card">
          <div class="item-header">
            <h4>{{ item.name }}</h4>
            <span class="brand">{{ item.brand }}</span>
          </div>
          <div class="item-details">
            <div class="item-info">
              <span class="label">Price:</span>
              <span class="value">${{ item.price }}</span>
            </div>
            <div class="item-info">
              <span class="label">Total Price:</span>
              <span class="value">${{ item.total_price }}</span>
            </div>
            <div class="item-info">
              <span class="label">Sale:</span>
              <span class="value">{{ item.sale }}%</span>
            </div>
            <div class="item-info">
              <span class="label">Size:</span>
              <span class="value">{{ item.size }}</span>
            </div>
            <div class="item-info">
              <span class="label">Status:</span>
              <span class="value">{{ item.status }}</span>
            </div>
            <div class="item-info">
              <span class="label">NM ID:</span>
              <span class="value">{{ item.nm_id }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'OrderDetails',
  props: {
    order: {
      type: Object,
      required: true
    }
  },
  methods: {
    formatDate(dateString) {
      return new Date(dateString).toLocaleString()
    },
    formatTimestamp(timestamp) {
      return new Date(timestamp * 1000).toLocaleString()
    }
  }
}
</script>

<style scoped>
.order-details {
  margin-top: 20px;
}

.order-header {
  text-align: center;
  margin-bottom: 30px;
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border-radius: 10px;
}

.order-header h2 {
  margin-bottom: 10px;
}

.order-id {
  font-size: 14px;
  opacity: 0.9;
}

.sections-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.section {
  background: #f8f9fa;
  padding: 20px;
  border-radius: 8px;
  border-left: 4px solid #3498db;
}

.section h3 {
  margin-bottom: 15px;
  color: #2c3e50;
  border-bottom: 2px solid #eee;
  padding-bottom: 10px;
}

.info-grid {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #eee;
}

.info-item:last-child {
  border-bottom: none;
}

.label {
  font-weight: bold;
  color: #555;
  min-width: 120px;
}

.value {
  color: #333;
  text-align: right;
  word-break: break-word;
}

.items-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 15px;
  margin-top: 15px;
}

.item-card {
  background: white;
  padding: 15px;
  border-radius: 8px;
  border: 1px solid #ddd;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.item-header {
  margin-bottom: 10px;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
}

.item-header h4 {
  margin-bottom: 5px;
  color: #2c3e50;
}

.brand {
  font-size: 12px;
  color: #666;
  font-style: italic;
}

.item-details {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.item-info {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
}

@media (max-width: 768px) {
  .sections-grid {
    grid-template-columns: 1fr;
  }
  
  .items-grid {
    grid-template-columns: 1fr;
  }
  
  .info-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 5px;
  }
  
  .value {
    text-align: left;
  }
}
</style>
