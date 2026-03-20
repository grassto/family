package model

import "database/sql"

var ValidRelationTypes = map[string]string{
	"parent":      "父/母",
	"child":       "子/女",
	"spouse":      "配偶",
	"sibling":     "兄弟姐妹",
	"in_law":      "姻亲",
	"grandparent": "祖父母/外祖父母",
	"grandchild":  "孙/外孙",
}

// ReverseMap: 存 A→type→B 时，同时推导出 B→reverse→A
var ReverseMap = map[string]string{
	"parent":      "child",
	"child":       "parent",
	"spouse":      "spouse",
	"sibling":     "sibling",
	"in_law":      "in_law",
	"grandparent": "grandchild",
	"grandchild":  "grandparent",
}

type Relation struct {
	ID        int64  `json:"id"`
	PersonID  int64  `json:"person_id"`
	RelatedID int64  `json:"related_id"`
	Type      string `json:"type"`
	TypeLabel string `json:"type_label"`
	CreatedAt string `json:"created_at"`
}

type RelationReq struct {
	PersonID  int64  `json:"person_id" binding:"required"`
	RelatedID int64  `json:"related_id" binding:"required"`
	Type      string `json:"type" binding:"required"`
}

type PersonRelation struct {
	RelationID int64  `json:"relation_id"`
	PersonID   int64  `json:"person_id"`
	PersonName string `json:"person_name"`
	Type       string `json:"type"`
	TypeLabel  string `json:"type_label"`
}

type RelationRepo struct {
	DB *sql.DB
}

func (r *RelationRepo) Create(req RelationReq) (*Relation, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 插入正向关系
	res, err := tx.Exec(
		"INSERT INTO relation (person_id, related_id, type) VALUES (?, ?, ?)",
		req.PersonID, req.RelatedID, req.Type,
	)
	if err != nil {
		return nil, err
	}

	// 插入反向关系（如果还没存在）
	reverse := ReverseMap[req.Type]
	if reverse != "" {
		var exists int
		tx.QueryRow("SELECT COUNT(*) FROM relation WHERE person_id = ? AND related_id = ? AND type = ?",
			req.RelatedID, req.PersonID, reverse).Scan(&exists)
		if exists == 0 {
			tx.Exec("INSERT INTO relation (person_id, related_id, type) VALUES (?, ?, ?)",
				req.RelatedID, req.PersonID, reverse)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	id, _ := res.LastInsertId()
	return r.GetByID(id)
}

func (r *RelationRepo) GetByID(id int64) (*Relation, error) {
	rel := &Relation{}
	err := r.DB.QueryRow(
		"SELECT id, person_id, related_id, type, created_at FROM relation WHERE id = ?", id,
	).Scan(&rel.ID, &rel.PersonID, &rel.RelatedID, &rel.Type, &rel.CreatedAt)
	if err != nil {
		return nil, err
	}
	rel.TypeLabel = ValidRelationTypes[rel.Type]
	return rel, nil
}

func (r *RelationRepo) GetByPersonID(personID int64) ([]PersonRelation, error) {
	rows, err := r.DB.Query(`
		SELECT r.id, r.related_id, p.name, r.type
		FROM relation r
		JOIN person p ON p.id = r.related_id
		WHERE r.person_id = ?
		UNION ALL
		SELECT r.id, r.person_id, p.name, r.type
		FROM relation r
		JOIN person p ON p.id = r.person_id
		WHERE r.related_id = ? AND r.type IN ('spouse','sibling','in_law')
		ORDER BY r.type, p.name
	`, personID, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rels []PersonRelation
	for rows.Next() {
		var pr PersonRelation
		if err := rows.Scan(&pr.RelationID, &pr.PersonID, &pr.PersonName, &pr.Type); err != nil {
			return nil, err
		}
		pr.TypeLabel = ValidRelationTypes[pr.Type]
		rels = append(rels, pr)
	}
	if rels == nil {
		rels = []PersonRelation{}
	}
	return rels, nil
}

type FamilyRelation struct {
	PersonID   int64  `json:"person_id"`
	RelatedID  int64  `json:"related_id"`
	Type       string `json:"type"`
	TypeLabel  string `json:"type_label"`
}

// GetByFamilyID returns all relations for persons in a family
func (r *RelationRepo) GetByFamilyID(familyID int64) ([]FamilyRelation, error) {
	rows, err := r.DB.Query(`
		SELECT r.person_id, r.related_id, r.type
		FROM relation r
		JOIN person p1 ON p1.id = r.person_id
		JOIN person p2 ON p2.id = r.related_id
		WHERE p1.family_id = ? AND p2.family_id = ?
	`, familyID, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rels []FamilyRelation
	for rows.Next() {
		var fr FamilyRelation
		if err := rows.Scan(&fr.PersonID, &fr.RelatedID, &fr.Type); err != nil {
			return nil, err
		}
		fr.TypeLabel = ValidRelationTypes[fr.Type]
		rels = append(rels, fr)
	}
	if rels == nil {
		rels = []FamilyRelation{}
	}
	return rels, nil
}

func (r *RelationRepo) Delete(id int64) error {
	// 先查出关系信息，删除反向关系
	var personID, relatedID int64
	var relType string
	err := r.DB.QueryRow("SELECT person_id, related_id, type FROM relation WHERE id = ?", id).
		Scan(&personID, &relatedID, &relType)
	if err != nil {
		return err
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 删除正向
	tx.Exec("DELETE FROM relation WHERE id = ?", id)
	// 删除反向
	reverse := ReverseMap[relType]
	if reverse != "" {
		tx.Exec("DELETE FROM relation WHERE person_id = ? AND related_id = ? AND type = ?",
			relatedID, personID, reverse)
	}

	return tx.Commit()
}
