<script>

import {BASE_URL, getGroup, getUserInfo} from "../services/axios.js";
import router from "../router";

export default {
	props: {
		show: Boolean,
		groupId: Number,
		userId: Number
	},
	emits: ['close'],
	data() {
		return {
			group: null,
			users: [],
			destroyed: false
		};
	},
	watch: {
		show(newVal) {
			if (newVal) {
				this.refresh();
			}
		}
	},
	beforeUnmount() {
		this.destroyed = true;
	},
	methods: {
		router(){
			return router;
		},
		BASE_URL(){
			return BASE_URL;
		},
		async refresh() {
			if (!this.groupId) return;

			try {
				const group = await getGroup(this.userId, this.groupId);
				if (this.destroyed) return; // 🔥 BLOCCA UPDATE

				this.group = group;

				const users = await Promise.all(
					group.participants.map(async user => {
						const userInfo = await getUserInfo(String(user.userId));
						return { ...user, avatar: userInfo.avatar };
					})
				);

				if (this.destroyed) return;
				this.users = users;

			} catch (e) {
				console.error(e);
			}
		},
		goTo(){
			router.push({ name: 'group', params: { groupId: this.groupId } });
		}
	}
};
</script>

<template>
  <div v-if="show" class="overlay">
    <div class="action-box">
      <div class="d-flex align-items-center justify-content-between mb-2">
        <button class="btn-close" @click="$emit('close')" />
        <h4 v-if="group" class="text-center">
          {{ group.name }} - Participants
        </h4>
        <br>
      </div>

      <div class="d-flex flex-column gap-3 mb-3">
        <div v-for="user in users" :key="user.userId" class="d-flex align-items-center gap-2">
          <img class="avatar rounded-circle" :src="`${BASE_URL()}/file?file=${user.avatar.url}`" width="50" height="50" alt="Photo">
          <span>{{ user.username }}</span>
        </div>
      </div>
      <div class="d-flex justify-content-end">
        <button class="btn btn-outline-primary justify-content-end" @click="goTo">Open</button>
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

.avatar {
	object-fit: cover;    /* taglia l’immagine mantenendo proporzioni 100x100 */
}

</style>
