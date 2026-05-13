<script>

import {BASE_URL} from "../services/axios";


export default {
	props: {
		show:Boolean,
		username: String,
		photo: String,
	},
	emits: ['close'],
	methods: {
		BASE_URL(){
			return BASE_URL;
		},
	}
}

</script>

<template>
  <div v-if="show" class="overlay">
    <div class="action-box">
      <div class="header d-flex align-items-center position-relative mb-2">
        <button class="btn btn-close" @click="$emit('close')" />
        <h4 class="header-title">Info Profile</h4>
      </div>
      <div class="d-flex align-items-center justify-content-center flex-column gap-3">
        <img class="avatar rounded-circle" :src="`${BASE_URL()}/file?file=${photo}`" width="200" height="200" alt="Photo">
        <h4 class="fw-bold text-center">{{ username }}</h4>
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

.btn{
	border-radius: 6px;
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

.avatar {
	object-fit: cover;    /* taglia l’immagine mantenendo proporzioni 100x100 */
}


</style>
