package model

import (
	"database/sql"
	"strings"
)

type Family struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	WebhookKey  string `json:"webhook_key,omitempty"`
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type FamilyCreateReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	WebhookKey  string `json:"webhook_key"`
}

type FamilyUpdateReq struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	WebhookKey  *string `json:"webhook_key"`
}

type FamilyRepo struct {
	DB *sql.DB
}

func (r *FamilyRepo) Create(req FamilyCreateReq) (*Family, error) {
	res, err := r.DB.Exec("INSERT INTO family (name, description, webhook_key) VALUES (?, ?, ?)", req.Name, req.Description, req.WebhookKey)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return r.GetByID(id)
}

func (r *FamilyRepo) GetByID(id int64) (*Family, error) {
	f := &Family{}
	err := r.DB.QueryRow(`SELECT f.id, f.name, COALESCE(f.description,''), COALESCE(f.webhook_key,''),
		COALESCE((SELECT COUNT(*) FROM person WHERE family_id = f.id), 0),
		f.created_at, f.updated_at FROM family f WHERE f.id = ?`, id).
		Scan(&f.ID, &f.Name, &f.Description, &f.WebhookKey, &f.MemberCount, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (r *FamilyRepo) List() ([]Family, error) {
	rows, err := r.DB.Query(`SELECT f.id, f.name, COALESCE(f.description,''), COALESCE(f.webhook_key,''),
		COALESCE((SELECT COUNT(*) FROM person WHERE family_id = f.id), 0),
		f.created_at, f.updated_at FROM family f ORDER BY f.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var families []Family
	for rows.Next() {
		var f Family
		if err := rows.Scan(&f.ID, &f.Name, &f.Description, &f.WebhookKey, &f.MemberCount, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		families = append(families, f)
	}
	if families == nil {
		families = []Family{}
	}
	return families, nil
}

func (r *FamilyRepo) Update(id int64, req FamilyUpdateReq) (*Family, error) {
	sets := []string{}
	args := []interface{}{}

	if req.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *req.Description)
	}
	if req.WebhookKey != nil {
		sets = append(sets, "webhook_key = ?")
		args = append(args, *req.WebhookKey)
	}

	if len(sets) > 0 {
		sets = append(sets, "updated_at = datetime('now','localtime')")
		query := "UPDATE family SET " + strings.Join(sets, ", ") + " WHERE id = ?"
		args = append(args, id)
		if _, err := r.DB.Exec(query, args...); err != nil {
			return nil, err
		}
	}
	return r.GetByID(id)
}

func (r *FamilyRepo) Delete(id int64) error {
	_, err := r.DB.Exec("DELETE FROM family WHERE id = ?", id)
	return err
}

// ListWithWebhook 返回配置了 webhook 的家族
func (r *FamilyRepo) ListWithWebhook() ([]Family, error) {
	rows, err := r.DB.Query("SELECT id, name, COALESCE(description,''), COALESCE(webhook_key,''), created_at, updated_at FROM family WHERE webhook_key != '' ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var families []Family
	for rows.Next() {
		var f Family
		if err := rows.Scan(&f.ID, &f.Name, &f.Description, &f.WebhookKey, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		families = append(families, f)
	}
	if families == nil {
		families = []Family{}
	}
	return families, nil
}
