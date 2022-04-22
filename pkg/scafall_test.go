package scafall

// Create a new project from a project template
func ExampleScaffold() {
	s := New(map[string]interface{}{}, []string{})

	s.Scaffold("http://github.com/AidanDelaney/scafall-python-eg.git", "python-pi")
}

// Create a new project from a project collection
func ExampleScaffoldCollection() {
	s := New(map[string]interface{}{}, []string{})

	s.ScaffoldCollection("http://github.com/AidanDelaney/scafall-collection.git",
		"Choose a type of project to scaffold",
		"python-pi")
}
