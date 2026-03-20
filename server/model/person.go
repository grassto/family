package model

import "database/sql"

type Person struct {
	ID         int64  `json:"id"`
	FamilyID   int64  `json:"family_id"`
	Name       string `json:"name"`
	Gender     string `json:"gender"`
	Birthday   string `json:"birthday,omitempty"`
	Generation *int   `json:"generation,omitempty"`
	PhotoURL   string `json:"photo_url,omitempty"`
	Phone      string `json:"phone,omitempty"`
	Address    string `json:"address,omitempty"`
	Notes      string `json:"notes,omitempty"`
	IsAlive    bool   `json:"is_alive"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

type PersonCreateReq struct {
	FamilyID   int64  `json:"family_id" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Gender     string `json:"gender"`
	Birthday   string `json:"birthday"`
	Generation *int   `json:"generation"`
	PhotoURL   string `json:"photo_url"`
	Phone      string `json:"phone"`
	Address    string `json:"address"`
	Notes      string `json:"notes"`
}

type PersonUpdateReq struct {
	Name       *string `json:"name"`
	Gender     *string `json:"gender"`
	Birthday   *string `json:"birthday"`
	Generation *int    `json:"generation"`
	PhotoURL   *string `json:"photo_url"`
	Phone      *string `json:"phone"`
	Address    *string `json:"address"`
	Notes      *string `json:"notes"`
	IsAlive    *bool   `json:"is_alive"`
}

type PersonRepo struct {
	DB *sql.DB
}

func (r *PersonRepo) Create(req PersonCreateReq) (*Person, error) {
	gender := req.Gender
	if gender == "" {
		gender = "unknown"
	}
	res, err := r.DB.Exec(
		`INSERT INTO person (family_id, name, gender, birthday, generation, photo_url, phone, address, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.FamilyID, req.Name, gender, req.Birthday, req.Generation, req.PhotoURL, req.Phone, req.Address, req.Notes,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return r.GetByID(id)
}

func (r *PersonRepo) GetByID(id int64) (*Person, error) {
	p := &Person{}
	var gen sql.NullInt64
	var birthday, photoURL, phone, address, notes sql.NullString
	err := r.DB.QueryRow(
		`SELECT id, family_id, name, gender, COALESCE(birthday,''), generation,
		        COALESCE(photo_url,''), COALESCE(phone,''), COALESCE(address,''), COALESCE(notes,''),
		        is_alive, created_at, updated_at
		 FROM person WHERE id = ?`, id,
	).Scan(&p.ID, &p.FamilyID, &p.Name, &p.Gender, &birthday, &gen,
		&photoURL, &phone, &address, &notes, &p.IsAlive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.Birthday = birthday.String
	p.PhotoURL = photoURL.String
	p.Phone = phone.String
	p.Address = address.String
	p.Notes = notes.String
	if gen.Valid {
		g := int(gen.Int64)
		p.Generation = &g
	}
	return p, nil
}

func (r *PersonRepo) ListByFamily(familyID int64, keyword string) ([]Person, error) {
	query := `SELECT id, family_id, name, gender, COALESCE(birthday,''), generation,
	          COALESCE(photo_url,''), COALESCE(phone,''), COALESCE(address,''), COALESCE(notes,''),
	          is_alive, created_at, updated_at
	          FROM person WHERE family_id = ?`
	args := []interface{}{familyID}

	if keyword != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}
	query += " ORDER BY generation, birthday"

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []Person
	for rows.Next() {
		var p Person
		var gen sql.NullInt64
		var birthday, photoURL, phone, address, notes sql.NullString
		if err := rows.Scan(&p.ID, &p.FamilyID, &p.Name, &p.Gender, &birthday, &gen,
			&photoURL, &phone, &address, &notes, &p.IsAlive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Birthday = birthday.String
		p.PhotoURL = photoURL.String
		p.Phone = phone.String
		p.Address = address.String
		p.Notes = notes.String
		if gen.Valid {
			g := int(gen.Int64)
			p.Generation = &g
		}
		persons = append(persons, p)
	}
	if persons == nil {
		persons = []Person{}
	}
	return persons, nil
}

func (r *PersonRepo) Update(id int64, req PersonUpdateReq) (*Person, error) {
	sets := []string{}
	args := []interface{}{}

	if req.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Gender != nil {
		sets = append(sets, "gender = ?")
		args = append(args, *req.Gender)
	}
	if req.Birthday != nil {
		sets = append(sets, "birthday = ?")
		args = append(args, *req.Birthday)
	}
	if req.Generation != nil {
		sets = append(sets, "generation = ?")
		args = append(args, *req.Generation)
	}
	if req.PhotoURL != nil {
		sets = append(sets, "photo_url = ?")
		args = append(args, *req.PhotoURL)
	}
	if req.Phone != nil {
		sets = append(sets, "phone = ?")
		args = append(args, *req.Phone)
	}
	if req.Address != nil {
		sets = append(sets, "address = ?")
		args = append(args, *req.Address)
	}
	if req.Notes != nil {
		sets = append(sets, "notes = ?")
		args = append(args, *req.Notes)
	}
	if req.IsAlive != nil {
		sets = append(sets, "is_alive = ?")
		args = append(args, *req.IsAlive)
	}

	if len(sets) > 0 {
		sets = append(sets, "updated_at = datetime('now','localtime')")
		query := "UPDATE person SET " + joinStrings(sets, ", ") + " WHERE id = ?"
		args = append(args, id)
		if _, err := r.DB.Exec(query, args...); err != nil {
			return nil, err
		}
	}
	return r.GetByID(id)
}

func (r *PersonRepo) Delete(id int64) error {
	_, err := r.DB.Exec("DELETE FROM person WHERE id = ?", id)
	return err
}

func (r *PersonRepo) GetBirthdayUpcoming(days int) ([]Person, error) {
	query := `
		SELECT id, family_id, name, gender, birthday, generation,
		       COALESCE(photo_url,''), COALESCE(phone,''), COALESCE(address,''), COALESCE(notes,''),
		       is_alive, created_at, updated_at
		FROM person
		WHERE birthday IS NOT NULL AND birthday != '' AND is_alive = 1
		  AND (
		    -- 今年还没过
		    (substr(birthday,6,5) >= strftime('%m-%d','now') AND
		     substr(birthday,6,5) <= strftime('%m-%d','now','+' || ? || ' days'))
		    OR
		    -- 跨年：现在年底，生日在明年初
		    (strftime('%m-%d','now','+' || ? || ' days') < strftime('%m-%d','now') AND
		     (substr(birthday,6,5) >= strftime('%m-%d','now') OR
		      substr(birthday,6,5) <= strftime('%m-%d','now','+' || ? || ' days')))
		  )
		ORDER BY substr(birthday,6,5)`

	rows, err := r.DB.Query(query, days, days, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []Person
	for rows.Next() {
		var p Person
		var gen sql.NullInt64
		var birthday, photoURL, phone, address, notes sql.NullString
		if err := rows.Scan(&p.ID, &p.FamilyID, &p.Name, &p.Gender, &birthday, &gen,
			&photoURL, &phone, &address, &notes, &p.IsAlive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Birthday = birthday.String
		p.PhotoURL = photoURL.String
		p.Phone = phone.String
		p.Address = address.String
		p.Notes = notes.String
		if gen.Valid {
			g := int(gen.Int64)
			p.Generation = &g
		}
		persons = append(persons, p)
	}
	if persons == nil {
		persons = []Person{}
	}
	return persons, nil
}

func (r *PersonRepo) GetBirthdayToday() ([]Person, error) {
	return r.GetBirthdayUpcoming(0)
}

// GetBirthdayTodayByFamily 查指定家族今天生日的人
func (r *PersonRepo) GetBirthdayTodayByFamily(familyID int64) ([]Person, error) {
	query := `
		SELECT id, family_id, name, gender, birthday, generation,
		       COALESCE(photo_url,''), COALESCE(phone,''), COALESCE(address,''), COALESCE(notes,''),
		       is_alive, created_at, updated_at
		FROM person
		WHERE birthday IS NOT NULL AND birthday != '' AND is_alive = 1
		  AND family_id = ?
		  AND substr(birthday,6,5) = strftime('%m-%d','now')
		ORDER BY name`

	rows, err := r.DB.Query(query, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []Person
	for rows.Next() {
		var p Person
		var gen sql.NullInt64
		var birthday, photoURL, phone, address, notes sql.NullString
		if err := rows.Scan(&p.ID, &p.FamilyID, &p.Name, &p.Gender, &birthday, &gen,
			&photoURL, &phone, &address, &notes, &p.IsAlive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Birthday = birthday.String
		p.PhotoURL = photoURL.String
		p.Phone = phone.String
		p.Address = address.String
		p.Notes = notes.String
		if gen.Valid {
			g := int(gen.Int64)
			p.Generation = &g
		}
		persons = append(persons, p)
	}
	if persons == nil {
		persons = []Person{}
	}
	return persons, nil
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
