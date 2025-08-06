package services

import (
	"os"
	"testing"

	"GopherTales/internal/models"
)

func TestStoryService_LoadStory(t *testing.T) {
	// Create a temporary test story file
	testStoryContent := `{
		"intro": {
			"title": "Test Story",
			"story": ["This is a test story."],
			"options": [
				{
					"text": "Go to chapter 2",
					"arc": "chapter2"
				}
			]
		},
		"chapter2": {
			"title": "Chapter 2",
			"story": ["This is chapter 2."],
			"options": []
		}
	}`

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test-story-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(testStoryContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Test loading the story
	service := NewStoryService(tmpFile.Name())
	err = service.LoadStory()
	if err != nil {
		t.Fatalf("Failed to load story: %v", err)
	}

	// Verify story was loaded correctly
	story := service.GetStoryData()
	if len(story.Arcs) != 2 {
		t.Errorf("Expected 2 arcs, got %d", len(story.Arcs))
	}

	// Check intro arc
	intro, exists := story.Arcs["intro"]
	if !exists {
		t.Error("Expected 'intro' arc to exist")
	}
	if intro.Title != "Test Story" {
		t.Errorf("Expected title 'Test Story', got '%s'", intro.Title)
	}
	if len(intro.Story) != 1 {
		t.Errorf("Expected 1 story paragraph, got %d", len(intro.Story))
	}
	if len(intro.Options) != 1 {
		t.Errorf("Expected 1 option, got %d", len(intro.Options))
	}
	if intro.Image != "gopher_intro.png" {
		t.Errorf("Expected image 'gopher_intro.png', got '%s'", intro.Image)
	}
}

func TestStoryService_GetArc(t *testing.T) {
	service := NewStoryService("nonexistent.json")

	// Test with empty story
	_, _, err := service.GetArc("intro")
	if err == nil {
		t.Error("Expected error when story not loaded")
	}

	// Load test data
	service.story = &models.Story{
		Arcs: map[string]models.Arc{
			"intro": {
				Title: "Introduction",
				Story: []string{"Welcome to the story."},
				Options: []models.Option{
					{Text: "Continue", Arc: "next"},
				},
				Image: "gopher_intro.png",
			},
		},
	}

	// Test getting existing arc
	arc, name, err := service.GetArc("intro")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if name != "intro" {
		t.Errorf("Expected arc name 'intro', got '%s'", name)
	}
	if arc.Title != "Introduction" {
		t.Errorf("Expected title 'Introduction', got '%s'", arc.Title)
	}

	// Test getting arc with empty name (should default to intro)
	arc, name, err = service.GetArc("")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if name != "intro" {
		t.Errorf("Expected default arc name 'intro', got '%s'", name)
	}

	// Test getting non-existent arc (should fallback to intro)
	arc, name, err = service.GetArc("nonexistent")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if name != "intro" {
		t.Errorf("Expected fallback to 'intro', got '%s'", name)
	}
	if arc.Title != "Introduction" {
		t.Errorf("Expected fallback arc title 'Introduction', got '%s'", arc.Title)
	}
}

func TestStoryService_ValidateArc(t *testing.T) {
	service := NewStoryService("nonexistent.json")
	service.story = &models.Story{
		Arcs: map[string]models.Arc{
			"intro":    {Title: "Introduction"},
			"chapter1": {Title: "Chapter 1"},
		},
	}

	tests := []struct {
		arcName  string
		expected bool
	}{
		{"intro", true},
		{"chapter1", true},
		{"", true}, // empty defaults to intro
		{"nonexistent", false},
		{"invalid", false},
	}

	for _, test := range tests {
		result := service.ValidateArc(test.arcName)
		if result != test.expected {
			t.Errorf("ValidateArc(%s): expected %v, got %v", test.arcName, test.expected, result)
		}
	}
}

func TestStoryService_GetAvailableArcs(t *testing.T) {
	service := NewStoryService("nonexistent.json")
	service.story = &models.Story{
		Arcs: map[string]models.Arc{
			"intro":    {Title: "Introduction"},
			"chapter1": {Title: "Chapter 1"},
			"ending":   {Title: "The End"},
		},
	}

	arcs := service.GetAvailableArcs()
	if len(arcs) != 3 {
		t.Errorf("Expected 3 arcs, got %d", len(arcs))
	}

	// Check that all expected arcs are present
	arcMap := make(map[string]bool)
	for _, arc := range arcs {
		arcMap[arc] = true
	}

	expectedArcs := []string{"intro", "chapter1", "ending"}
	for _, expected := range expectedArcs {
		if !arcMap[expected] {
			t.Errorf("Expected arc '%s' not found in available arcs", expected)
		}
	}
}

func TestStoryService_GetStoryStats(t *testing.T) {
	service := NewStoryService("nonexistent.json")

	// Test with empty story
	stats := service.GetStoryStats()
	if stats["loaded"].(bool) != false {
		t.Error("Expected loaded to be false for empty story")
	}
	if stats["total_arcs"].(int) != 0 {
		t.Error("Expected total_arcs to be 0 for empty story")
	}

	// Test with loaded story
	service.story = &models.Story{
		Arcs: map[string]models.Arc{
			"intro": {
				Title: "Introduction",
				Story: []string{"Para 1", "Para 2"},
				Options: []models.Option{
					{Text: "Option 1", Arc: "chapter1"},
					{Text: "Option 2", Arc: "chapter2"},
				},
			},
			"chapter1": {
				Title: "Chapter 1",
				Story: []string{"Chapter 1 content"},
				Options: []models.Option{
					{Text: "Go to ending", Arc: "ending"},
				},
			},
		},
	}

	stats = service.GetStoryStats()
	if stats["loaded"].(bool) != true {
		t.Error("Expected loaded to be true for loaded story")
	}
	if stats["total_arcs"].(int) != 2 {
		t.Errorf("Expected total_arcs to be 2, got %v", stats["total_arcs"])
	}
	if stats["total_options"].(int) != 3 {
		t.Errorf("Expected total_options to be 3, got %v", stats["total_options"])
	}
	if stats["total_story_paragraphs"].(int) != 3 {
		t.Errorf("Expected total_story_paragraphs to be 3, got %v", stats["total_story_paragraphs"])
	}
}

func TestStoryService_getImageFromArc(t *testing.T) {
	service := NewStoryService("nonexistent.json")

	tests := []struct {
		arcName  string
		expected string
	}{
		{"intro", "gopher_intro.png"},
		{"new-york", "gopher_new-york.png"},
		{"debate", "gopher_debate.png"},
		{"sean-kelly", "gopher_sean-kelly.png"},
		{"mark-bates", "gopher_mark-bates.png"},
		{"denver", "gopher_denver.png"},
		{"home", "gopher_home.png"},
		{"unknown", "default_gopher.png"},
		{"", "default_gopher.png"},
	}

	for _, test := range tests {
		result := service.getImageFromArc(test.arcName)
		if result != test.expected {
			t.Errorf("getImageFromArc(%s): expected %s, got %s", test.arcName, test.expected, result)
		}
	}
}

func TestStoryService_LoadStory_InvalidFile(t *testing.T) {
	service := NewStoryService("nonexistent.json")
	err := service.LoadStory()
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}
}

func TestStoryService_LoadStory_InvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tmpFile, err := os.CreateTemp("", "invalid-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("invalid json content"); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	service := NewStoryService(tmpFile.Name())
	err = service.LoadStory()
	if err == nil {
		t.Error("Expected error when loading invalid JSON")
	}
}
