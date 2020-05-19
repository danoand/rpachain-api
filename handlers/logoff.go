package handlers

import (
	"net/http"

	ginsession "github.com/go-session/gin-session"

	"github.com/gin-gonic/gin"
)

// Logoff clears a session and effectively logs off a user
func (hlr *HandlerEnv) Logoff(c *gin.Context) {
	ginsession.Destroy(c)

	var rsp = make(map[string]string)
	rsp["msg"] = "you are no logged off"

	c.JSON(http.StatusOK, rsp)
}
