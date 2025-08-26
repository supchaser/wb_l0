<template>
  <div class="order-search">
    <form @submit.prevent="handleSubmit" class="search-form">
      <input
        type="text"
        v-model="orderUid"
        placeholder="Enter Order UID"
        required
        class="search-input"
      />
      <button type="submit" class="search-button">Search Order</button>
    </form>
  </div>
</template>

<script>
export default {
  name: 'OrderSearch',
  data() {
    return {
      orderUid: ''
    }
  },
  methods: {
    isValidOrderUid(uid) {
        return uid && uid.length >= 10 && /^[a-zA-Z0-9-]+$/.test(uid)
    },
    handleSubmit() {
        const trimmedUid = this.orderUid.trim()
        if (this.isValidOrderUid(trimmedUid)) {
        this.$emit('search-order', trimmedUid)
        this.orderUid = ''
        } else {
        this.$emit('invalid-input', 'Please enter a valid Order UID')
        }
    }
  }
}
</script>

<style scoped>
.order-search {
  margin-bottom: 30px;
}

.search-form {
  display: flex;
  gap: 10px;
  justify-content: center;
  align-items: center;
}

.search-input {
  padding: 12px 16px;
  border: 2px solid #ddd;
  border-radius: 6px;
  font-size: 16px;
  min-width: 300px;
  transition: border-color 0.3s;
}

.search-input:focus {
  outline: none;
  border-color: #3498db;
}

.search-button {
  padding: 12px 24px;
  background-color: #3498db;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  cursor: pointer;
  transition: background-color 0.3s;
}

.search-button:hover {
  background-color: #2980b9;
}

.search-button:active {
  transform: translateY(1px);
}
</style>
