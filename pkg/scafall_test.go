package scafall

// Create a new project from a project template
func ExampleScafall_Scaffold() {
	s := NewScafall(WithOutputFolder("python-pi"))

	s.Scaffold("http://github.com/AidanDelaney/scafall-python-eg.git")
}

func ExampleScafall_Scaffold_overrides() {
	overrides := map[string]string{
		"PythonVersion": "python3.10",
	}
	s := NewScafall(WithOverrides(overrides), WithOutputFolder("python-pi"))

	// User is not prompted for PythonVersion
	s.Scaffold("http://github.com/AidanDelaney/scafall-python-eg.git")
}

func ExampleScafall_Scaffold_variables() {
	defaults := map[string]interface{}{
		"PythonVersion": []string{"python3.10", "python3.9"},
	}
	s := NewScafall(WithDefaultValues(defaults), WithOutputFolder("python-pi"))

	// User is prompted for PythonVersion, but the default choices are provided
	// programmatically
	s.Scaffold("http://github.com/AidanDelaney/scafall-python-eg.git")
}

// Create a new project from a project collection
func ExampleScafall_ScaffoldCollection() {
	s := NewScafall(WithOutputFolder("collection-eg"))

	s.ScaffoldCollection(
		"http://github.com/AidanDelaney/scafall-collection.git",
		"Choose a type of project to scaffold",
	)
}
