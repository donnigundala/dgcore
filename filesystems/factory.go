package filesystems

import (
	"fmt"
	"log/slog"
)

// Factory uses a registry to create drivers.
type Factory struct{}

func NewFactory() *Factory { return &Factory{} }

type driverConstructor func(config interface{}, logger *slog.Logger, traceIDKey any) (Storage, error)

var driverRegistry = map[string]driverConstructor{}

// RegisterDriver registers a driver constructor (call from driver init()).
func RegisterDriver(name string, ctor driverConstructor) {
	driverRegistry[name] = ctor
}

// Create returns a Storage for the given driver name using the provided config.
func (f *Factory) Create(driver string, config interface{}, logger *slog.Logger, traceIDKey any) (Storage, error) {
	if ctor, ok := driverRegistry[driver]; ok {
		return ctor(config, logger, traceIDKey)
	}
	return nil, fmt.Errorf("unsupported driver: %s", driver)
}
