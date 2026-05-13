import axios from "axios";

const instance = axios.create({
	baseURL: __API_URL__,
	timeout: 1000 * 5
});


export const BASE_URL = __API_URL__

instance.interceptors.request.use(config => {
	const token = sessionStorage.getItem("token");
	if (token) {
		config.headers.Authorization = `Bearer ${token}`;
	}

	return config;
});
// modificare questa cacata
export const doLogin = async (username, password) => {
	try {
		const response = await instance.post("/session", {
			name: username,
			password: password
		});

		return {
			userId: response.data.userId,
			token: response.data.token,
			time: response.data.time
		};
	} catch (error) {

    console.error("Login error:", error);
    if (error.response?.data?.error) {
      throw new Error(error.response.data.error);
    }
    throw new Error("Errore durante il login");

  }
};


export const getConversations = async (userId) => {
	try{
		const response = await instance.get(`users/${userId}/conversations`);
		return response.data.conversations;
	} catch (error) {
		console.error("Get conversations error:", error);
		throw error;
	}
}



export const getGroups = async  (userId) => {
	try{
		const response = await instance.get(`users/${userId}/groups`);
		return response.data.groups;
	} catch (error) {
		console.error("Get conversations error:", error);
		throw error;
	}
}


export const getMessages = async (userId,chatId,originType) => {
	let url;
	if (originType === "direct") {
		url = `users/${userId}/conversations/${chatId}/messages`;
	} else {
		url = `users/${userId}/groups/${chatId}/messages`;
	}

	try{
		const response = await instance.get(url);
		return response.data.messages;
	}catch(error){
		console.error("Get conversations error:", error);
		throw error;
	}
}

export const getConversation = async (userId,chatId,originType) => {
	let url;
	if (originType === "direct") {
		url = `users/${userId}/conversations/${chatId}`;
	} else {
		url = `users/${userId}/groups/${chatId}`;
	}
	try{
		const response = await instance.get(url);
		return response.data;
	}catch(error){
		console.error("Get conversation error:", error);
		throw error;
	}
};


export const postMessage = async (userId,chatId,messageText,messagePhoto,type,replyTo) => {
	try{
		const query = replyTo ? `?replyTo=${replyTo}` : "";
		let response;
		if (messagePhoto) {
			const formData = new FormData();
			formData.append("bodyText",messageText);
			formData.append("photo",messagePhoto);
			if(type === "direct") {
				response = await instance.post(`users/${userId}/conversations/${chatId}/messages${query}`, formData,{headers:{"Content-Type": "multipart/form-data"}});
			} else if(type === "group") {
				response = await instance.post(`users/${userId}/groups/${chatId}/messages${query}`, formData,{headers:{"Content-Type": "multipart/form-data"}});
			}
		}else{
			if(type === "direct") {
				response = await instance.post(`users/${userId}/conversations/${chatId}/messages${query}`, {bodyText: messageText});
			} else if(type === "group") {
				response = await instance.post(`users/${userId}/groups/${chatId}/messages${query}`,{bodyText: messageText});
			}
		}
		return response.data;
	}catch(error){
		console.error("PostMessage error:", error);
		throw error;
	}
}

export const forwardMessage = async (
	userId,
	originChatId,
	messageId,
	originType, // "direct" | "group"
	forwardToConversation,
	forwardToGroup
) => {
	try {
		const body = {
			forwardToConversation,
			forwardToGroup
		};

		let url;
		if (originType === "direct") {
			url = `users/${userId}/conversations/${originChatId}/messages/${messageId}/forward_message`;
		} else {
			url = `users/${userId}/groups/${originChatId}/messages/${messageId}/forward_message`;
		}

		const response = await instance.post(url, body);
		return response.data;

	} catch (error) {
		console.error("ForwardMessage error:", error);
		throw error;
	}
};



export const deleteMessage = async (userId,conversationId,messageId,type) =>{
	try{
		let url;
		if (type === "direct") {
			url = `users/${userId}/conversations/${conversationId}/messages/${messageId}`;
		} else {
			url = `users/${userId}/groups/${conversationId}/messages/${messageId}`;
		}

		const response = await instance.delete(url);
		return response.data;
	}catch (error) {
		console.error("DeleteMessage error:", error);
		throw error;
	}
}


export const commentMessage = async (userId,conversationId,messageId,type,emoji) =>{
	try{
		let url;
		if (type === "direct") {
			url = `users/${userId}/conversations/${conversationId}/messages/${messageId}/comments`;
		} else {
			url = `users/${userId}/groups/${conversationId}/messages/${messageId}/comments`;
		}
		const response = await instance.post(url, {emoji:emoji});
		return response.data;
	}catch(error){
		console.error("CommentMessage error:", error);
		throw error;
	}
}

export const getComments = async (userId,conversationId,messageId,type) =>{
	try{
		let url;
		if (type === "direct") {
			url = `users/${userId}/conversations/${conversationId}/messages/${messageId}/comments`;
		} else {
			url = `users/${userId}/groups/${conversationId}/messages/${messageId}/comments`;
		}
		const response = await instance.get(url);
		return response.data.comments;
	}catch(error){
		console.error("CommentMessage error:", error);
		throw error;
	}
}

export const unComment = async (userId,conversationId,messageId,commentId,type) =>{
	try{
		let url;
		if (type === "direct") {
			url = `users/${userId}/conversations/${conversationId}/messages/${messageId}/comments/${commentId}`;
		} else {
			url = `users/${userId}/groups/${conversationId}/messages/${messageId}/comments/${commentId}`;
		}
		const response = await instance.delete(url);
		return response.data;
	}catch(error){
		console.error("CommentMessage error:", error);
		throw error;
	}
}

export const getUsers = async () =>{
	try{
		const response= await instance.get("users");
		return response.data.users;
	}catch(error){
		console.error("GetUserss error:", error);
		throw error;
	}
}

export const createConversation = async (userId,username) =>{
	try{
		const response = await instance.post(`users/${userId}/conversations`, {name: username});
		return response.data;
	}catch(error){
		console.error("CreateConversation error:", error);
		throw error;
	}
}



export const createGroup = async (userId,participants,name) =>{
	try{
		const response = await instance.post(`users/${userId}/groups`, {name: name,participants:participants});
		return response.data;
	}catch(error){
		console.error("CreateGroup error:", error);
		throw error;
	}
}


export const getUserInfo = async (userId) => {
	try {
		const response = await instance.get("users", {
			params: { userId: userId }
		}); // users?userId=userID
		if (!response.data || response.data.length === 0) {
			new Error("Utente non trovato");
		}
		return response.data.users[0];
	} catch (error) {
		console.error("GetUserInfo error:", error);
		throw error;
	}
};

export const setUsername= async (userId,username) =>{
	try{
		const response = await instance.put(`users/${userId}/info/username`, {newusername: username});
		return response.data;
	}catch(error){
		throw error;
	}
}


export const setPhotoUser = async (userId, file) => {
	try {
		const formData = new FormData();
		formData.append("photo", file);

		const response = await instance.put(
			`users/${userId}/info/photo`,
			formData,
			{
				headers: { "Content-Type": "multipart/form-data" }
			}
		);

		return response.data;
	} catch (error) {
		console.error("SetPhotoUser error:", error);
		throw error;
	}
};


export const getGroup=async (userId,groupId) =>{
	try{
		const response = await instance.get(`users/${userId}/groups/${groupId}`)
		return response.data;
	}catch(error){
		console.error("GetGroup error:", error);
		throw error;
	}
}

export const deleteFromGroup = async (userId,groupId) =>{
	try{
		const response = await instance.delete(`users/${userId}/groups/${groupId}/members/me`);
		return response.data;
	}catch(error){
		console.error("DeleteFromGroup error:", error);
		throw error;
	}
}


export const addToGroup = async (userId,groupId,username) =>{
	try{
		const response = await instance.post(`users/${userId}/groups/${groupId}/members`, {name: username});
		return response.data;
	}catch(error){
		console.error("AddToGroup error:", error);
		throw error;
	}
}


export const setGroupName = async (userId,groupId,groupName) =>{
	try{
		const response = await instance.put(`users/${userId}/groups/${groupId}/info/name`, {name: groupName});
		return response.data;
	}catch(error){
		console.error("SetGroupName error:", error);
		throw error;
	}
}


export const setGroupPhoto = async (userId,groupId,file) =>{
	try {
		const formData = new FormData();
		formData.append("photo", file);

		const response = await instance.put(
			`users/${userId}/groups/${groupId}/info/photo`,
			formData,
			{
				headers: { "Content-Type": "multipart/form-data" }
			}
		);

		return response.data;
	} catch (error) {
		console.error("SetPhotoUser error:", error);
		throw error;
	}
}





export default instance;
