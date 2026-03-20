package handler

import (
	"net/http"
	"strconv"

	"family-tree/model"

	"github.com/gin-gonic/gin"
)

type PersonHandler struct {
	Repo *model.PersonRepo
}

func (h *PersonHandler) Register(r *gin.RouterGroup) {
	r.POST("/persons", h.Create)
	r.GET("/persons", h.List)
	r.GET("/persons/:id", h.Get)
	r.PUT("/persons/:id", h.Update)
	r.DELETE("/persons/:id", h.Delete)
	r.GET("/birthdays/today", h.TodayBirthdays)
	r.GET("/birthdays/upcoming", h.UpcomingBirthdays)
}

func (h *PersonHandler) Create(c *gin.Context) {
	var req model.PersonCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.Repo.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *PersonHandler) List(c *gin.Context) {
	familyIDStr := c.Query("family_id")
	if familyIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "family_id is required"})
		return
	}
	familyID, err := strconv.ParseInt(familyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid family_id"})
		return
	}
	keyword := c.Query("keyword")
	persons, err := h.Repo.ListByFamily(familyID, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, persons)
}

func (h *PersonHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	p, err := h.Repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "person not found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *PersonHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.PersonUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	p, err := h.Repo.Update(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *PersonHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.Repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *PersonHandler) TodayBirthdays(c *gin.Context) {
	persons, err := h.Repo.GetBirthdayToday()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, persons)
}

func (h *PersonHandler) UpcomingBirthdays(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 {
		days = 30
	}
	persons, err := h.Repo.GetBirthdayUpcoming(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, persons)
}
