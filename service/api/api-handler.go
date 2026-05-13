package api

import (
	"net/http"
)

type ConversationsResponse struct {
	Conversations []Conversation `json:"conversations"`
}
type GroupsResponse struct {
	Groups []Group `json:"groups"`
}
type UsersResponse struct {
	Users []Info `json:"users"` // o []Info a seconda di cosa usi
}
type MessagesResponse struct {
	Messages []Message `json:"messages"`
}
type CommentsResponse struct {
	Comments []Comment `json:"comments"`
}

// Handler returns an instance of httprouter.Router that handle APIs registered here
func (rt *_router) Handler() http.Handler {
	// Register routes
	rt.router.GET("/", rt.getHelloWorld)
	rt.router.GET("/context", rt.wrap(rt.getContextReply))

	// paths
	rt.router.POST("/session", rt.wrap(rt.postSession)) // daje roma
	rt.router.GET("/users", rt.wrap(rt.searchUser))
	rt.router.PUT("/users/:userId/info/username", rt.wrap(rt.authHandler(rt.setMyUserName)))
	rt.router.PUT("/users/:userId/info/photo", rt.wrap(rt.authHandler(rt.setMyPhoto)))
	rt.router.GET("/users/:userId/conversations", rt.wrap(rt.authHandler(rt.getConversations)))
	rt.router.POST("/users/:userId/conversations", rt.wrap(rt.authHandler(rt.createConversation)))
	rt.router.GET("/users/:userId/conversations/:conversationId", rt.wrap(rt.authHandler(rt.getConversation)))
	rt.router.GET("/users/:userId/conversations/:conversationId/messages", rt.wrap(rt.authHandler(rt.getMessages)))
	rt.router.POST("/users/:userId/conversations/:conversationId/messages", rt.wrap(rt.authHandler(rt.postMessage)))
	rt.router.DELETE("/users/:userId/conversations/:conversationId/messages/:messageId", rt.wrap(rt.authHandler(rt.deleteMessage)))
	rt.router.POST("/users/:userId/conversations/:conversationId/messages/:messageId/forward_message", rt.wrap(rt.authHandler(rt.forwardMessage)))
	rt.router.GET("/users/:userId/conversations/:conversationId/messages/:messageId/comments", rt.wrap(rt.authHandler(rt.getComments)))
	rt.router.POST("/users/:userId/conversations/:conversationId/messages/:messageId/comments", rt.wrap(rt.authHandler(rt.postComment)))
	rt.router.DELETE("/users/:userId/conversations/:conversationId/messages/:messageId/comments/:commentId", rt.wrap(rt.authHandler(rt.unComment)))
	rt.router.GET("/users/:userId/groups", rt.wrap(rt.authHandler(rt.viewGroups)))
	rt.router.POST("/users/:userId/groups", rt.wrap(rt.authHandler(rt.createGroup)))
	rt.router.GET("/users/:userId/groups/:groupId", rt.wrap(rt.authHandler(rt.getGroup)))
	rt.router.PUT("/users/:userId/groups/:groupId/info/photo", rt.wrap(rt.authHandler(rt.setGroupPhoto)))
	rt.router.PUT("/users/:userId/groups/:groupId/info/name", rt.wrap(rt.authHandler(rt.setGroupName)))
	rt.router.POST("/users/:userId/groups/:groupId/members", rt.wrap(rt.authHandler(rt.addToGroup)))
	rt.router.GET("/users/:userId/groups/:groupId/members", rt.wrap(rt.authHandler(rt.getMembers)))
	rt.router.DELETE("/users/:userId/groups/:groupId/members/me", rt.wrap(rt.authHandler(rt.leaveGroup)))
	rt.router.POST("/users/:userId/groups/:groupId/messages", rt.wrap(rt.authHandler(rt.postGroupMessage)))
	rt.router.GET("/users/:userId/groups/:groupId/messages", rt.wrap(rt.authHandler(rt.getGroupMessages)))
	rt.router.DELETE("/users/:userId/groups/:groupId/messages/:messageId", rt.wrap(rt.authHandler(rt.deleteGroupMessage)))
	rt.router.POST("/users/:userId/groups/:groupId/messages/:messageId/forward_message", rt.wrap(rt.authHandler(rt.forwardGroupMessage)))
	rt.router.GET("/users/:userId/groups/:groupId/messages/:messageId/comments", rt.wrap(rt.authHandler(rt.getGroupComments)))
	rt.router.POST("/users/:userId/groups/:groupId/messages/:messageId/comments", rt.wrap(rt.authHandler(rt.postGroupComment)))
	rt.router.DELETE("/users/:userId/groups/:groupId/messages/:messageId/comments/:commentId", rt.wrap(rt.authHandler(rt.unGroupComment)))
	rt.router.GET("/file", rt.wrap(rt.getFile))

	// Special routes
	rt.router.GET("/liveness", rt.liveness)

	return rt.router
}
