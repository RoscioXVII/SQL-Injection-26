<script>
import { deleteMessage } from "@/services/axios";

export default {
	name: 'DeleteMessage',
	props: {
		show: Boolean,
		messageId: Number,
		chatId: Number,
		type: String
	},
	emits: ['close'],

	data() {
		return {
			userId: Number(sessionStorage.getItem('userId'))
		}
	},

	methods: {
		doDelete() {
			deleteMessage(this.userId, this.chatId, this.messageId, this.type)
			this.$emit('close')
		}
	}
}
</script>

<template>
  <div v-if="show" class="overlay">
    <div class="action-box">
      <h2 class="text-center">Are you sure?</h2>
      <p class="text-center">This action will be irreversible.</p>

      <div class="actions">
        <button class="btn btn-secondary" @click="$emit('close')">Cancel</button>
        <button class="btn btn-danger" @click="doDelete">Delete</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.overlay {
	position: fixed;
	inset: 0;
	background: rgba(0,0,0,0.5);
	display: flex;
	justify-content: center;
	align-items: center;
	z-index: 1200;
}

.action-box {
	background: white;
	padding: 20px;
	width: 400px;
	max-height: 80vh;
	overflow-y: auto;
	border-radius: 8px;
	box-shadow: 0 4px 15px rgba(0,0,0,0.3);
	z-index: 1201;
	display: flex;
	flex-direction: column;
}

.actions {
	display: flex;
	gap: 10px;
	margin-top: 10px;
	justify-content: space-between;
}
</style>
