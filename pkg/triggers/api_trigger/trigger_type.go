package api_trigger

import (
	"github.com/The-New-Fork/pipeline/pkg/triggers/api_trigger/config"
	"github.com/The-New-Fork/pipeline/pkg/triggers/api_trigger/listener"
	"github.com/unchainio/interfaces/adapter"
)

type Trigger struct {
	cfg    *config.Config
	stub   adapter.Stub
	port   string
	client *listener.Server
}
