package handlers

import (
	"net/http"
	"time"

	"maplocationshare/backend/models"
	"maplocationshare/backend/storage"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var store storage.Storage

func SetStorage(s storage.Storage) {
	store = s
}

func CreateSession(c *gin.Context) {
	var req models.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	sessionID := uuid.New().String()
	location := models.Location{
		UserID:    req.UserID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Timestamp: time.Now().UnixMilli(),
	}

	if err := store.CreateSession(sessionID, location); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sessionId": sessionID,
		"message":   "会话创建成功",
	})
}

func GetSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "会话ID不能为空"})
		return
	}

	session, err := store.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	c.JSON(http.StatusOK, session)
}

func UpdateLocation(c *gin.Context) {
	var req models.UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	if !store.SessionExists(req.SessionID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "会话不存在"})
		return
	}

	location := models.Location{
		UserID:    req.UserID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Timestamp: time.Now().UnixMilli(),
	}

	if err := store.UpdateLocation(req.SessionID, location); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新位置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "位置更新成功",
	})
}
