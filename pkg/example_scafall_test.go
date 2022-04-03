package scafall

// Create a new project from a project template
func ExampleScafall_Scaffold() {
	s, _ := NewScafall("http://github.com/AidanDelaney/scafall-python-eg.git",
		WithOutputFolder("python-pi"))

	s.Scaffold()
}

func ExampleScafall_Scaffold_arguments() {
	arguments := map[string]string{
		"PythonVersion": "python3.10",
	}
	s, _ := NewScafall("http://github.com/AidanDelaney/scafall-python-eg.git",
		WithArguments(arguments),
		WithOutputFolder("python-pi"))

	// User is not prompted for PythonVersion
	s.Scaffold()
}
