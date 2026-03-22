package model

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"family-tree/pkg/lunar"
)

type Person struct {
	ID           int64  `json:"id"`
	FamilyID     int64  `json:"family_id"`
	Name         string `json:"name"`
	Gender       string `json:"gender"`
	Birthday     string `json:"birthday,omitempty"`
	BirthdayType string `json:"birthday_type"`
	Generation   *int   `json:"generation,omitempty"`
	PhotoURL     string `json:"photo_url,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Address      string `json:"address,omitempty"`
	Notes        string `json:"notes,omitempty"`
	IsAlive      bool   `json:"is_alive"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type PersonCreateReq struct {
	FamilyID     int64  `json:"family_id" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Gender       string `json:"gender"`
	Birthday     string `json:"birthday"`
	BirthdayType string `json:"birthday_type"`
	Generation   *int   `json:"generation"`
	PhotoURL     string `json:"photo_url"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
	Notes        string `json:"notes"`
}

type PersonUpdateReq struct {
	Name         *string `json:"name"`
	Gender       *string `json:"gender"`
	Birthday     *string `json:"birthday"`
	BirthdayType *string `json:"birthday_type"`
	Generation   *int    `json:"generation"`
	PhotoURL     *string `json:"photo_url"`
	Phone        *string `json:"phone"`
	Address      *string `json:"address"`
	Notes        *string `json:"notes"`
	IsAlive      *bool   `json:"is_alive"`
}

type PersonRepo struct {
	DB *sql.DB
}

const personSelectCols = `id, family_id, name, gender, COALESCE(birthday,''), COALESCE(birthday_type,'solar'),
       generation, COALESCE(photo_url,''), COALESCE(phone,''), COALESCE(address,''), COALESCE(notes,''),
       is_alive, created_at, updated_at`

func scanPerson(p *Person, scanner interface {
	Scan(dest ...interface{}) error
}, gen *sql.NullInt64, birthday, birthdayType, photoURL, phone, address, notes *sql.NullString) error {
	return scanner.Scan(
		&p.ID, &p.FamilyID, &p.Name, &p.Gender, birthday, birthdayType, gen,
		photoURL, phone, address, notes, &p.IsAlive, &p.CreatedAt, &p.UpdatedAt,
	)
}

func assignPersonFields(p *Person, birthday, birthdayType, photoURL, phone, address, notes *sql.NullString, gen *sql.NullInt64) {
	p.Birthday = birthday.String
	p.BirthdayType = birthdayType.String
	p.PhotoURL = photoURL.String
	p.Phone = phone.String
	p.Address = address.String
	p.Notes = notes.String
	if gen.Valid {
		g := int(gen.Int64)
		p.Generation = &g
	}
}

func (r *PersonRepo) Create(req PersonCreateReq) (*Person, error) {
	gender := req.Gender
	if gender == "" {
		gender = "unknown"
	}
	bt := req.BirthdayType
	if bt == "" {
		bt = "solar"
	}
	res, err := r.DB.Exec(
		`INSERT INTO person (family_id, name, gender, birthday, birthday_type, generation, photo_url, phone, address, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.FamilyID, req.Name, gender, req.Birthday, bt, req.Generation, req.PhotoURL, req.Phone, req.Address, req.Notes,
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
	var birthday, birthdayType, photoURL, phone, address, notes sql.NullString
	err := r.DB.QueryRow(
		"SELECT "+personSelectCols+" FROM person WHERE id = ?", id,
	).Scan(&p.ID, &p.FamilyID, &p.Name, &p.Gender, &birthday, &birthdayType, &gen,
		&photoURL, &phone, &address, &notes, &p.IsAlive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	assignPersonFields(p, &birthday, &birthdayType, &photoURL, &phone, &address, &notes, &gen)
	return p, nil
}

func (r *PersonRepo) ListByFamily(familyID int64, keyword string) ([]Person, error) {
	query := "SELECT " + personSelectCols + " FROM person WHERE family_id = ?"
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
		var birthday, birthdayType, photoURL, phone, address, notes sql.NullString
		if err := rows.Scan(&p.ID, &p.FamilyID, &p.Name, &p.Gender, &birthday, &birthdayType, &gen,
			&photoURL, &phone, &address, &notes, &p.IsAlive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		assignPersonFields(&p, &birthday, &birthdayType, &photoURL, &phone, &address, &notes, &gen)
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
	if req.BirthdayType != nil {
		sets = append(sets, "birthday_type = ?")
		args = append(args, *req.BirthdayType)
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

// GetAllAlive returns all alive persons with birthdays (both solar and lunar).
func (r *PersonRepo) GetAllAlive() ([]Person, error) {
	query := "SELECT " + personSelectCols + " FROM person WHERE is_alive = 1 AND birthday IS NOT NULL AND birthday != ''"
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var persons []Person
	for rows.Next() {
		var p Person
		var gen sql.NullInt64
		var birthday, birthdayType, photoURL, phone, address, notes sql.NullString
		if err := rows.Scan(&p.ID, &p.FamilyID, &p.Name, &p.Gender, &birthday, &birthdayType, &gen,
			&photoURL, &phone, &address, &notes, &p.IsAlive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		assignPersonFields(&p, &birthday, &birthdayType, &photoURL, &phone, &address, &notes, &gen)
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

// BirthdayPerson wraps a Person with the actual solar birthday date for display.
type BirthdayPerson struct {
	Person
	NextBirthday string `json:"next_birthday"` // actual solar date "YYYY-MM-DD"
	IsToday      bool   `json:"is_today"`
	Age          int    `json:"age"`
	LunarLabel   string `json:"lunar_label,omitempty"` // e.g. "六月十五" for lunar birthdays
}

// GetBirthdayUpcoming returns persons whose next birthday (solar) falls within `days` from today.
func (r *PersonRepo) GetBirthdayUpcoming(days int) ([]BirthdayPerson, error) {
	persons, err := r.GetAllAlive()
	if err != nil {
		return nil, err
	}
	return filterBirthdayUpcoming(persons, days), nil
}

// GetBirthdayToday returns persons whose birthday is today.
func (r *PersonRepo) GetBirthdayToday() ([]BirthdayPerson, error) {
	return r.GetBirthdayUpcoming(0)
}

// GetBirthdayTodayByFamily returns persons in a family whose birthday is today.
func (r *PersonRepo) GetBirthdayTodayByFamily(familyID int64) ([]BirthdayPerson, error) {
	all, err := r.GetAllAlive()
	if err != nil {
		return nil, err
	}
	var filtered []Person
	for _, p := range all {
		if p.FamilyID == familyID {
			filtered = append(filtered, p)
		}
	}
	return filterBirthdayUpcoming(filtered, 0), nil
}

func filterBirthdayUpcoming(persons []Person, days int) []BirthdayPerson {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endDate := today.AddDate(0, 0, days+1) // inclusive range

	var results []BirthdayPerson

	for _, p := range persons {
		if p.Birthday == "" {
			continue
		}

		var nextSolar time.Time
		var lunarLabel string

		if p.BirthdayType == "lunar" {
			month, day, err := lunar.ParseLunarBirthday(p.Birthday)
			if err != nil {
				continue
			}
			nextSolar = lunar.GetNextSolarBirthday(month, day, false, today)
			ld := &lunar.LunarDate{Year: now.Year(), Month: month, Day: day}
			lunarLabel = ld.Format()
		} else {
			// Solar birthday: parse MM-DD and find next occurrence
			var m, d int
			if _, err := fmt.Sscanf(p.Birthday[5:], "%d-%d", &m, &d); err != nil {
				continue
			}
			candidate := time.Date(now.Year(), time.Month(m), d, 0, 0, 0, 0, now.Location())
			if candidate.Before(today) {
				candidate = time.Date(now.Year()+1, time.Month(m), d, 0, 0, 0, 0, now.Location())
			}
			nextSolar = candidate
		}

		// Check if within range
		if nextSolar.Before(today) || !nextSolar.Before(endDate) {
			continue
		}

		isToday := nextSolar.Equal(today)
		birthYear := 0
		fmt.Sscanf(p.Birthday[:4], "%d", &birthYear)
		age := 0
		if birthYear > 0 {
			age = nextSolar.Year() - birthYear
		}

		results = append(results, BirthdayPerson{
			Person:       p,
			NextBirthday: nextSolar.Format("2006-01-02"),
			IsToday:      isToday,
			Age:          age,
			LunarLabel:   lunarLabel,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].NextBirthday < results[j].NextBirthday
	})

	return results
}