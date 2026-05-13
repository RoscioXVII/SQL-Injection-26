<script>

import {
	BASE_URL,
	deleteFromGroup,
	getGroup,
	getUserInfo,
	setGroupName,
	setGroupPhoto
} from "../services/axios";
import router from "../router";
import AddUser from "./AddUser.vue";

export default {
	components: {AddUser},
	props:{
		show: Boolean,
		groupId:Number,
		userId:Number,
	},
	emits:['close'],
	data(){
		return {
			group:null,
			users:[],
			editMode:false,
			addUserMode:false,
			pollInterval: null,
			newGroupName:"",
			selectedFile:null,
			previewUrl:undefined,
		}
	},
	mounted(){
		this.refresh();
		this.pollInterval = setInterval(() => {
			if (!this.addUserMode && this.show) {   //non serve menre aggiungo utenti
				this.refresh();
			}
		}, 3000);
	},
	beforeUnmount(){
		if (this.pollInterval) {
			clearInterval(this.pollInterval);
		}
	},
	methods:{
		router(){
			return router;
		},
		BASE_URL(){
			return BASE_URL;
		},
		async refresh(){
			if(this.groupId){
			try{
				const group = await getGroup(this.userId, this.groupId);
				// aggiorna solo se cambia qualcosa
				if (JSON.stringify(group.participants) !== JSON.stringify(this.users)) {
					this.group = group;
					this.users = await Promise.all(
						group.participants.map(async user => {
							const userInfo = await getUserInfo(String(user.userId));
							return {
								...user,
								avatar: userInfo.avatar
							};
						})
					);
				}
			}catch(error){
				console.error(error);
			}
		}},
		closePanel(){
			this.editMode = false;
			this.$emit('close');
		},
		leaveGroup(){
			try{
				deleteFromGroup(this.userId,this.groupId);
			}catch(error){
				console.error(error);
			}finally {
				router.push('/home');
				this.$emit('close');
			}

		},

		addUserLayout(){
			this.addUserMode = true;
			this.editMode = false;
		},
		EditGroup(){
			this.editMode = true;
			this.addUserMode=false;
		},
		EditGroupRevert(){
			this.editMode = false;
			this.newGroupName = "";
			this.previewUrl=undefined;
			this.selectedFile = null;
		},
		onFileChange(e) {
			const file = e.target.files[0];
			if (!file) return;

			this.selectedFile = file;
			this.previewUrl = URL.createObjectURL(file);
		},
		async saveChanges(){
			if(this.newGroupName !== this.group.name && this.newGroupName !== ""){
				try{
					await setGroupName(this.userId,this.groupId,this.newGroupName);
				}catch(error){
					console.error(error);
				}
				this.newGroupName = "";
			}
			if(this.selectedFile){
				try{
					await setGroupPhoto(this.userId,this.groupId,this.selectedFile);
				}catch(error){
					console.error(error);
				}
				this.previewUrl=undefined;
			}

			this.editMode = false;
			await this.refresh();
		}
	}
}

</script>

<template>
  <div v-if="show" class="overlay">
    <div class="action-box">
      <div class="header d-flex align-items-center justify-content-between">
        <button class="btn btn-close" @click="closePanel" />
        <h4 class="h4">Info Group</h4>
        <button v-if="!editMode" class="btn btn-clean" :disabled="addUserMode" style="border-radius: 50px; padding: 0; height: 45px; width: 45px;" @click="EditGroup">
          <img src="../icons/edit.png" alt="Edit" width="20" height="20" class="mb-1">
        </button>
        <div v-if="editMode" class="btn-group">
          <button class="btn btn-clean" style="border-radius: 50px; padding: 0; height: 45px; width: 45px;" @click="EditGroupRevert">
            <img src="../icons/back-arrow.png" alt="Edit" width="25" height="25" class="mb-1">
          </button>
          <button v-if="editMode" class="btn btn-clean" type="submit" @click="saveChanges">
            <img src="../icons/check.png" alt="Edit" width="25" height="25" class="mb-1">
          </button>
        </div>
      </div>
      <div class="d-flex align-items-center justify-content-center flex-column gap-3">
        <div class="d-flex align-items-center gap-3">
          <img class="avatar rounded-circle" :src="`${BASE_URL()}/file?file=${group.upload.url}`" :width="previewUrl && editMode? 100:200" :height="previewUrl && editMode? 100:200" alt="Photo">
          <img v-if="previewUrl && editMode" src="../icons/right-arrow.png" alt="Arrow to ..." width="50" height="50">
          <img
            v-if="previewUrl && editMode"
            :src="previewUrl"
            alt="Avatar"
            class="rounded-circle avatar"
            width="100"
            height="100"
          >
        </div>

        <div v-if="editMode">
          <label for="fileInput" class="btn btn-outline-primary">Select Group Photo</label>
          <input id="fileInput" type="file" class="d-none" @change="onFileChange">
        </div>

        <div class="d-flex justify-content-between">
          <div>
            <h4 v-if="!editMode" class="fw-bold text-center name-display">{{ group.name }}</h4>
            <input v-if="editMode" v-model="newGroupName" type="text" class="name-input mb-0 text-center name-input" :placeholder="group.name">
          </div>
        </div>
      </div>
      <AddUser
        :show="addUserMode && !editMode"
        :group-id="groupId"
        :user-id="userId"
        @close="addUserMode=false"
        @members-updated="refresh"
      />
      <div v-if="!addUserMode" class="d-flex flex-column gap-3 mx-2 participants-wrapper">
        <div class="d-flex justify-content-between align-items-center">
          <h2 class="h2">Participants</h2>
          <button
            class="btn d-flex align-items-center justify-content-center btn-clean"
            style="border-radius: 50px; padding: 0; height: 45px; width: 45px;"
            :disabled="editMode"
            @click="addUserLayout"
          >
            <img src="../icons/plus.png" alt="Add" width="20" height="20">
          </button>
        </div>

        <div class="participants-list ">
          <div v-for="user in users" :key="user.userId" class="pb-3 d-flex gap-2  align-items-center">
            <img class="avatar rounded-circle" :src="`${BASE_URL()}/file?file=${user.avatar.url}`" width="50" height="50" alt="Photo">
            <span class="text-muted">{{ user.username }}</span>
          </div>
        </div>
      </div>

      <div class="d-flex justify-content-between">
        <button v-if="!addUserMode" class="btn btn-danger" @click="leaveGroup">Leave Group</button>
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
	max-height: 85vh;
	border-radius: 8px;
	box-shadow: 0 4px 15px rgba(0,0,0,0.3);
	z-index: 1300;
	overflow-y: auto;
	display: flex;
	flex-direction: column;
}


.header {
	height: 40px; /* riferimento comune */
}

.avatar {
	object-fit: cover;    /* taglia l’immagine mantenendo proporzioni 100x100 */
}

.participants-wrapper {
	flex: 1;
	display: flex;
	flex-direction: column;
	overflow: hidden; /* evita overflow esterno */
}

.participants-list {
	overflow-y: auto;          /* scroll verticale */
	max-height: 300px;         /* altezza massima per scroll */
	padding-right: 5px;        /* evita che il scrollbar tocchi il testo */
}
.name-display,
.name-input {
	font-size: 1.5rem;
	font-weight: bold;
	line-height: 1.2;
	width: 100%;
	max-width: 200px;
	box-sizing: border-box;
	text-align: center;
}

.btn-clean {
	outline: none !important;
	box-shadow: none !important;
	border: none !important;
}

.btn-clean:focus,
.btn-clean:focus-visible,
.btn-clean:active,
.btn-clean:active:focus {
	outline: none !important;
	box-shadow: none !important;
}

</style>
