<script>
import { commentMessage, getComments, unComment } from "../services/axios";

export default {
	name: 'Comment',
	props: {
		userId: Number,
		show: Boolean,
		messageId: Number,
		chatId: Number,
		type: String
	},
	emits: ['close'],

	data() {
		return {
			emojiInput: '',
			comments: [],
			pollingInterval: null
		}
	},

	computed: {
		// nessun computed extra nel tuo setup, ma puoi aggiungere se vuoi
	},

	watch: {
		messageId(newId) {
			if (newId) this.loadComments()
		}
	},

	mounted() {
		if (this.messageId) this.loadComments()
		this.startPolling()
	},

	beforeUnmount() {
		this.stopPolling()
	},

	methods: {
		filterEmoji() {
			this.emojiInput = Array.from(this.emojiInput)
				.filter(char => /\p{Emoji}/u.test(char))
				.join('')
		},

		confirmComment() {
			if (this.emojiInput.trim()) {
				commentMessage(this.userId, this.chatId, this.messageId, this.type, this.emojiInput)
				this.emojiInput = ''
			}
		},

		insertEmoji(emoji) {
			this.emojiInput = emoji
			this.filterEmoji()
		},

		deleteEmoji(commentId) {
			unComment(this.userId, this.chatId, this.messageId, commentId, this.type)
		},

		async loadComments() {
			if (!this.userId || !this.chatId || !this.messageId) return
			const comm = await getComments(this.userId, this.chatId, this.messageId, this.type) || []
			this.comments = [...comm]
		},

		startPolling() {
			this.stopPolling()
			this.pollingInterval = setInterval(() => {
				if (this.show) {
					this.loadComments()
				}
			}, 2000)
		},

		stopPolling() {
			if (this.pollingInterval) {
				clearInterval(this.pollingInterval)
				this.pollingInterval = null
			}
		}
	}
}
</script>

<template>
  <div v-if="show" class="overlay">
    <div class="action-box">
      <h2 class="text-center">Comments</h2>
      <div v-if="comments.length > 0 " class="comment-box">
        <div v-for="comment in comments" :key="comment.id" class="d-flex justify-content-between align-items-center">
          <div class="pb-0 m-0">
            <span v-if="comment.sender.userId !== userId" class="emoji">{{ comment.sender.username }}</span>
            <span v-else class="emoji">You</span>
          </div>
          <div class="d-flex align-items-center gap-2">
            <button v-if="comment.sender.userId === userId" class="btn btn-danger" @click="deleteEmoji(comment.commentId)">
              <img src="../icons/trash3-fill.svg" alt="Delete" width="16" height="16" class="d-flex align-items-center">
            </button>
            <span class="emoji">{{ comment.emoji }}</span>
          </div>
        </div>
      </div>
      <input
        v-model="emojiInput"
        type="text"
        placeholder="Comment with an emoji..."
        @input="filterEmoji"
      >
      <div class="d-flex justify-content-center">
        <button class="btn" @click="insertEmoji(`🕶`)">🕶</button>
        <button class="btn" @click="insertEmoji(`✨`)">✨</button>
        <button class="btn" @click="insertEmoji(`😍`)">😍</button>
        <button class="btn" @click="insertEmoji(`💕`)">💕</button>
      </div>
      <div class="actions">
        <button class="btn btn-secondary" @click="$emit('close')">Cancel</button>
        <button class="btn btn-primary" @click="confirmComment">Comment</button>
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
	justify-content: space-between;
	gap: 10px;
	margin-top: 10px;
}

.comment-box {
	background: rgba(0, 0, 0, 0.11);
	padding: 20px;
	overflow-y: auto;
	display: flex;
	flex-direction: column;
	border-top-right-radius: 10px;
	border-top-left-radius: 10px;
}

.emoji {
	font-size: 1.2rem;
}
</style>
