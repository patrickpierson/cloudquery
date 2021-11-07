package console

// ModuleCallRequest is the request used to call a module.
type ModuleCallRequest struct {
	// Name of the module
	Name string

	// Params are the invocation parameters specific to the module
	Params interface{}

	// ModConfigPath is the path to the module config file to use.
	ModConfigPath string

	// OutputPath is the filename to save output to
	OutputPath string
}