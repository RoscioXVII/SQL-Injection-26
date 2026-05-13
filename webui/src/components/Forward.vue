<script>

import { getConversations, getGroups, forwardMessage,getUsers ,createConversation} from '@/services/axios'
import ErrorMsg from "@/components/ErrorMsg.vue";

export default {
	name: 'ForwardMessage',
	components: {ErrorMsg},
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
			chats: [],
			selected: new Set(),
			errormsg: null,
			errorTimeout: null,
		}
	},

	watch: {
		show(newVal) {
			if (newVal) {
				this.selected.clear()
			}
		}
	},

	mounted() {
		this.loadChats()
	},

	methods: {
		showError(msg) {
			this.errormsg = msg;

			// cancella eventuale timeout precedente
			if (this.errorTimeout) clearTimeout(this.errorTimeout);

			// setta il nuovo timeout
			this.errorTimeout = setTimeout(() => {
				this.errormsg = null;
				this.errorTimeout = null;
			}, 5000);
		},
		async loadChats() {
			const conversations = await getConversations(this.userId) || []
			const groups = await getGroups(this.userId) || []
			const usersList = await getUsers() || [];
			const conversationNames = new Set(conversations.map(c => c.name));
			const users = usersList
				.filter(u => !conversationNames.has(u.username) && u.username !== sessionStorage.getItem('username'))
				.map(u => ({ username: u.username , name: u.username})) || []; //prendo gli utenti di cui non ho già una conversazione
			this.chats = [...conversations, ...groups, ...users];
		},

		getChatId(chat) {
			if (chat.conversationId) return `c-${chat.conversationId}`
			if (chat.groupId) return `g-${chat.groupId}`
			if (chat.username) return `u-${chat.username}`
		},

		toggle(chat) {
			const id = this.getChatId(chat)
			this.selected.has(id)
				? this.selected.delete(id)
				: this.selected.add(id)
		},

		async confirmForward() {
			const conversations = [];
			const groups = [];

			for (const chat of this.chats) {
				const id = this.getChatId(chat);
				if (this.selected.has(id)) {

					if (chat.conversationId) { // se c'è già una conversazione, la metto direttamente
						conversations.push(Number(chat.conversationId));
					} else if (chat.username) { // altrimenti creo la conversazione e poi la metto
						try {
							// crea conversazione diretta con quell'utente
							const newConv = await createConversation(this.userId, chat.username);
							conversations.push(Number(newConv.conversationId));
						} catch (err) {
							console.error("Errore creando conversazione:", err);
							this.showError(`Non posso creare conversazione con ${chat.username}`);
						}
					} else if (chat.groupId) { // i gruppi li metto direttamente
						groups.push(Number(chat.groupId));
					}
				}
			}

			try {
				await forwardMessage(
					this.userId,
					this.chatId,
					this.messageId,
					this.type,
					conversations,
					groups
				);
			} catch (err) {
				console.error("Errore forward:", err);
				this.showError("Non è stato possibile inoltrare il messaggio");
				return;
			}

			this.$emit('close');
		},
	}}
</script>

<template>
  <div v-if="show" class="overlay">
    <div class="action-box">
      <ErrorMsg v-if="errormsg" :msg="errormsg" />
      <h4 class="text-center">Forward message</h4>
      <div
        v-for="chat in chats"
        :key="getChatId(chat)"
        class="chat-item"
        @click="toggle(chat)"
      >
        <input
          type="checkbox"
          :checked="selected.has(getChatId(chat))"
          class="selected"
        >
        <span class="ms-2">{{ chat.name }}</span>
      </div>

      <div class="actions">
        <button class="btn btn-secondary" @click="$emit('close')">Cancel</button>
        <button class="btn btn-primary" @click="confirmForward">Forward</button>
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

.chat-item {
	cursor: pointer;
	padding: 8px 4px;
	display: flex;
	align-items: center;
}

.chat-item:hover {
	background-color: #f0f0f0;
}

.actions {
	display: flex;
	justify-content: space-between;
	gap: 10px;
	margin-top: 10px;
}

.selected {
	appearance: none;
	-webkit-appearance: none;
	width: 18px;
	height: 18px;
	border: 2px solid #0d6efd;
	border-radius: 4px;
	cursor: pointer;
	display: flex;
	align-items: center;
	justify-content: center;
	transition: all 0.2s ease;
	background-color: white;
}

.selected:checked {
	background-color: #0d6efd;
	border-color: #0d6efd;
}

.selected:checked::after {
	content: "✓";
	color: white;
	font-size: 14px;
	font-weight: bold;
	line-height: 1;
}
</style>
