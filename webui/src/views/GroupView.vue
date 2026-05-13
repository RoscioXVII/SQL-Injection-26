<script>
import {BASE_URL, getConversation, getMessages, getUserInfo, postMessage} from "../services/axios";
import router from "../router";
import LoadingSpinner from "../components/LoadingSpinner.vue";
import Comment from "../components/Comment.vue";
import Delete from "../components/Delete.vue";
import Forward from "../components/Forward.vue";
import InfoGroup from "../components/InfoGroup.vue";


export default {
	components: {InfoGroup, Forward, Delete, Comment, LoadingSpinner},
	data() {
		return {
			errormsg: null,
			errorTimeout:null,
			loading: false,

			messages: [],
			currentGroup: null,

			username: null,
			userId: null,
			token: null,
			groupId: null,

			searchMessage: "",

			pollingInterval: null, // per il polling

			refreshDebounced: null, //per evitare troppe richieste al server

			openMessageOptions:null,

			selectedMessageId : null,

			showForward: false,
			forwardMessageId: null,
			deleteMessageId: null,
			deleteMessage: false,
			showComments: false,
			commentMessageId: null,

			infoGroup:false,
			photoSend:false,

			replyToMsg: null,
		};
	},
	computed: {
		filteredMessages() {
			let result = this.messages

			if(this.searchMessage.trim().length !== ""){
				const q = this.searchMessage.toLowerCase();
				result=result.filter(chat => chat.body.text.toLowerCase().includes(q));
			}
			return result;
		}
	},
	watch: {
		'$route.params.groupId'(newId) {
			this.stopPolling();
			this.groupId = newId;
			this.messages=[];
			this.replyToMsg=null;
			this.openMessageOptions=null;
			this.refresh();
			this.startPolling();
		}
	},
	created() {
		this.username = sessionStorage.getItem("username");
		this.token = sessionStorage.getItem("token");
		this.userId = sessionStorage.getItem("userId");
		this.groupId = this.$route.params.groupId;

		if (this.token && this.userId) {
			this.refresh();
			this.startPolling();
		}
	},
	beforeUnmount() {
		this.stopPolling();
	},
	methods: {
		router() {
			return router;
		},
		BASE_URL() {
			return BASE_URL;
		},

		async fetchMessages() {
			try {
				const msgs = await getMessages(this.userId, this.groupId,"group") || [];

				const container = this.$refs.messageContainer;
				let isAtBottom = false;

				if (container) {
					// verifica se siamo già in fondo (tolleranza 20px)
					isAtBottom = container.scrollTop + container.clientHeight >= container.scrollHeight - 20;
				}

				this.messages = msgs;
				await this.$nextTick();
				if (container && isAtBottom) {
					container.scrollTop = container.scrollHeight;
				}

			} catch (err) {
				if(err.response?.data?.includes("database is locked")) {
					setTimeout(() => this.fetchMessages(), 2000); // ritenta dopo 2 secondi
				} else {
					console.error("Errore fetching messages:", err);
				}

			}
		},

		async refresh() {
			if(this.refreshDebounced)   clearTimeout(this.refreshDebounced);

			this.refreshDebounced = setTimeout(async() => {

				this.errormsg = null;


				this.userId = this.userId || sessionStorage.getItem("userId");
				this.token = this.token || sessionStorage.getItem("token");
				this.groupId = this.groupId || this.$route.params.groupId;

				if (!this.userId || !this.token) {
					this.errormsg = "Do login!";
					this.loading = false;
					return;
				}

				try {
					this.currentGroup = await getConversation(this.userId, this.groupId,"group");
					this.loading = true;
					await this.fetchMessages(); // primo caricamento messaggi
					this.loading = false;
					for (const msg of this.messages) {
						const user = await getUserInfo(msg.sender.userId);
						msg.sender.avatar = user.avatar;
					}
				} catch (err) {
					this.errormsg = err.toString();
				} },300); //300ms tra una richiesta e l'altra
		},

		startPolling() {
			this.fetchMessages(); // fetch iniziale
			this.pollingInterval = setInterval(() => {
				this.refresh();
				this.fetchMessages();
			}, 2000); // ogni 2 secondi
		},

		stopPolling() {
			if (this.pollingInterval) {
				clearInterval(this.pollingInterval);
				this.pollingInterval = null;
			}
		},
		async sendMessageButton() {
			const textInput = this.$refs.messageText;
			const photoInput = this.$refs.messagePhoto;

			const text = textInput.value;
			const photo = photoInput.files[0];

			// controllo estensione foto
			if (photo && photo.type !== "image/png" && photo.type !== "image/jpeg") {
				this.showError("Only PNG or JPEG!");
				photoInput.value = "";
				return;
			}

			if (text || photo) {
				try {
					await postMessage(this.userId, this.groupId, text, photo,"group",this.replyToMsg ? this.replyToMsg.messageId : null);
					textInput.value = "";
					photoInput.value = "";
					this.replyToMsg = null;
					await this.fetchMessages();
				} catch (e) {
					console.error("Error send message:", e);
					this.showError("Error sending message");
				}finally {
					await this.refresh();
				}
			}
			this.photoSend = false;
		},

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
		showOptions(messageId){
			if (this.openMessageOptions === messageId) {
				this.openMessageOptions = null;
			} else {
				this.openMessageOptions = messageId;
			}
		},
		forward(messageId) {
			this.forwardMessageId = messageId;
			this.showForward = true;
		},
		comment(messageId){
			this.showComments = true;
			this.commentMessageId = messageId;
		},
		deleteM(messageId){
			this.deleteMessage = true;
			this.deleteMessageId = messageId;
		},
		onPhotoSelected(event) {
			const file = event.target.files[0];
			if (!file) return;

			if (file.type !== "image/png" && file.type !== "image/jpeg") {
				this.showError("Only PNG or JPEG!");
				event.target.value = "";
				return;
			}
			this.photoSend = true;
		},
		replyMessage(message){
			this.replyToMsg = message;
		},
		getMessage(messageId){
			const message= this.messages.find(message => message.messageId === messageId);
			return message ? message : null;
		}
	}
};
</script>

<template>
  <LoadingSpinner v-if="loading ===true" />
  <Comment
    :user-id="userId"
    :show="showComments"
    :message-id="commentMessageId"
    :chat-id="groupId"
    :type="`group`"
    @close="showComments=false"
  />
  <Delete
    :user-id="userId"
    :show="deleteMessage"
    :message-id="deleteMessageId"
    :chat-id="groupId"
    :type="`group`"
    @close="deleteMessage=false"
  />
  <Forward
    :user-id="userId"
    :show="showForward"
    :message-id="forwardMessageId"
    :chat-id="groupId"
    :type="`group`"
    @close="showForward = false"
  />
  <InfoGroup
    :user-id="userId"
    :group-id="groupId"
    :show="infoGroup"
    @close="infoGroup=false"
  />
  <div class="chat-header-wrapper mb-2">
    <div class="d-flex justify-content-between align-items-center">
      <div class="d-flex align-items-center gap-2 p-lg-2 info" @click="infoGroup=true">
        <img
          :src="currentGroup.upload.url ? `${BASE_URL()}/file?file=${currentGroup.upload.url}` : '' "
          alt="photo"
          class="rounded-circle align-self-end avatar"
          width="45"
          height="45"
        >
        <h1 class="mb-0 align-bottom">{{ currentGroup.name }}</h1>
      </div>
      <div>
        <input v-model="searchMessage" type="text" placeholder="Search message...">
      </div>
    </div>
  </div>

  <div class="chat-body">
    <div ref="messageContainer" class="messages-container">
      <ErrorMsg v-if="errormsg" :msg="errormsg" />
      <div v-if="filteredMessages.length === 0" class="no-messages">
        <h1>No Messages...</h1>
      </div>
      <div
        v-for="msg in filteredMessages"
        :key="msg.messageId"
        :class="['message-row gap-3', msg.sender.userId === userId ? 'mine' : 'theirs']"
      >
        <img
          v-if="msg.sender.userId !== userId"
          :src="msg.sender.avatar?.url ? `${BASE_URL()}/file?file=${msg.sender.avatar.url}` : '' "
          alt="Avatar"
          class="rounded-circle align-self-end avatar"
          width="35"
          height="35"
        >
        <div class="message-bubble">
          <div
            v-if="openMessageOptions === msg.messageId"
            class="d-flex justify-content-end"
          >
            <button class="btn btn-outline-info" @click="replyMessage(msg)"><img src="../icons/reply.png" alt="Reply" width="25" height="25"></button>
            <button class="btn btn-outline-secondary" @click="forward(msg.messageId)"><img src="../icons/share-icon_4662621.png" alt="Forward" width="25" height="25"></button>
            <button class="btn btn-outline-primary" @click="comment(msg.messageId)"><img src="../icons/chat-dots-fill.svg" alt="comment" width="23" height="23"></button>
            <button v-if="msg.sender.userId === userId" class="btn btn-outline-danger" @click="deleteM(msg.messageId)"><img src="../icons/trash3-fill.svg" alt="Delete" width="23" height="23"></button>
          </div>
          <div v-if="msg.replyTo" class="reply d-flex flex-column">
            <template v-if="(repliedMsg = getMessage(msg.replyTo))">
              <span class="text-muted">{{ repliedMsg.sender.username }}</span>
              <img v-if="repliedMsg.body.photo && repliedMsg.body.photo.url" :src="`${BASE_URL()}/file?file=${repliedMsg.body.photo.url}`" alt="PhotoMessage" class="message-photo-reply">
              <span class="text-muted">{{ repliedMsg.body.text }}</span>
            </template>
            <template v-else>
              <span class="text-muted text-center">(message unavailable)</span>
            </template>
          </div>
          <div class="d-flex justify-content-between">
            <small v-if="msg.sender.userId !== userId" class="sender">{{ msg.sender.username }}</small>
            <small v-if="msg.sender.userId === userId" class="sender" />
            <div class="d-flex">
              <small v-if="msg.isForwarded" class="text-end">(forwarded)</small>
              <button class="icon-btn" @click="showOptions(msg.messageId)"><img src="../icons/dots_16164512.png" width="15" height="15" alt="Options"></button>
            </div>
          </div>
          <img v-if="msg.body.photo && msg.body.photo.url" :src="`${BASE_URL()}/file?file=${msg.body.photo.url}`" alt="PhotoMessage" class="message-photo">
          <p class="text">{{ msg.body.text }}</p>
          <div class="justify-content-between">
            <span>{{ msg.time }}</span>
            <img v-if="msg.read==='received'" src="../icons/icons8-double-tick-100.png" width="15" height="15" alt="received">
            <img v-if="msg.read==='read'" src="../icons/icons8-double-tick-100-2.png" width="15" height="15" alt="read">
          </div>
        </div>
      </div>
    </div>
    <div class="d-flex flex-column gap-2 mt-2">
      <div v-if="replyToMsg" class="d-flex justify-content-between">
        <div class="d-flex flex-column align-items-start">
          <span>Reply To {{ replyToMsg.sender.username }}</span>
          <span class="text-truncate">Message: {{ replyToMsg.body.photo ? "📷 " : "" }} {{ replyToMsg.body.photo && replyToMsg.body.text ? "|" : "" }} {{ replyToMsg.body.text }}  </span>
        </div>

        <button class="btn-close align-self-center " @click="replyToMsg=null" />
      </div>
      <div class="d-flex justify-content-between gap-2">
        <div class="d-flex icon">
          <input id="fileInput" ref="messagePhoto" type="file" class="d-none" @change="onPhotoSelected">
          <label
            for="fileInput"
            class="btn btn-primary btn-sm d-flex align-items-center justify-content-center"
          >
            <img v-if="!photoSend" src="../icons/photo-svgrepo-com.svg" alt="Send Photo" width="25" height="25" class="icon">
            <img
              v-if="photoSend"
              src="../icons/check.png"
              width="25" height="25"
              alt="Preview"
            >
          </label>
        </div>

        <input ref="messageText" class="input-group" type="text" placeholder="Write message...">
        <button class="btn btn-dark" @click="sendMessageButton">
          <img src="../icons/send-message-svgrepo-com.svg" alt="Send Message" width="25" height="25" class="icon">
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.chat-header-wrapper {
	margin-top: 1rem;
}

.chat-body {
	height: calc(100vh - 140px); /* header + margini */
	display: flex;
	flex-direction: column;
}

.messages-container {
	position: relative;
	flex: 1;
	overflow-y: auto;
	padding: 1rem;
	background-color: #f8f9fa;
	border-radius: 20px;
	flex-direction: column-reverse;
	display: flex;
}


.sender {
	font-weight: bold;
	font-size: 0.85rem;
}

.text {
	margin: 0;
}

.message-row {
	display: flex;
	margin-bottom: 0.75rem;
}

.message-row.mine {
	justify-content: flex-end;
}

.message-row.theirs {
	justify-content: flex-start;
}

.message-bubble {
	position: relative;
	max-width: 70%;
	padding: 0.6rem 0.9rem;
	border-radius: 14px;
	background-color: #ffffff;
	box-shadow: 0 1px 3px rgba(0,0,0,0.1);
	word-wrap: break-word;
}

/* colori diversi */
.message-row.mine .message-bubble {
	background-color: #7cd65a;
	color: white;
	border-bottom-right-radius: 4px;
}

.message-row.theirs .message-bubble {
	background-color: #9adf97;
	color: #212529;
	border-bottom-left-radius: 4px;
}

.sender {
	font-size: 0.7rem;
	opacity: 0.7;
	display: block;
	margin-bottom: 2px;
}

.message-photo {
	max-width: 300px;   /* larghezza massima */
	max-height: 300px;  /* altezza massima */
	width: auto;        /* scala proporzionalmente */
	height: auto;
	border-radius: 10px;
	object-fit: contain; /* mantiene proporzioni */
}


.message-photo-reply {
	max-width: 100px;   /* larghezza massima */
	max-height: 100px;  /* altezza massima */
	width: auto;        /* scala proporzionalmente */
	height: auto;
	border-radius: 10px;
	object-fit: contain; /* mantiene proporzioni */
}

.icon-btn {
	padding-bottom: 5px;
	border: none;
	background: none;
	display: inline-flex;
	align-items: center;
	justify-content: center;
}

.avatar {
	object-fit: cover;    /* taglia l’immagine mantenendo proporzioni 100x100 */
}

.btn-primary img {
	filter: brightness(0) invert(1);
}

.btn-dark img {
	filter: brightness(0) invert(1);

}

.btn{
	border-radius: 6px;
}


.no-messages {
	position: absolute;
	inset: 0; /* top right bottom left = 0 */
	display: flex;
	align-items: center;
	justify-content: center;
	pointer-events: none; /* non blocca scroll/click */
}

.no-messages h1 {
	opacity: 0.5;
}

.info{
	cursor: pointer;
}


.message-row.mine .reply {
	background-color: #9adf97; /* colore "theirs" */
	color: #212529;
}

.message-row.theirs .reply {
	background-color: #7cd65a; /* colore "mine" */
	color: white;
}

.reply {
	padding: 4px 8px;
	border-radius: 8px;
	font-size: 0.8rem;
	margin-bottom: 6px;
}
</style>
