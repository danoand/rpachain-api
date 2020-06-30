package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Logoff clears a session and effectively logs off a user
func (hlr *HandlerEnv) Logoff(c *gin.Context) {

	var rsp = make(map[string]string)
	rsp["msg"] = "you are now logged off"

	c.JSON(http.StatusOK, rsp)
}
