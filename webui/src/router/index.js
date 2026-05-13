import {createRouter, createWebHashHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginView from '../views/LoginView.vue'
import ChatView from '../views/ChatView.vue'

import GroupView from "../views/GroupView.vue";

const router = createRouter({
	history: createWebHashHistory(import.meta.env.BASE_URL),
	routes: [
		{path: '/', component: LoginView},
		{path: '/home', component: HomeView},
		{	path: '/conversation/:conversationId',
			name: 'conversation',
			component: ChatView},
		{	path: '/group/:groupId',
			name: 'group',
			component: GroupView},
	]
})

export default router
