package services

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"GopherTales/internal/models"
)

// StoryService handles story-related business logic
type StoryService struct {
	story         *models.Story
	gopherStories map[string]map[string]models.Arc
	dataFile      string
}

// NewStoryService creates a new story service
func NewStoryService(dataFile string) *StoryService {
	return &StoryService{
		dataFile:      dataFile,
		story:         &models.Story{Arcs: make(map[string]models.Arc)},
		gopherStories: make(map[string]map[string]models.Arc),
	}
}

// LoadStory loads the story data from the JSON file
func (s *StoryService) LoadStory() error {
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return fmt.Errorf("failed to read story file %s: %w", s.dataFile, err)
	}

	// Try to load as gopher-based structure first
	var gopherData map[string]map[string]models.Arc
	if err := json.Unmarshal(data, &gopherData); err == nil {
		// Check if this is actually gopher data (has nested structure)
		if s.isGopherStructure(gopherData) {
			s.gopherStories = gopherData

			// Create a default story from first gopher's intro for classic mode
			if len(gopherData) > 0 {
				for _, arcs := range gopherData {
					if introArc, exists := arcs["intro"]; exists {
						introArc.Image = s.getImageFromArc("intro")
						s.story.Arcs = map[string]models.Arc{"intro": introArc}
						break
					}
				}
			}
			return nil
		}
	}

	// Load as classic structure
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

// isGopherStructure checks if the data has the gopher-based nested structure
func (s *StoryService) isGopherStructure(data map[string]map[string]models.Arc) bool {
	// Check if we have color names as top-level keys
	gopherColors := []string{"blue", "cyan", "brown", "green", "pink", "purple"}
	for _, color := range gopherColors {
		if _, exists := data[color]; exists {
			return true
		}
	}
	return false
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

// GetGopherArc retrieves an arc for a specific gopher
func (s *StoryService) GetGopherArc(gopher, arcName string) (models.Arc, string, error) {
	if len(s.gopherStories) == 0 {
		return models.Arc{}, "", fmt.Errorf("gopher stories not loaded")
	}

	gopherArcs, exists := s.gopherStories[gopher]
	if !exists {
		return models.Arc{}, "", fmt.Errorf("gopher '%s' not found", gopher)
	}

	if arcName == "" {
		arcName = "intro"
	}

	arc, exists := gopherArcs[arcName]
	if !exists {
		return models.Arc{}, arcName, fmt.Errorf("arc '%s' not found for gopher '%s'", arcName, gopher)
	}

	arc.Image = s.getImageFromGopherArc(gopher, arcName)
	return arc, arcName, nil
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
	// For gopher_six.json, we use gopher-specific images
	return "home_gopher.png"
}

// getImageFromGopherArc maps gopher colors and arc names to images
func (s *StoryService) getImageFromGopherArc(gopher, arcName string) string {
	// Use placeholder images for now
	gopherImages := map[string]string{
		"blue":   "gopher_blue.png",
		"cyan":   "gopher_cyan.png",
		"brown":  "gopher_brown.png",
		"green":  "gopher_green.png",
		"pink":   "gopher_pink.png",
		"purple": "gopher_purple.png",
	}

	if image, exists := gopherImages[gopher]; exists {
		return image
	}
	return "home_gopher.png"
}

// GetAvailableGophers returns all available gopher colors
func (s *StoryService) GetAvailableGophers() []string {
	gophers := make([]string, 0, len(s.gopherStories))
	for gopher := range s.gopherStories {
		gophers = append(gophers, gopher)
	}
	return gophers
}

// GetGopherStats returns detailed statistics for each gopher with caching
func (s *StoryService) GetGopherStats() map[string]map[string]any {

	stats := make(map[string]map[string]any)

	for gopher, arcs := range s.gopherStories {
		arcCount := len(arcs)
		totalWords := 0
		totalOptions := 0

		for _, arc := range arcs {
			for _, paragraph := range arc.Story {
				totalWords += len(strings.Fields(paragraph))
			}
			totalOptions += len(arc.Options)
		}

		// Estimate reading time (average 200 words per minute)
		readTime := totalWords / 200
		if readTime < 1 {
			readTime = 1
		}

		stats[gopher] = map[string]any{
			"arc_count":     arcCount,
			"total_words":   totalWords,
			"total_options": totalOptions,
			"read_time":     readTime,
		}
	}

	return stats
}

// ValidateStoryIntegrity checks for broken story links
func (s *StoryService) ValidateStoryIntegrity() map[string][]string {
	issues := make(map[string][]string)

	for gopher, arcs := range s.gopherStories {
		for arcName, arc := range arcs {
			for _, option := range arc.Options {
				if _, exists := arcs[option.Arc]; !exists {
					key := fmt.Sprintf("%s:%s", gopher, arcName)
					issues[key] = append(issues[key], fmt.Sprintf("Broken link to arc '%s'", option.Arc))
				}
			}
		}
	}

	return issues
}

// GetStoryStats returns statistics about the story
func (s *StoryService) GetStoryStats() map[string]any {
	// Count gopher stories if available
	if len(s.gopherStories) > 0 {
		totalArcs := 0
		totalOptions := 0
		totalStoryParagraphs := 0

		for _, arcs := range s.gopherStories {
			totalArcs += len(arcs)
			for _, arc := range arcs {
				totalOptions += len(arc.Options)
				totalStoryParagraphs += len(arc.Story)
			}
		}

		return map[string]any{
			"total_arcs":             totalArcs,
			"total_options":          totalOptions,
			"total_story_paragraphs": totalStoryParagraphs,
			"loaded":                 true,
			"gopher_count":           len(s.gopherStories),
		}
	}

	// Fallback to classic story stats
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
