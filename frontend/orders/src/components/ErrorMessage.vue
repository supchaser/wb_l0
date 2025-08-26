<template>
  <div v-if="message" class="error-message">
    <div class="error-icon">⚠️</div>
    <div class="error-text">{{ message }}</div>
  </div>
</template>

<script>
export default {
  name: 'ErrorMessage',
  props: {
    message: String
  }
}
</script>

<style scoped>
.error-message {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 15px;
  background-color: #ffe6e6;
  border: 1px solid #ffcccc;
  border-radius: 6px;
  color: #d8000c;
  margin: 20px 0;
}

.error-icon {
  font-size: 20px;
}

.error-text {
  font-weight: 500;
}
</style>
