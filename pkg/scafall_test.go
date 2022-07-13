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
