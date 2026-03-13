package curriculum

import "fmt"

var registry []Lesson

// Register adds a lesson to the global registry. Called from init() in each lesson file.
func Register(l Lesson) {
	registry = append(registry, l)
}

// All returns all lessons in order.
func All() []Lesson {
	return registry
}

// Get returns a lesson by ID.
func Get(id LessonID) (*Lesson, error) {
	for i := range registry {
		if registry[i].ID == id {
			return &registry[i], nil
		}
	}
	return nil, fmt.Errorf("lesson %q not found", id)
}

// Next returns the lesson after the given ID, or nil if it's the last one.
func Next(id LessonID) *Lesson {
	for i, l := range registry {
		if l.ID == id && i+1 < len(registry) {
			return &registry[i+1]
		}
	}
	return nil
}

// Prev returns the lesson before the given ID, or nil if it's the first one.
func Prev(id LessonID) *Lesson {
	for i, l := range registry {
		if l.ID == id && i > 0 {
			return &registry[i-1]
		}
	}
	return nil
}
