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

// StoredRelationTypes: only these can be created directly; others are derived
var StoredRelationTypes = map[string]bool{
	"parent": true,
	"child":  true,
	"spouse": true,
}

// ReverseMap: 存 A→type→B 时，同时推导出 B→reverse→A
var ReverseMap = map[string]string{
	"parent": "child",
	"child":  "parent",
	"spouse": "spouse",
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
	Derived    bool   `json:"derived,omitempty"`
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

// GetByPersonID returns stored relations + derived relations (sibling, grandparent, in_law)
func (r *RelationRepo) GetByPersonID(personID int64) ([]PersonRelation, error) {
	// 1) Get stored relations directly involving this person
	rows, err := r.DB.Query(`
		SELECT r.id, r.related_id, p.name, r.type
		FROM relation r
		JOIN person p ON p.id = r.related_id
		WHERE r.person_id = ?
		ORDER BY r.type, p.name
	`, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rels []PersonRelation
	seen := make(map[string]bool)

	for rows.Next() {
		var pr PersonRelation
		if err := rows.Scan(&pr.RelationID, &pr.PersonID, &pr.PersonName, &pr.Type); err != nil {
			return nil, err
		}
		pr.TypeLabel = ValidRelationTypes[pr.Type]
		key := pr.Type + ":" + string(rune(pr.PersonID+'0'))
		if seen[key] {
			continue
		}
		seen[key] = true
		rels = append(rels, pr)
	}

	// 2) Derive siblings: people who share a parent with personID
	siblings, _ := r.deriveSiblings(personID, seen)
	rels = append(rels, siblings...)

	// 3) Derive grandparents: parents of my parents
	grandparents, _ := r.deriveGrandparents(personID, seen)
	rels = append(rels, grandparents...)

	// 4) Derive in-laws: spouse's parents, spouse's siblings, children's spouses
	inLaws, _ := r.deriveInLaws(personID, seen)
	rels = append(rels, inLaws...)

	if rels == nil {
		rels = []PersonRelation{}
	}
	return rels, nil
}

// deriveSiblings finds people who share a parent with personID
// r1: parent → personID (type='parent', person_id=parent, related_id=personID)
// r2: parent → other (type='parent', person_id=parent, related_id=other)
func (r *RelationRepo) deriveSiblings(personID int64, seen map[string]bool) ([]PersonRelation, error) {
	rows, err := r.DB.Query(`
		SELECT DISTINCT p.id, p.name
		FROM relation r1
		JOIN relation r2 ON r1.person_id = r2.person_id AND r2.type = 'parent' AND r2.related_id != ?
		JOIN person p ON p.id = r2.related_id
		WHERE r1.related_id = ? AND r1.type = 'parent'
	`, personID, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PersonRelation
	for rows.Next() {
		var pr PersonRelation
		if err := rows.Scan(&pr.PersonID, &pr.PersonName); err != nil {
			return nil, err
		}
		key := "sibling:" + string(rune(pr.PersonID+'0'))
		if seen[key] {
			continue
		}
		seen[key] = true
		pr.Type = "sibling"
		pr.TypeLabel = ValidRelationTypes["sibling"]
		pr.Derived = true
		result = append(result, pr)
	}
	return result, nil
}

// deriveGrandparents finds parents of personID's parents
// r1: parent → personID (type='parent', person_id=parent, related_id=personID)
// r2: grandparent → parent (type='parent', person_id=grandparent, related_id=parent)
func (r *RelationRepo) deriveGrandparents(personID int64, seen map[string]bool) ([]PersonRelation, error) {
	rows, err := r.DB.Query(`
		SELECT DISTINCT gp.id, gp.name
		FROM relation r1
		JOIN relation r2 ON r2.related_id = r1.person_id AND r2.type = 'parent'
		JOIN person gp ON gp.id = r2.person_id
		WHERE r1.related_id = ? AND r1.type = 'parent'
	`, personID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PersonRelation
	for rows.Next() {
		var pr PersonRelation
		if err := rows.Scan(&pr.PersonID, &pr.PersonName); err != nil {
			return nil, err
		}
		key := "grandparent:" + string(rune(pr.PersonID+'0'))
		if seen[key] {
			continue
		}
		seen[key] = true
		pr.Type = "grandparent"
		pr.TypeLabel = ValidRelationTypes["grandparent"]
		pr.Derived = true
		result = append(result, pr)
	}
	return result, nil
}

// deriveInLaws finds:
// 1. Parents of personID's spouse
// 2. Siblings of personID's spouse (via shared parents)
// 3. Spouses of personID's children
func (r *RelationRepo) deriveInLaws(personID int64, seen map[string]bool) ([]PersonRelation, error) {
	var result []PersonRelation

	// 1) Spouse's parents
	// sp: person → spouse (type='spouse', person_id=person, related_id=spouse)
	// par_rel: spouse's_parent → spouse (type='parent', person_id=spouse's_parent, related_id=spouse)
	rows, err := r.DB.Query(`
		SELECT DISTINCT p.id, p.name
		FROM relation sp
		JOIN relation par_rel ON par_rel.related_id = sp.related_id AND par_rel.type = 'parent'
		JOIN person p ON p.id = par_rel.person_id
		WHERE sp.person_id = ? AND sp.type = 'spouse'
	`, personID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var pr PersonRelation
			if err := rows.Scan(&pr.PersonID, &pr.PersonName); err != nil {
				continue
			}
			key := "in_law:" + string(rune(pr.PersonID+'0'))
			if seen[key] {
				continue
			}
			seen[key] = true
			pr.Type = "in_law"
			pr.TypeLabel = ValidRelationTypes["in_law"]
			pr.Derived = true
			result = append(result, pr)
		}
	}

	// 2) Spouse's siblings (people who share parents with the spouse, excluding spouse and personID)
	rows2, err := r.DB.Query(`
		SELECT DISTINCT p.id, p.name
		FROM relation sp
		JOIN relation par_rel ON par_rel.related_id = sp.related_id AND par_rel.type = 'parent'
		JOIN relation sib_rel ON sib_rel.person_id = par_rel.person_id AND sib_rel.type = 'parent' AND sib_rel.related_id != sp.related_id
		JOIN person p ON p.id = sib_rel.related_id
		WHERE sp.person_id = ? AND sp.type = 'spouse' AND sib_rel.related_id != ?
	`, personID, personID)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var pr PersonRelation
			if err := rows2.Scan(&pr.PersonID, &pr.PersonName); err != nil {
				continue
			}
			key := "in_law:" + string(rune(pr.PersonID+'0'))
			if seen[key] {
				continue
			}
			seen[key] = true
			pr.Type = "in_law"
			pr.TypeLabel = ValidRelationTypes["in_law"]
			pr.Derived = true
			result = append(result, pr)
		}
	}

	// 3) Children's spouses (spouse of a child of personID, excluding personID's spouse)
	// par_rel: person → child (type='parent', person_id=person, related_id=child)
	// sp_rel: child → their_spouse (type='spouse', person_id=child, related_id=their_spouse)
	rows3, err := r.DB.Query(`
		SELECT DISTINCT p.id, p.name
		FROM relation par_rel
		JOIN relation sp_rel ON sp_rel.person_id = par_rel.related_id AND sp_rel.type = 'spouse' AND sp_rel.related_id != ?
		JOIN person p ON p.id = sp_rel.related_id
		WHERE par_rel.person_id = ? AND par_rel.type = 'parent'
	`, personID, personID)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var pr PersonRelation
			if err := rows3.Scan(&pr.PersonID, &pr.PersonName); err != nil {
				continue
			}
			key := "in_law:" + string(rune(pr.PersonID+'0'))
			if seen[key] {
				continue
			}
			seen[key] = true
			pr.Type = "in_law"
			pr.TypeLabel = ValidRelationTypes["in_law"]
			pr.Derived = true
			result = append(result, pr)
		}
	}

	return result, nil
}

type FamilyRelation struct {
	PersonID  int64  `json:"person_id"`
	RelatedID int64  `json:"related_id"`
	Type      string `json:"type"`
	TypeLabel string `json:"type_label"`
	Derived   bool   `json:"derived,omitempty"`
}

// GetByFamilyID returns stored + derived relations for a family
func (r *RelationRepo) GetByFamilyID(familyID int64) ([]FamilyRelation, error) {
	// 1) Get stored relations
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
	seen := make(map[string]bool)

	for rows.Next() {
		var fr FamilyRelation
		if err := rows.Scan(&fr.PersonID, &fr.RelatedID, &fr.Type); err != nil {
			return nil, err
		}
		fr.TypeLabel = ValidRelationTypes[fr.Type]
		key := fr.Type + ":" + string(rune(fr.PersonID+'0')) + ":" + string(rune(fr.RelatedID+'0'))
		seen[key] = true
		rels = append(rels, fr)
	}

	// 2) Derive siblings within this family (people who share a parent)
	sibRows, err := r.DB.Query(`
		SELECT DISTINCT r1.related_id, r2.related_id
		FROM relation r1
		JOIN relation r2 ON r1.person_id = r2.person_id AND r1.type = 'parent' AND r2.type = 'parent' AND r1.related_id != r2.related_id
		JOIN person p1 ON p1.id = r1.related_id
		JOIN person p2 ON p2.id = r2.related_id
		WHERE p1.family_id = ? AND p2.family_id = ?
	`, familyID, familyID)
	if err == nil {
		defer sibRows.Close()
		for sibRows.Next() {
			var a, b int64
			if err := sibRows.Scan(&a, &b); err != nil {
				continue
			}
			// Always put smaller ID first to avoid duplicates
			if a > b {
				a, b = b, a
			}
			key := "sibling:" + string(rune(a+'0')) + ":" + string(rune(b+'0'))
			if seen[key] {
				continue
			}
			seen[key] = true
			rels = append(rels, FamilyRelation{
				PersonID: a, RelatedID: b,
				Type: "sibling", TypeLabel: ValidRelationTypes["sibling"], Derived: true,
			})
		}
	}

	// 3) Derive grandparents within this family
	// r1: parent → child (type='parent')
	// r2: grandparent → parent (type='parent')
	gpRows, err := r.DB.Query(`
		SELECT DISTINCT r2.person_id, r1.related_id
		FROM relation r1
		JOIN relation r2 ON r2.related_id = r1.person_id AND r2.type = 'parent'
		JOIN person p1 ON p1.id = r1.related_id
		JOIN person p2 ON p2.id = r2.person_id
		WHERE r1.type = 'parent' AND p1.family_id = ? AND p2.family_id = ?
	`, familyID, familyID)
	if err == nil {
		defer gpRows.Close()
		for gpRows.Next() {
			var gp, child int64
			if err := gpRows.Scan(&gp, &child); err != nil {
				continue
			}
			key := "grandparent:" + string(rune(gp+'0')) + ":" + string(rune(child+'0'))
			if seen[key] {
				continue
			}
			seen[key] = true
			rels = append(rels, FamilyRelation{
				PersonID: gp, RelatedID: child,
				Type: "grandparent", TypeLabel: ValidRelationTypes["grandparent"], Derived: true,
			})
			rels = append(rels, FamilyRelation{
				PersonID: child, RelatedID: gp,
				Type: "grandchild", TypeLabel: ValidRelationTypes["grandchild"], Derived: true,
			})
		}
	}

	// 4) Derive in-laws within this family
	// Spouse's parents
	// sp: person → spouse (type='spouse')
	// par_rel: spouse's_parent → spouse (type='parent')
	ilRows, err := r.DB.Query(`
		SELECT DISTINCT par_rel.person_id, sp.person_id
		FROM relation sp
		JOIN relation par_rel ON par_rel.related_id = sp.related_id AND par_rel.type = 'parent'
		JOIN person p1 ON p1.id = sp.person_id
		JOIN person p2 ON p2.id = par_rel.person_id
		WHERE sp.type = 'spouse' AND p1.family_id = ? AND p2.family_id = ?
	`, familyID, familyID)
	if err == nil {
		defer ilRows.Close()
		for ilRows.Next() {
			var inlaw, person int64
			if err := ilRows.Scan(&inlaw, &person); err != nil {
				continue
			}
			key := "in_law:" + string(rune(inlaw+'0')) + ":" + string(rune(person+'0'))
			if seen[key] {
				continue
			}
			seen[key] = true
			rels = append(rels, FamilyRelation{
				PersonID: inlaw, RelatedID: person,
				Type: "in_law", TypeLabel: ValidRelationTypes["in_law"], Derived: true,
			})
			rels = append(rels, FamilyRelation{
				PersonID: person, RelatedID: inlaw,
				Type: "in_law", TypeLabel: ValidRelationTypes["in_law"], Derived: true,
			})
		}
	}

	// Children's spouses
	// par_rel: parent → child (type='parent')
	// sp_rel: child → their_spouse (type='spouse')
	ilRows2, err := r.DB.Query(`
		SELECT DISTINCT sp_rel.related_id, par_rel.person_id
		FROM relation par_rel
		JOIN relation sp_rel ON sp_rel.person_id = par_rel.related_id AND sp_rel.type = 'spouse' AND sp_rel.related_id != par_rel.person_id
		JOIN person p1 ON p1.id = par_rel.person_id
		JOIN person p2 ON p2.id = sp_rel.related_id
		WHERE par_rel.type = 'parent' AND p1.family_id = ? AND p2.family_id = ?
	`, familyID, familyID)
	if err == nil {
		defer ilRows2.Close()
		for ilRows2.Next() {
			var inlaw, parent int64
			if err := ilRows2.Scan(&inlaw, &parent); err != nil {
				continue
			}
			key := "in_law:" + string(rune(inlaw+'0')) + ":" + string(rune(parent+'0'))
			if seen[key] {
				continue
			}
			seen[key] = true
			rels = append(rels, FamilyRelation{
				PersonID: inlaw, RelatedID: parent,
				Type: "in_law", TypeLabel: ValidRelationTypes["in_law"], Derived: true,
			})
			rels = append(rels, FamilyRelation{
				PersonID: parent, RelatedID: inlaw,
				Type: "in_law", TypeLabel: ValidRelationTypes["in_law"], Derived: true,
			})
		}
	}

	if rels == nil {
		rels = []FamilyRelation{}
	}
	return rels, nil
}

func (r *RelationRepo) Update(id int64, req RelationReq) (*Relation, error) {
	// 查出旧关系信息
	var oldPersonID, oldRelatedID int64
	var oldType string
	err := r.DB.QueryRow("SELECT person_id, related_id, type FROM relation WHERE id = ?", id).
		Scan(&oldPersonID, &oldRelatedID, &oldType)
	if err != nil {
		return nil, err
	}

	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 更新正向关系
	_, err = tx.Exec("UPDATE relation SET person_id = ?, related_id = ?, type = ? WHERE id = ?",
		req.PersonID, req.RelatedID, req.Type, id)
	if err != nil {
		return nil, err
	}

	// 删除旧反向关系
	oldReverse := ReverseMap[oldType]
	if oldReverse != "" {
		tx.Exec("DELETE FROM relation WHERE person_id = ? AND related_id = ? AND type = ?",
			oldRelatedID, oldPersonID, oldReverse)
	}

	// 创建新反向关系
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
	return r.GetByID(id)
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
