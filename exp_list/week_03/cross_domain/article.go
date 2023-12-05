package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func CreateArticle(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var a Article
		if err := c.ShouldBindJSON(&a); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		//从 context 获取 userID
		userID := c.Value("session_id")
		a.UserId = userID.(uint)

		if res := db.Create(&a); res.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": res.Error.Error(),
			})
			return
		}
		c.JSON(http.StatusCreated, a)
	}
}

func DeleteArticle(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var a Article
		//查询文章是否存在
		if res := db.First(&a, id); res.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": res.Error.Error(),
			})
			return
		}
		if res := db.Delete(&a, id); res.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": res.Error.Error(),
			})
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func UpdateArticle(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var a Article
		//查询文章是否存在
		if res := db.First(&a, id); res.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": res.Error.Error(),
			})
			return
		}
		//解析传入的更新数据
		if err := c.ShouldBindJSON(&a); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		}
		//保存更新数据
		if res := db.Save(&a); res.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": res.Error.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, a)
	}
}

func GetArticle(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Param("id"))
		var a Article
		//查询文章是否存在
		if res := db.First(&a, id); res.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": res.Error.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, a)
	}
}

func GetAllArticles(db *gorm.DB) func(c *gin.Context) {
	return func(c *gin.Context) {
		var as []Article
		if res := db.Find(&as); res.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": res.Error.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, as)
	}
}
