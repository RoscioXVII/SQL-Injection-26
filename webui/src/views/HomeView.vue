<script>
import {BASE_URL, getConversations, getGroups, getUserInfo, setPhotoUser, setUsername} from "../services/axios";
import router from "../router";
import {nextTick} from "vue";
import LoadingSpinner from "../components/LoadingSpinner.vue";
import NewChat from "../components/NewChat.vue";
import ShowParticipants from "../components/ShowParticipants.vue";
import ErrorMsg from "@/components/ErrorMsg.vue";

export default {
	components: {ErrorMsg, ShowParticipants, LoadingSpinner,NewChat},
	data() {
		return {
			errormsg: null,
			errorTimeout: null,
			loading: false,

			conversations: [],
			groups: [],

			username: null,
			userId: null,
			token: null,

			activeFilter: "all",
			sidebarOpen: false,

			searchQuery: "",

			newChat:false,

			pollingInterval: null, //per il polling

			firstLoad: true, //per evitare blink effect

			editMode: false, //per modifica nome e foto

			myActualPhoto: null,

			newUsername:"",

			selectedFile: null,

			previewUrl:undefined,

			showParticipants: false,
			groupId_info: null,
		}
	},
	computed: {
		filteredChats() {
			// unisce conversations e groups
			let result = [...this.conversations, ...this.groups];
			result.sort((a, b) => { // sorto per novità
				// a senza messaggi, b con messaggi → a prima
				if (!a.lastMessage && b.lastMessage) return -1;

				// b senza messaggi, a con messaggi → b prima
				if (!b.lastMessage && a.lastMessage) return 1;

				// entrambe senza messaggi → ordine stabile
				if (!a.lastMessage && !b.lastMessage) return 0;

				// entrambe con messaggi → ordina per data
				return Date.parse(b.lastMessage.time) - Date.parse(a.lastMessage.time);
			});
			// filtro per tipo
			if(this.activeFilter === "direct") {
				result = result.filter(chat => chat.conversationId);
			} else if(this.activeFilter === "group") {
				result = result.filter(chat => chat.groupId);
			}

			// filtro per ricerca
			if(this.searchQuery.trim() !== "") {
				const q = this.searchQuery.toLowerCase();
				result = result.filter(chat =>
					chat.name.toLowerCase().startsWith(q)
				);
			}

			return result;
		}
	},
	watch: {
		sidebarOpen(newVal) {
			if (newVal) {
				nextTick(() => {
					const navbarHeight = document.querySelector('header.navbar')?.offsetHeight || 0;
					const sidebar = document.querySelector('.sidebar');
					if(sidebar){
						sidebar.style.top = `${navbarHeight}px`;
						sidebar.style.height = `calc(100% - ${navbarHeight}px)`;
					}
				});
			}
		}
	},
	created() {
		this.username = sessionStorage.getItem("username");
		this.token = sessionStorage.getItem("token");
		this.userId = sessionStorage.getItem("userId");
		this.newUsername= this.username;

		this.sidebarOpen = false;
		this.infoGroup = false;
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
		async refresh() {
			if (this.firstLoad) {
				this.loading = true;
			}
			this.errormsg = null;
			this.userId = Number(sessionStorage.getItem('userId'));
			this.token = sessionStorage.getItem('token');
			const info = await getUserInfo(this.userId);
			this.username = info.username;
			this.myActualPhoto = info.avatar;

			if(!this.userId || !this.token){
				this.errormsg="Do the Login"
				this.loading=false;
				return;
			}
			try{
				let convResponse = await getConversations(this.userId);
				this.conversations = convResponse || [];

				let groupResponse = await getGroups(this.userId);
				this.groups = groupResponse || [];
			}catch(err){
				this.errormsg=err.toString();
			}
			if (this.firstLoad) {
				this.loading = false;
				this.firstLoad = false;
			}
		},
		openChat(id, type){
			if (type === "direct"){
				router.push({ name: 'conversation', params: { conversationId: id } });
			} else if(type === "group"){
				router.push({ name: 'group', params: { groupId: id } });
			} else console.error("type must be direct or group");
		},
		toggleSidebar() {
			this.sidebarOpen = !this.sidebarOpen;
		},
		Logout(){
			sessionStorage.removeItem('userId');
			sessionStorage.removeItem('token');
			sessionStorage.removeItem('username');
			router.push('/');
		},
		goToParticipants(groupId) {
			this.groupId_info = groupId;
			this.showParticipants = true;
		},
		openNewChat() {
			this.newChat = !this.newChat;
		},
		startPolling() {
			// evita doppi interval
			if (this.pollingInterval) return;

			this.pollingInterval = setInterval(() => {
				this.refresh();
			}, 2000);
		},

		stopPolling() {
			if (this.pollingInterval) {
				clearInterval(this.pollingInterval);
				this.pollingInterval = null;
			}
		},
		EditMode(){
			this.editMode = true;
		},
		Cancel(){
			this.editMode = false;
			this.selectedFile = null;
			this.previewUrl = undefined;
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

		async CommitChanges() {
			if (this.newUsername !== this.username) {
				try {
					const result = await setUsername(this.userId, this.newUsername);
					sessionStorage.setItem('username', result.username);
				} catch (err) {
					sessionStorage.setItem('username',this.username);
					this.newUsername = '';
					if (err.response.status === 409) {
						this.showError("Username già in uso");
						return
					} else {
						console.error("error while saving changes", err);
						this.showError("Error while saving changes", err);
						return;
					}
				}

			}
			if (this.selectedFile) {
				try {
					await setPhotoUser(this.userId, this.selectedFile);
				} catch (err) {
					console.error(err);
				}
			}
			this.previewUrl = undefined;
			this.editMode = false;
			await this.refresh();

		},

		onFileChange(e) {
			const file = e.target.files[0];
			if (!file) return;

			this.selectedFile = file;
			this.previewUrl = URL.createObjectURL(file);
		},

		closePanel() {
			this.groupId_info=null;
			this.showParticipants = false;
		},


	}
}
</script>

<template>
  <LoadingSpinner v-if="loading ===true" />
  <NewChat
    :show="newChat"
    @close="newChat = false"
  />
  <div>
    <div class="row align-items-center pt-3 pb-2 mb-3 border-bottom">
      <div class="col-3 d-flex justify-content-start">
        <div>
          <button class="hamburger" @click="toggleSidebar">
            <span />
            <span />
            <span />
          </button>

          <div class="sidebar" :class="{ open: sidebarOpen }">
            <div class="d-flex justify-content-center flex-column align-items-center gap-2">
              <img
                :src="`${BASE_URL()}/file?file=${myActualPhoto.url}`"
                alt="Avatar"
                class="rounded-circle avatar "
                width="100"
                height="100"
              >
              <img
                v-if="previewUrl && editMode"
                src="../icons/arrow-downward-svgrepo-com.svg"
                alt="Arrow down"
                width="50"
                height="50"
              >

              <img
                v-if="previewUrl && editMode"
                :src="previewUrl"
                alt="Avatar"
                class="rounded-circle avatar"
                width="100"
                height="100"
              >


              <div>
                <h1 v-if="!editMode" class="name-display mb-0">{{ username }}</h1>
                <input v-if="editMode" v-model="newUsername" type="text" class="name-input mb-0 text-center" :placeholder="username">
              </div>
              <ErrorMsg v-if="errormsg" :msg="errormsg" />
              <div v-if="editMode" class="d-flex justify-content-center flex-column align-items-center mt-2">
                <label for="fileInput" class="btn btn-outline-primary">
                  Choose New Photo
                </label>
                <input
                  id="fileInput"
                  ref="fileInput"
                  type="file"
                  class="d-none"
                  @change="onFileChange"
                >
              </div>
            </div>
            <button v-if="!editMode" class="btn btn-outline-primary w-100" @click="EditMode">Edit Profile</button>
            <button v-if="editMode" class="btn btn-outline-success w-100" @click="CommitChanges">Save</button>
            <button v-if="editMode" class="btn btn-outline-danger w-100" @click="Cancel">Cancel</button>
            <button id="Logout" class="btn w-100" @click="Logout">Logout</button>
          </div>

          <div v-if="sidebarOpen" class="overlay" @click="toggleSidebar" />
        </div>
      </div>

      <div class="col-6 text-center">
        <h1 class="h2 mb-0">Welcome Back, {{ username }}!</h1>
      </div>
    </div>


    <LoadingSpinner v-if="loading" />

    <div v-if="!loading" class="row">
      <div class="d-flex justify-content-between mb-3">
        <div class="d-flex gap-2">
          <button
            class="btn btn-outline-primary"
            :class="{ selected: activeFilter === 'all'}"
            @click="activeFilter = 'all'"
          >
            All
          </button>
          <button
            class="btn btn-outline-primary"
            :class="{ selected: activeFilter === 'direct'}"
            @click="activeFilter = 'direct'"
          >
            Chats
          </button>
          <button
            class="btn btn-outline-primary"
            :class="{ selected: activeFilter === 'group'}"
            @click="activeFilter = 'group'"
          >
            Groups
          </button>
        </div>
        <div class="d-flex justify-content-end gap-2">
          <input v-model="searchQuery" type="text" class="form-control" placeholder="Search...">
          <button class="btn btn-primary btn-sm btn-rad" @click="openNewChat">
            <img src="../icons/new-svgrepo-com.svg" width="25" height="25" alt="New">
          </button>
        </div>
      </div>

      <div v-if="filteredChats.length === 0" class="alert alert-secondary">
        No chats found.
      </div>
      <ShowParticipants
        :show="showParticipants"
        :group-id="groupId_info"
        :user-id="userId"
        @close="closePanel"
      />
      <ul class="list-group shadow-sm">
        <li v-for="chat in filteredChats" :key="chat.conversationId || chat.groupId" class="list-group-item d-flex justify-content-between align-items-center py-3">
          <div>
            <img
              :src="chat.conversationId ? `${BASE_URL()}/file?file=${chat.avatar.url}` : `${BASE_URL()}/file?file=${chat.upload.url}`"
              alt="Avatar"
              class="rounded-circle avatar"
              width="40"
              height="40"
            >
            <span class="fw-bold ms-2">{{ chat.name }}</span>
            <br>

            <div v-if="chat.lastMessage">
              <small v-if="!chat.lastMessage.body.photo" class="text-muted">
                {{ chat.lastMessage.sender.userId === String(userId) ? 'You' : chat.lastMessage.sender.username }} :
                {{ (!chat.lastMessage.body.text && !chat.lastMessage.body.photo) && chat.lastMessage.isForwarded ? "Message forwarded more times..." : chat.lastMessage.body.text }}
              </small>

              <small v-else class="text-muted">
                {{ chat.lastMessage.sender.userId === String(userId) ? 'You' : chat.lastMessage.sender.username }} : Photo {{ chat.lastMessage.body.text ? ": " + chat.lastMessage.body.text : "" }}
              </small>
            </div>

            <div v-else>
              <small class="text-muted fst-italic">No messages yet</small>
            </div>
          </div>
          <button class="btn btn-outline-primary btn-sm" @click="openChat(chat.conversationId || chat.groupId, chat.conversationId ? 'direct' : 'group')">
            Open
          </button>
        </li>
      </ul>
    </div>
  </div>
</template>

<style scoped>

.hamburger {
	display: flex;
	flex-direction: column;
	justify-content: space-around;
	width: 30px;
	height: 25px;
	background: transparent;
	border: none;
	cursor: pointer;
	padding: 0;
	z-index: 1100;
}
.selected {
	font-weight: bold;
	background-color: #0d55c7;
	color: white;
	border-color: #0d6efd;
}
.hamburger span {
	display: block;
	width: 100%;
	height: 3px;
	background-color: #333;
	border-radius: 2px;
}
.sidebar {
	position: fixed;
	left: -260px;
	width: 260px;
	height: 100%;
	background-color: #ffffff;
	padding: 1.5rem 1rem;
	box-shadow: 2px 0 10px rgba(0,0,0,0.2);
	transition: left 0.3s ease;
	z-index: 1000;
	display: flex;
	flex-direction: column;
	gap: 0.75rem;
}
.sidebar.open {
	left: 0;
}
.overlay {
	position: fixed;
	top: 0;
	left: 0;
	width: 100%;
	height: 100%;
	background: rgba(0,0,0,0.3);
	z-index: 999;
}
.btn {
	transition: background-color 0.2s ease, color 0.2s ease, border-color 0.2s ease;
}
#Logout {
	background-color: red;
	color: white;
	border-radius: 5px;
	font-weight: bold;
	margin-top: auto;
}
.participant{
	color: rgba(128, 128, 128, 0.38);
	transition: background-color 0.2s ease, color 0.2s ease, border-color 0.2s ease;
}
.participant:hover {
	color: rgb(108, 108, 108);
	transition: background-color 0.2s ease, color 0.2s ease, border-color 0.2s ease;
}

.name-display,
.name-input {
	font-size: 2rem;
	font-weight: bold;
	line-height: 1.2;
	width: 100%;
	max-width: 200px;
	box-sizing: border-box;
	text-align: center;
}
.name-input {
	padding: 0.25rem 0.5rem;
}

.avatar {
	object-fit: cover;    /* taglia l’immagine mantenendo proporzioni 100x100 */
}

.btn img {
	filter: brightness(0) invert(1);
}

.btn-rad{
	border-radius: 6px;
}

</style>
