package scafall

// Create a new project from a project template
func ExampleScaffold() {
	s := Scafall{}

	s.Scaffold("http://github.com/AidanDelaney/scafall-python-eg.git", "python-pi")
}

func ExampleScaffold_Overrides() {
	overrides := map[string]string{
		"PythonVersion": "python3.10",
	}
	s := New(overrides, map[string]interface{}{})

	// User is not prompted for PythonVersion
	s.Scaffold("http://github.com/AidanDelaney/scafall-python-eg.git", "python-pi")
}

func ExampleScaffold_Variables() {
	defaults := map[string]interface{}{
		"PythonVersion": []string{"python3.10", "python3.9"},
	}
	s := New(map[string]string{}, defaults)

	// User is prompted for PythonVersion, but the default choices are provided
	// programmatically
	s.Scaffold("http://github.com/AidanDelaney/scafall-python-eg.git", "python-pi")
}

// Create a new project from a project collection
func ExampleScaffoldCollection() {
	s := Scafall{}

	s.ScaffoldCollection("http://github.com/AidanDelaney/scafall-collection.git",
		"Choose a type of project to scaffold",
		"python-pi")
}
