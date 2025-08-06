package services

import (
	"encoding/json"
	"fmt"
	"os"

	"GopherTales/internal/models"
)

// StoryService handles story-related business logic
type StoryService struct {
	story    *models.Story
	dataFile string
}

// NewStoryService creates a new story service
func NewStoryService(dataFile string) *StoryService {
	return &StoryService{
		dataFile: dataFile,
		story:    &models.Story{Arcs: make(map[string]models.Arc)},
	}
}

// LoadStory loads the story data from the JSON file
func (s *StoryService) LoadStory() error {
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return fmt.Errorf("failed to read story file %s: %w", s.dataFile, err)
	}

	var arcs map[string]models.Arc
	if err := json.Unmarshal(data, &arcs); err != nil {
		return fmt.Errorf("failed to unmarshal story data: %w", err)
	}

	// Add images to arcs
	for name, arc := range arcs {
		arc.Image = s.getImageFromArc(name)
		arcs[name] = arc
	}

	s.story.Arcs = arcs
	return nil
}

// GetArc retrieves an arc by name with proper error handling
func (s *StoryService) GetArc(arcName string) (models.Arc, string, error) {
	if len(s.story.Arcs) == 0 {
		return models.Arc{}, "", fmt.Errorf("story not loaded")
	}

	arc, finalName := s.story.GetArc(arcName)
	if arc.Title == "" && finalName != "" {
		return models.Arc{}, finalName, fmt.Errorf("arc '%s' not found", arcName)
	}

	return arc, finalName, nil
}

// GetStoryData returns the complete story data
func (s *StoryService) GetStoryData() *models.Story {
	return s.story
}

// ValidateArc checks if an arc name is valid
func (s *StoryService) ValidateArc(arcName string) bool {
	return s.story.HasArc(arcName) || arcName == "" // empty defaults to intro
}

// GetAvailableArcs returns all available arc names
func (s *StoryService) GetAvailableArcs() []string {
	return s.story.GetArcNames()
}

// getImageFromArc maps arc names to their corresponding images
func (s *StoryService) getImageFromArc(arcName string) string {
	imageMap := map[string]string{
		"intro":      "gopher_intro.png",
		"new-york":   "gopher_new-york.png",
		"debate":     "gopher_debate.png",
		"sean-kelly": "gopher_sean-kelly.png",
		"mark-bates": "gopher_mark-bates.png",
		"denver":     "gopher_denver.png",
		"home":       "gopher_home.png",
	}

	if image, exists := imageMap[arcName]; exists {
		return image
	}
	return "home_gopher.png"
}

// GetStoryStats returns statistics about the story
func (s *StoryService) GetStoryStats() map[string]any {
	if len(s.story.Arcs) == 0 {
		return map[string]any{
			"total_arcs": 0,
			"loaded":     false,
		}
	}

	totalOptions := 0
	totalStoryParagraphs := 0

	for _, arc := range s.story.Arcs {
		totalOptions += len(arc.Options)
		totalStoryParagraphs += len(arc.Story)
	}

	return map[string]any{
		"total_arcs":             len(s.story.Arcs),
		"total_options":          totalOptions,
		"total_story_paragraphs": totalStoryParagraphs,
		"loaded":                 true,
		"arcs":                   s.GetAvailableArcs(),
	}
}
