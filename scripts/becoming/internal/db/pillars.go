package db

import (
	"fmt"
)

// Pillar represents a growth area / vision.
type Pillar struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	ParentID    *int64 `json:"parent_id,omitempty"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`

	// Populated on read
	Children       []*Pillar `json:"children,omitempty"`
	PracticeCount  int       `json:"practice_count"`
	TaskCount      int       `json:"task_count"`
	CompletionRate float64   `json:"completion_rate,omitempty"` // From recent activity
}

// PillarLink represents a link between a practice/task and a pillar.
type PillarLink struct {
	PillarID   int64  `json:"pillar_id"`
	PillarName string `json:"pillar_name,omitempty"`
	PillarIcon string `json:"pillar_icon,omitempty"`
}

// --- Default Pillars ---

var defaultPillars = []struct {
	Name        string
	Description string
	Icon        string
}{
	{"Spiritual", "Faith, prayer, scripture study, temple worship", "🙏"},
	{"Social", "Relationships, service, community, family", "🤝"},
	{"Intellectual", "Learning, study, skill development, wisdom", "📚"},
	{"Physical", "Health, fitness, rest, nutrition", "💪"},
}

// SeedPillars checks if the onboarding flag is set. Returns default pillar suggestions
// without creating them — the frontend handles onboarding.
func (db *DB) HasPillars() (bool, error) {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pillars`).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetDefaultPillarSuggestions returns the 4 default pillar suggestions for onboarding.
func GetDefaultPillarSuggestions() []map[string]string {
	result := make([]map[string]string, len(defaultPillars))
	for i, p := range defaultPillars {
		result[i] = map[string]string{
			"name":        p.Name,
			"description": p.Description,
			"icon":        p.Icon,
		}
	}
	return result
}

// CreatePillar inserts a new pillar.
func (db *DB) CreatePillar(p *Pillar) error {
	result, err := db.Exec(`
		INSERT INTO pillars (name, description, icon, parent_id, sort_order)
		VALUES (?, ?, ?, ?, ?)`,
		p.Name, p.Description, p.Icon, p.ParentID, p.SortOrder,
	)
	if err != nil {
		return fmt.Errorf("inserting pillar: %w", err)
	}
	p.ID, _ = result.LastInsertId()
	row := db.QueryRow(`SELECT created_at FROM pillars WHERE id = ?`, p.ID)
	_ = row.Scan(&p.CreatedAt)
	return nil
}

// ListPillars returns all pillars as a flat list with practice/task counts.
func (db *DB) ListPillars() ([]*Pillar, error) {
	rows, err := db.Query(`
		SELECT p.id, p.name, COALESCE(p.description, ''), COALESCE(p.icon, ''), p.parent_id, p.sort_order, p.created_at,
		       (SELECT COUNT(*) FROM practice_pillars pp WHERE pp.pillar_id = p.id) as practice_count,
		       (SELECT COUNT(*) FROM task_pillars tp WHERE tp.pillar_id = p.id) as task_count
		FROM pillars p
		ORDER BY p.sort_order, p.id`)
	if err != nil {
		return nil, fmt.Errorf("listing pillars: %w", err)
	}
	defer rows.Close()

	var pillars []*Pillar
	for rows.Next() {
		p := &Pillar{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Icon, &p.ParentID,
			&p.SortOrder, &p.CreatedAt, &p.PracticeCount, &p.TaskCount); err != nil {
			return nil, fmt.Errorf("scanning pillar: %w", err)
		}
		pillars = append(pillars, p)
	}
	return pillars, rows.Err()
}

// ListPillarsTree returns pillars organized as a tree (parents with children).
func (db *DB) ListPillarsTree() ([]*Pillar, error) {
	all, err := db.ListPillars()
	if err != nil {
		return nil, err
	}

	// Build map and tree
	byID := make(map[int64]*Pillar)
	for _, p := range all {
		p.Children = []*Pillar{}
		byID[p.ID] = p
	}

	var roots []*Pillar
	for _, p := range all {
		if p.ParentID != nil {
			if parent, ok := byID[*p.ParentID]; ok {
				parent.Children = append(parent.Children, p)
				continue
			}
		}
		roots = append(roots, p)
	}
	return roots, nil
}

// GetPillar returns a single pillar by ID.
func (db *DB) GetPillar(id int64) (*Pillar, error) {
	p := &Pillar{}
	err := db.QueryRow(`
		SELECT p.id, p.name, COALESCE(p.description, ''), COALESCE(p.icon, ''), p.parent_id, p.sort_order, p.created_at,
		       (SELECT COUNT(*) FROM practice_pillars pp WHERE pp.pillar_id = p.id),
		       (SELECT COUNT(*) FROM task_pillars tp WHERE tp.pillar_id = p.id)
		FROM pillars p WHERE p.id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Icon, &p.ParentID,
		&p.SortOrder, &p.CreatedAt, &p.PracticeCount, &p.TaskCount)
	if err != nil {
		return nil, fmt.Errorf("getting pillar: %w", err)
	}
	return p, nil
}

// UpdatePillar updates an existing pillar.
func (db *DB) UpdatePillar(p *Pillar) error {
	_, err := db.Exec(`
		UPDATE pillars SET name=?, description=?, icon=?, parent_id=?, sort_order=?
		WHERE id=?`,
		p.Name, p.Description, p.Icon, p.ParentID, p.SortOrder, p.ID,
	)
	return err
}

// DeletePillar removes a pillar and its links (cascades to children via FK).
func (db *DB) DeletePillar(id int64) error {
	_, err := db.Exec(`DELETE FROM pillars WHERE id = ?`, id)
	return err
}

// --- Practice ↔ Pillar linking ---

// LinkPracticePillar adds a link between a practice and a pillar.
func (db *DB) LinkPracticePillar(practiceID, pillarID int64) error {
	_, err := db.Exec(`INSERT OR IGNORE INTO practice_pillars (practice_id, pillar_id) VALUES (?, ?)`,
		practiceID, pillarID)
	return err
}

// UnlinkPracticePillar removes a link between a practice and a pillar.
func (db *DB) UnlinkPracticePillar(practiceID, pillarID int64) error {
	_, err := db.Exec(`DELETE FROM practice_pillars WHERE practice_id = ? AND pillar_id = ?`,
		practiceID, pillarID)
	return err
}

// SetPracticePillars replaces all pillar links for a practice.
func (db *DB) SetPracticePillars(practiceID int64, pillarIDs []int64) error {
	if _, err := db.Exec(`DELETE FROM practice_pillars WHERE practice_id = ?`, practiceID); err != nil {
		return err
	}
	for _, pid := range pillarIDs {
		if err := db.LinkPracticePillar(practiceID, pid); err != nil {
			return err
		}
	}
	return nil
}

// GetPracticePillars returns pillar links for a practice.
func (db *DB) GetPracticePillars(practiceID int64) ([]PillarLink, error) {
	rows, err := db.Query(`
		SELECT p.id, p.name, COALESCE(p.icon, '')
		FROM pillars p
		JOIN practice_pillars pp ON pp.pillar_id = p.id
		WHERE pp.practice_id = ?
		ORDER BY p.sort_order`, practiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []PillarLink
	for rows.Next() {
		l := PillarLink{}
		if err := rows.Scan(&l.PillarID, &l.PillarName, &l.PillarIcon); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

// --- Task ↔ Pillar linking ---

// LinkTaskPillar adds a link between a task and a pillar.
func (db *DB) LinkTaskPillar(taskID, pillarID int64) error {
	_, err := db.Exec(`INSERT OR IGNORE INTO task_pillars (task_id, pillar_id) VALUES (?, ?)`,
		taskID, pillarID)
	return err
}

// UnlinkTaskPillar removes a link between a task and a pillar.
func (db *DB) UnlinkTaskPillar(taskID, pillarID int64) error {
	_, err := db.Exec(`DELETE FROM task_pillars WHERE task_id = ? AND pillar_id = ?`,
		taskID, pillarID)
	return err
}

// SetTaskPillars replaces all pillar links for a task.
func (db *DB) SetTaskPillars(taskID int64, pillarIDs []int64) error {
	if _, err := db.Exec(`DELETE FROM task_pillars WHERE task_id = ?`, taskID); err != nil {
		return err
	}
	for _, pid := range pillarIDs {
		if err := db.LinkTaskPillar(taskID, pid); err != nil {
			return err
		}
	}
	return nil
}
