package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"family-tree/model"
	"family-tree/service"

	"github.com/gin-gonic/gin"
)

type FamilyHandler struct {
	Repo        *model.FamilyRepo
	TransferSvc *service.FamilyTransferService
}

func (h *FamilyHandler) Register(r *gin.RouterGroup) {
	r.POST("/families", h.Create)
	r.GET("/families", h.List)
	r.GET("/families/:id", h.Get)
	r.PUT("/families/:id", h.Update)
	r.DELETE("/families/:id", h.Delete)
	r.GET("/families/:id/export", h.Export)
	r.POST("/families/import", h.Import)
}

func (h *FamilyHandler) Create(c *gin.Context) {
	var req model.FamilyCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f, err := h.Repo.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, f)
}

func (h *FamilyHandler) List(c *gin.Context) {
	families, err := h.Repo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, families)
}

func (h *FamilyHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	f, err := h.Repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "family not found"})
		return
	}
	c.JSON(http.StatusOK, f)
}

func (h *FamilyHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.FamilyUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	f, err := h.Repo.Update(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, f)
}

func (h *FamilyHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.Repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *FamilyHandler) Export(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	payload, err := h.TransferSvc.ExportFamily(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "family not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payload)
}

func (h *FamilyHandler) Import(c *gin.Context) {
	var payload service.FamilyTransferPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	family, err := h.TransferSvc.ImportFamily(payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, family)
}
