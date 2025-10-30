package filesystems

import "fmt"
 
type configConverter func(map[string]interface{}) (interface{}, error)

var converterRegistry = map[string]configConverter{}

// RegisterConfigConverter registers a converter for a driver (call from driver init()).
func RegisterConfigConverter(driver string, conv configConverter) {
	converterRegistry[driver] = conv
}

func convertConfig(driver string, config map[string]interface{}) (interface{}, error) {
	if conv, ok := converterRegistry[driver]; ok {
		return conv(config)
	}
	return nil, fmt.Errorf("unsupported driver for config conversion: %s", driver)
}
