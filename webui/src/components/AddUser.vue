<script>

import {addToGroup, BASE_URL, getGroup, getUsers} from "../services/axios";

export default {
	props: {
		show:Boolean,
		groupId:Number,
		userId:Number,
	},
	emits:[`close`,'members-updated'],
	data(){
		return {
			users:[],
			selectedUsers:[],
			participants:[],
			isSubmitting: false
		}
	},
	watch: {
		show(newVal) {
			if (newVal) {
				this.refresh();
			}
		}
	},
	mounted(){
		this.refresh();
	},
	methods: {
		BASE_URL(){
			return BASE_URL;
		},
		async refresh(){
			try{
				this.users = await getUsers();
				const group= await getGroup(this.userId,this.groupId);
				this.participants=group.participants;
				this.selectedUsers = this.participants.map(p => p.username); //preseleziono i già membri
			}catch (error){
				console.error(error);
			}
		},
		closePanel(){
			this.selectedUsers=[];
			this.$emit('close');
		},
		async addSelectedUsers() {
			this.isSubmitting = true;
			try {
				const newUsers = this.selectedUsers.filter(
					u => !this.isAlreadyMember(u)
				);

				for (const username of newUsers) {
					await addToGroup(this.userId, this.groupId, username);
				}

				this.closePanel();
			} catch (error) {
				console.error(error);
			}finally {
				this.isSubmitting = false;
			}
			this.$emit('members-updated');
		},
		isAlreadyMember(username) {
			return this.participants.some(p => p.username === username);
		}
	}
}
</script>

<template>
  <div v-if="show" class="d-flex flex-column gap-3 mx-2 mb-3">
    <div class="d-flex justify-content-between align-items-center">
      <h2 class="h2">Add Members</h2>
      <button
        class="btn d-flex align-items-center justify-content-center btn-clean"
        style="border-radius: 50px; padding: 0; height: 45px; width: 45px;"
        @click="closePanel"
      >
        <img src="../icons/minus-sign.png" alt="Add" width="20" height="20">
      </button>
    </div>
    <div class="box">
      <div v-for="user in users" :key="user.userId" class="user-row " :class="{ 'already-member': isAlreadyMember(user.username) }">
        <input
          :id="user.userId"
          v-model="selectedUsers"
          type="checkbox"
          :value="user.username"
          class="selected"
          :disabled="isAlreadyMember(user.username)"
          :style="isAlreadyMember(user.username) ? `cursor : not-allowed` : `cursor: pointer;`"
        > <!--se è già membro non lo posso riselezionale-deselezionare -->
        <img
          :src="`${BASE_URL()}/file?file=${user.avatar.url}`"
          alt="User Photo"
          :class="['rounded-circle','avatar','mx-2']"
          width="50"
          height="50"
        >
        {{ user.username }}
      </div>
    </div>
    <button :disabled="isSubmitting" class=" btn btn-success mt-2" @click="addSelectedUsers">Add Members</button>
  </div>
</template>



<style scoped>
.avatar {
	object-fit: cover;    /* taglia l’immagine mantenendo proporzioni 100x100 */
}

.selected {
	appearance: none;
	-webkit-appearance: none;
	width: 18px;
	height: 18px;
	border: 2px solid #0d6efd;
	border-radius: 4px;

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

.box{
	max-height: 35vh;
	overflow-y: auto;
}

.already-member {
	opacity: 0.6;
	cursor: not-allowed;
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
