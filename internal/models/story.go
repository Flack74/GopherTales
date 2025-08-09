package models

// Option represents a choice in the story that leads to another arc
type Option struct {
	Text string `json:"text"`
	Arc  string `json:"arc"`
}

// Arc represents a single story segment with choices
type Arc struct {
	Title   string   `json:"title"`
	Story   []string `json:"story"`
	Options []Option `json:"options"`
	Image   string   `json:"-"` // Not stored in JSON, computed dynamically
}

// Story represents the complete story with all arcs
type Story struct {
	Arcs map[string]Arc `json:"-"`
}

// PageData represents the data passed to templates
type PageData struct {
	Arc     Arc
	ArcName string
	Gopher  string
	User    *User
}

// GetArc retrieves an arc by name, returns default "intro" if not found
func (s *Story) GetArc(arcName string) (Arc, string) {
	if arcName == "" {
		arcName = "intro"
	}

	arc, exists := s.Arcs[arcName]
	if !exists {
		// Fallback to intro if arc doesn't exist
		if introArc, introExists := s.Arcs["intro"]; introExists {
			return introArc, "intro"
		}
		// Return empty arc if intro doesn't exist either
		return Arc{}, arcName
	}

	return arc, arcName
}

// HasArc checks if an arc exists in the story
func (s *Story) HasArc(arcName string) bool {
	_, exists := s.Arcs[arcName]
	return exists
}

// GetArcNames returns all available arc names
func (s *Story) GetArcNames() []string {
	names := make([]string, 0, len(s.Arcs))
	for name := range s.Arcs {
		names = append(names, name)
	}
	return names
}
