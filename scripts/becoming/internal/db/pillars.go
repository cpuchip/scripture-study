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
func (db *DB) HasPillars(userID int64) (bool, error) {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pillars WHERE user_id = ?`, userID).Scan(&count); err != nil {
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

// CreatePillar inserts a new pillar, scoped to user.
func (db *DB) CreatePillar(userID int64, p *Pillar) error {
	id, err := db.InsertReturningID(`
		INSERT INTO pillars (user_id, name, description, icon, parent_id, sort_order)
		VALUES (?, ?, ?, ?, ?, ?)`,
		userID, p.Name, p.Description, p.Icon, p.ParentID, p.SortOrder,
	)
	if err != nil {
		return fmt.Errorf("inserting pillar: %w", err)
	}
	p.ID = id
	row := db.QueryRow(`SELECT created_at FROM pillars WHERE id = ?`, p.ID)
	_ = row.Scan(&p.CreatedAt)
	return nil
}

// ListPillars returns all pillars as a flat list with practice/task counts, scoped to user.
func (db *DB) ListPillars(userID int64) ([]*Pillar, error) {
	rows, err := db.Query(`
		SELECT p.id, p.name, COALESCE(p.description, ''), COALESCE(p.icon, ''), p.parent_id, p.sort_order, p.created_at,
		       (SELECT COUNT(*) FROM practice_pillars pp WHERE pp.pillar_id = p.id) as practice_count,
		       (SELECT COUNT(*) FROM task_pillars tp WHERE tp.pillar_id = p.id) as task_count
		FROM pillars p
		WHERE p.user_id = ?
		ORDER BY p.sort_order, p.id`, userID)
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

// ListPillarsTree returns pillars organized as a tree (parents with children), scoped to user.
func (db *DB) ListPillarsTree(userID int64) ([]*Pillar, error) {
	all, err := db.ListPillars(userID)
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

// GetPillar returns a single pillar by ID, scoped to user.
func (db *DB) GetPillar(userID, id int64) (*Pillar, error) {
	p := &Pillar{}
	err := db.QueryRow(`
		SELECT p.id, p.name, COALESCE(p.description, ''), COALESCE(p.icon, ''), p.parent_id, p.sort_order, p.created_at,
		       (SELECT COUNT(*) FROM practice_pillars pp WHERE pp.pillar_id = p.id),
		       (SELECT COUNT(*) FROM task_pillars tp WHERE tp.pillar_id = p.id)
		FROM pillars p WHERE p.id = ? AND p.user_id = ?`, id, userID,
	).Scan(&p.ID, &p.Name, &p.Description, &p.Icon, &p.ParentID,
		&p.SortOrder, &p.CreatedAt, &p.PracticeCount, &p.TaskCount)
	if err != nil {
		return nil, fmt.Errorf("getting pillar: %w", err)
	}
	return p, nil
}

// UpdatePillar updates an existing pillar, scoped to user.
func (db *DB) UpdatePillar(userID int64, p *Pillar) error {
	_, err := db.Exec(`
		UPDATE pillars SET name=?, description=?, icon=?, parent_id=?, sort_order=?
		WHERE id=? AND user_id=?`,
		p.Name, p.Description, p.Icon, p.ParentID, p.SortOrder, p.ID, userID,
	)
	return err
}

// DeletePillar removes a pillar and its links (cascades to children via FK), scoped to user.
func (db *DB) DeletePillar(userID, id int64) error {
	_, err := db.Exec(`DELETE FROM pillars WHERE id = ? AND user_id = ?`, id, userID)
	return err
}

// --- Practice ↔ Pillar linking ---

// LinkPracticePillar adds a link between a practice and a pillar (both must belong to user).
func (db *DB) LinkPracticePillar(userID, practiceID, pillarID int64) error {
	_, err := db.Exec(`
		INSERT OR IGNORE INTO practice_pillars (practice_id, pillar_id)
		SELECT ?, ? WHERE
			EXISTS (SELECT 1 FROM practices WHERE id = ? AND user_id = ?) AND
			EXISTS (SELECT 1 FROM pillars WHERE id = ? AND user_id = ?)`,
		practiceID, pillarID, practiceID, userID, pillarID, userID)
	return err
}

// UnlinkPracticePillar removes a link between a practice and a pillar.
func (db *DB) UnlinkPracticePillar(userID, practiceID, pillarID int64) error {
	_, err := db.Exec(`
		DELETE FROM practice_pillars WHERE practice_id = ? AND pillar_id = ?
		AND practice_id IN (SELECT id FROM practices WHERE user_id = ?)`,
		practiceID, pillarID, userID)
	return err
}

// SetPracticePillars replaces all pillar links for a practice.
func (db *DB) SetPracticePillars(userID, practiceID int64, pillarIDs []int64) error {
	// Verify practice belongs to user
	var owner int64
	err := db.QueryRow(`SELECT user_id FROM practices WHERE id = ?`, practiceID).Scan(&owner)
	if err != nil || owner != userID {
		return fmt.Errorf("practice %d not found", practiceID)
	}

	if _, err := db.Exec(`DELETE FROM practice_pillars WHERE practice_id = ?`, practiceID); err != nil {
		return err
	}
	for _, pid := range pillarIDs {
		if err := db.LinkPracticePillar(userID, practiceID, pid); err != nil {
			return err
		}
	}
	return nil
}

// GetPracticePillars returns pillar links for a practice.
func (db *DB) GetPracticePillars(userID, practiceID int64) ([]PillarLink, error) {
	rows, err := db.Query(`
		SELECT p.id, p.name, COALESCE(p.icon, '')
		FROM pillars p
		JOIN practice_pillars pp ON pp.pillar_id = p.id
		WHERE pp.practice_id = ? AND p.user_id = ?
		ORDER BY p.sort_order`, practiceID, userID)
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

// LinkTaskPillar adds a link between a task and a pillar (both must belong to user).
func (db *DB) LinkTaskPillar(userID, taskID, pillarID int64) error {
	_, err := db.Exec(`
		INSERT OR IGNORE INTO task_pillars (task_id, pillar_id)
		SELECT ?, ? WHERE
			EXISTS (SELECT 1 FROM tasks WHERE id = ? AND user_id = ?) AND
			EXISTS (SELECT 1 FROM pillars WHERE id = ? AND user_id = ?)`,
		taskID, pillarID, taskID, userID, pillarID, userID)
	return err
}

// UnlinkTaskPillar removes a link between a task and a pillar.
func (db *DB) UnlinkTaskPillar(userID, taskID, pillarID int64) error {
	_, err := db.Exec(`
		DELETE FROM task_pillars WHERE task_id = ? AND pillar_id = ?
		AND task_id IN (SELECT id FROM tasks WHERE user_id = ?)`,
		taskID, pillarID, userID)
	return err
}

// SetTaskPillars replaces all pillar links for a task.
func (db *DB) SetTaskPillars(userID, taskID int64, pillarIDs []int64) error {
	// Verify task belongs to user
	var owner int64
	err := db.QueryRow(`SELECT user_id FROM tasks WHERE id = ?`, taskID).Scan(&owner)
	if err != nil || owner != userID {
		return fmt.Errorf("task %d not found", taskID)
	}

	if _, err := db.Exec(`DELETE FROM task_pillars WHERE task_id = ?`, taskID); err != nil {
		return err
	}
	for _, pid := range pillarIDs {
		if err := db.LinkTaskPillar(userID, taskID, pid); err != nil {
			return err
		}
	}
	return nil
}
