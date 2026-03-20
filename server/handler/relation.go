package handler

import (
	"net/http"
	"strconv"

	"family-tree/model"

	"github.com/gin-gonic/gin"
)

type RelationHandler struct {
	Repo *model.RelationRepo
}

func (h *RelationHandler) Register(r *gin.RouterGroup) {
	r.POST("/relations", h.Create)
	r.GET("/persons/:id/relations", h.ListByPerson)
	r.GET("/families/:id/relations", h.ListByFamily)
	r.DELETE("/relations/:id", h.Delete)
	r.GET("/relation-types", h.ListTypes)
}

func (h *RelationHandler) Create(c *gin.Context) {
	var req model.RelationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.PersonID == req.RelatedID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot create relation with self"})
		return
	}
	if _, ok := model.ValidRelationTypes[req.Type]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid relation type"})
		return
	}
	rel, err := h.Repo.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rel)
}

func (h *RelationHandler) ListByPerson(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rels, err := h.Repo.GetByPersonID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rels)
}

func (h *RelationHandler) ListByFamily(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	rels, err := h.Repo.GetByFamilyID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rels)
}

func (h *RelationHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.Repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *RelationHandler) ListTypes(c *gin.Context) {
	types := make([]map[string]string, 0, len(model.ValidRelationTypes))
	for k, v := range model.ValidRelationTypes {
		types = append(types, map[string]string{"value": k, "label": v})
	}
	c.JSON(http.StatusOK, types)
}
