package module

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/cloudquery/cq-provider-sdk/cqproto"
)

type Module interface {
	// ID returns the name of the module
	ID() string
	// Configure configures the module to run
	Configure(context.Context, hcl.Body, ModuleRunParams) error
	// Execute executes the module, using given args in ExecuteRequest
	Execute(context.Context, *ExecuteRequest) *ExecutionResult
	// ExampleConfig returns an example configuration to be put in config.hcl
	ExampleConfig() string
}

type ModuleRunParams interface{}

type ExecuteRequest struct {
	// Params are the invocation parameters specific to the module
	Params ModuleRunParams

	// Providers is the list of providers to process
	Providers []*cqproto.GetProviderSchemaResponse
	// Conn is the db connection to use
	Conn *pgxpool.Conn
}

type ExecutionResult struct {
	Result   interface{} `json:"result"`
	Error    error       `json:"-"`
	ErrorMsg string      `json:"error,omitempty"`
}

type ExitCoder interface {
	ExitCode() int
}
