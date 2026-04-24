package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"family-tree/model"
)

type FamilyTransferMeta struct {
	Version    string `json:"version"`
	ExportedAt string `json:"exported_at"`
}

type FamilyTransferPerson struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Gender       string `json:"gender"`
	Birthday     string `json:"birthday,omitempty"`
	BirthdayType string `json:"birthday_type,omitempty"`
	BirthDate    string `json:"birth_date,omitempty"`
	DeathDate    string `json:"death_date,omitempty"`
	Generation   *int   `json:"generation,omitempty"`
	PhotoURL     string `json:"photo_url,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Address      string `json:"address,omitempty"`
	Notes        string `json:"notes,omitempty"`
	IsAlive      bool   `json:"is_alive"`
}

type FamilyTransferRelation struct {
	PersonID  int64  `json:"person_id"`
	RelatedID int64  `json:"related_id"`
	Type      string `json:"type"`
}

type FamilyTransferPayload struct {
	Family    model.Family             `json:"family"`
	Persons   []FamilyTransferPerson   `json:"persons"`
	Relations []FamilyTransferRelation `json:"relations"`
	Meta      FamilyTransferMeta       `json:"meta"`
}

type FamilyTransferService struct {
	DB *sql.DB
}

func (s *FamilyTransferService) ExportFamily(familyID int64) (*FamilyTransferPayload, error) {
	familyRepo := &model.FamilyRepo{DB: s.DB}
	family, err := familyRepo.GetByID(familyID)
	if err != nil {
		return nil, err
	}

	personRows, err := s.DB.Query(`
		SELECT id, name, gender, COALESCE(birthday, ''), COALESCE(birthday_type, 'solar'),
		       COALESCE(birth_date, ''), COALESCE(death_date, ''), generation,
		       COALESCE(photo_url, ''), COALESCE(phone, ''), COALESCE(address, ''), COALESCE(notes, ''), is_alive
		FROM person
		WHERE family_id = ?
		ORDER BY id
	`, familyID)
	if err != nil {
		return nil, err
	}
	defer personRows.Close()

	persons := make([]FamilyTransferPerson, 0)
	for personRows.Next() {
		var p FamilyTransferPerson
		var generation sql.NullInt64
		if err := personRows.Scan(&p.ID, &p.Name, &p.Gender, &p.Birthday, &p.BirthdayType, &p.BirthDate, &p.DeathDate, &generation, &p.PhotoURL, &p.Phone, &p.Address, &p.Notes, &p.IsAlive); err != nil {
			return nil, err
		}
		if generation.Valid {
			g := int(generation.Int64)
			p.Generation = &g
		}
		persons = append(persons, p)
	}

	relationRows, err := s.DB.Query(`
		SELECT r.person_id, r.related_id, r.type
		FROM relation r
		JOIN person p1 ON p1.id = r.person_id
		JOIN person p2 ON p2.id = r.related_id
		WHERE p1.family_id = ? AND p2.family_id = ?
		ORDER BY r.id
	`, familyID, familyID)
	if err != nil {
		return nil, err
	}
	defer relationRows.Close()

	relations := make([]FamilyTransferRelation, 0)
	for relationRows.Next() {
		var rel FamilyTransferRelation
		if err := relationRows.Scan(&rel.PersonID, &rel.RelatedID, &rel.Type); err != nil {
			return nil, err
		}
		relations = append(relations, rel)
	}

	payload := &FamilyTransferPayload{
		Family:    *family,
		Persons:   persons,
		Relations: relations,
		Meta: FamilyTransferMeta{
			Version:    "1.0",
			ExportedAt: time.Now().Format(time.RFC3339),
		},
	}
	return payload, nil
}

func (s *FamilyTransferService) ImportFamily(payload FamilyTransferPayload) (*model.Family, error) {
	if payload.Family.Name == "" {
		return nil, errors.New("family name is required")
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	familyRes, err := tx.Exec(
		"INSERT INTO family (name, description, webhook_key) VALUES (?, ?, ?)",
		payload.Family.Name,
		payload.Family.Description,
		payload.Family.WebhookKey,
	)
	if err != nil {
		return nil, err
	}
	newFamilyID, _ := familyRes.LastInsertId()

	personIDMap := make(map[int64]int64, len(payload.Persons))
	for _, p := range payload.Persons {
		gender := p.Gender
		if gender == "" {
			gender = "unknown"
		}
		birthdayType := p.BirthdayType
		if birthdayType == "" {
			birthdayType = "solar"
		}
		res, err := tx.Exec(
			`INSERT INTO person (family_id, name, gender, birthday, birthday_type, birth_date, death_date, generation, photo_url, phone, address, notes, is_alive)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			newFamilyID,
			p.Name,
			gender,
			p.Birthday,
			birthdayType,
			p.BirthDate,
			p.DeathDate,
			p.Generation,
			p.PhotoURL,
			p.Phone,
			p.Address,
			p.Notes,
			p.IsAlive,
		)
		if err != nil {
			return nil, err
		}
		newPersonID, _ := res.LastInsertId()
		personIDMap[p.ID] = newPersonID
	}

	for _, rel := range payload.Relations {
		if !model.StoredRelationTypes[rel.Type] {
			continue
		}
		newPersonID, okA := personIDMap[rel.PersonID]
		newRelatedID, okB := personIDMap[rel.RelatedID]
		if !okA || !okB || newPersonID == newRelatedID {
			continue
		}
		if _, err := tx.Exec(
			"INSERT OR IGNORE INTO relation (person_id, related_id, type) VALUES (?, ?, ?)",
			newPersonID,
			newRelatedID,
			rel.Type,
		); err != nil {
			return nil, fmt.Errorf("insert relation failed: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return (&model.FamilyRepo{DB: s.DB}).GetByID(newFamilyID)
}
