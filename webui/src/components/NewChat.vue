<script>
import router from "../router";
import {BASE_URL, createConversation, createGroup, getConversations, getUsers} from "../services/axios";


export default {
	name: "NewChat",
	props: {
		show: Boolean
	},
	emits: ['close'],
	data(){
		return {
			errormsg: null,

			users:[],
			conversations: [],

			isGroup: false,

			pollingInterval: null, // per il polling

			refreshDebounced: null, //per evitare troppe richieste al server
			searchUsers: "",

			selectedUsers: new Set(), //per i selezionati del gruppo
			groupName:"",
			errorTimeout: null,

		}
	},
	computed: {
		filteredUsers() {
			if (!this.searchUsers) {
				return this.users.filter(
					u => u.userId !== Number(sessionStorage.getItem('userId'))
				);
			}

			const search = this.searchUsers.toLowerCase();

			return this.users.filter(user =>
				user.userId !== Number(sessionStorage.getItem('userId')) &&
				user.username.toLowerCase().includes(search)
			);
		}
	},
	watch: {
		show(newVal) {
			if (newVal) {
				this.startPolling();
			} else {
				this.stopPolling();
			}
		}
	},
	beforeUnmount() {
		this.stopPolling();
	},
	methods:{
		router(){
			return router;
		},
		BASE_URL(){
			return BASE_URL;
		},
		async fetchData() {
			try {
				this.users = await getUsers();
			} catch (e) {
				this.showError("Errore getUsers");
			}

			try {
				this.conversations = await getConversations(
					Number(sessionStorage.getItem('userId'))
				);
			} catch (e) {
				this.showError("Errore getConversations");
			}
		},
		getConversationWith(username) {
			// ritorna la conversazione con quell'utente, se esiste
			return this.conversations.find(conv => {
				return conv.name.includes(username);
			});
		},

		startPolling(){
			this.fetchData();
			this.pollingInterval = setInterval(this.fetchData,2000); // 2 secondi
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

		stopPolling(){
			if (this.pollingInterval) clearInterval(this.pollingInterval);
		},
		async goToConversation(username) {
			let conversation = this.getConversationWith(username);

			if (!conversation) {
				try {
					conversation = await createConversation(
						Number(sessionStorage.getItem('userId')),
						username
					);
					await this.fetchData()
				} catch (e) {
					console.error(e);
					return;
				}
			}

			await router.push({
				name: 'conversation',
				params: {conversationId: conversation.conversationId}
			});
		},
		createGroupBegin(){
			this.isGroup = !this.isGroup;
			if(!this.isGroup){
				this.selectedUsers.clear();
			}
		},
		toggleUser(user){
			const id = user.username;
			this.selectedUsers.has(id) ? this.selectedUsers.delete(id) : this.selectedUsers.add(id);
		},

		async newGroup(){
			const members=[...this.selectedUsers];
			const name = this.groupName;
			if (name !== ""){
				try{
					const group= await createGroup(Number(sessionStorage.getItem('userId')), members, name);
					await router.push({
						name: 'group',
						params: {groupId: group.groupId}
					});
				}catch(e){
					console.error(e);
				}
			}else{
				this.showError("Name can't be empty");
			}

		},
		ClosePanel(){

			this.isGroup = false;
			this.selectedUsers.clear();
			this.groupName = "";

			this.$emit('close');
		}

	},
}

</script>

<template>
  <div v-if="$props.show" class="overlay">
    <div class="action-box">
      <ErrorMsg v-if="errormsg" :msg="errormsg" />
      <div class="header d-flex align-items-center position-relative mb-2">
        <button class="btn btn-close header-close" @click="ClosePanel" />
        <h4 class="header-title">New Chat</h4>
      </div>

      <input v-model="searchUsers" type="text" placeholder="Search user..." class="input-group">
      <div class="top-controls pt-2">
        <button
          :class="['btn', 'h-25', isGroup ? 'btn-danger' : 'btn-outline-primary']"
          :style="{ width: !isGroup ? '100%' : '15%' }"
          @click="createGroupBegin"
        >
          <span v-if="!isGroup">Create Group</span>
          <img v-else src="../icons/reject.png" alt="Cancel" width="25" height="25">
        </button>
        <input v-if="isGroup" v-model="groupName" type="text" placeholder="Name group" class="w-100 ">
      </div>
      <div class="users-box">
        <div v-for="user in filteredUsers" :key="user.userId" class="user-row" @click="isGroup && toggleUser(user)">
          <div v-if="isGroup" class="user-left">
            <input type="checkbox" class="selected" :checked="selectedUsers.has(user.username)">
          </div>
          <div class="user-name">
            <img
              :src="`${BASE_URL()}/file?file=${user.avatar.url}`"
              alt="User Photo"
              :class="['rounded-circle','avatar','mx-2']"
              width="25"
              height="25"
            >
            {{ user.username }}
          </div>
          <div class="user-right">
            <button v-if="!isGroup" class="btn btn-success btn-sm" @click.stop=" goToConversation(user.username)">{{ getConversationWith(user.username) ? 'Open' : 'Create' }}</button>
          </div>
        </div>
      </div>

      <div class="actions">
        <button v-if="isGroup" class="btn btn-success" @click="newGroup">Create</button>
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
	justify-content: end;
	gap: 10px;
	margin-top: 10px;
}

.users-box > div {
	display: flex;
	justify-content: space-between;
	align-items: center;
	overflow-y: auto;
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

.user-row {
	display: flex;
	align-items: center;
	padding: 8px 6px;
	border-radius: 6px;
	cursor: pointer;
}

.user-row:hover {
	background-color: #f0f4ff;
}

.user-left {
	width: 26px;
	display: flex;
	justify-content: center;
}

.user-name {
	flex: 1;
	min-width: 0;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.user-right {
	width: 80px;
	display: flex;
	justify-content: flex-end;
}


.top-controls {
	display: flex;
	gap: 8px;
	width: 100%;
}


.top-controls > input {
	flex: 1; /* prende lo spazio rimanente */
}

.avatar {
	object-fit: cover;    /* taglia l’immagine mantenendo proporzioni 100x100 */
}

.header {
	height: 40px; /* riferimento comune */
}

.header-close {
	position: absolute;
	left: 0;
	top: 50%;
	transform: translateY(-50%);
}

.header-title {
	position: absolute;
	left: 50%;
	transform: translate(-50%, -50%);
	top: 50%;
	margin: 0;
}
</style>
