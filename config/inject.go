package config

import (
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
)

// ------------------------- Inject / Unmarshal -------------------------

// Unmarshal is an improved version that merges registry defaults, YAML, and ENV properly.
// It flattens all sources, overlays them, rebuilds the nested map, and decodes into target.
func Unmarshal(prefix string, target any) error {
	syncEnv(prefix)

	// Step 1: Flatten defaults from registry
	flat := make(map[string]any)
	for k, v := range registry {
		if strings.HasPrefix(k, prefix+".") {
			flat[k] = v
		}
	}

	// Step 2: Flatten YAML/config file
	sub := viperInstance.Sub(prefix)
	if sub != nil {
		yamlFlat := flattenMap(prefix, sub.AllSettings())
		for k, v := range yamlFlat {
			flat[k] = v
		}
	}

	// Step 3: Overlay ENV variables from viper
	// (viper flattens nested keys with dots, so we can directly check flat keys)
	for key := range flat {
		if viperInstance.IsSet(key) {
			flat[key] = viperInstance.Get(key)
		}
	}

	// Step 4: Rebuild nested map
	nested := make(map[string]any)
	for fullKey, val := range flat {
		short := fullKey[len(prefix)+1:]
		assignNested(nested, short, val)
	}

	// Step 5: Decode
	decoderCfg := &mapstructure.DecoderConfig{
		TagName:          "mapstructure",
		Result:           target,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToIPNetHookFunc(),
			mapstructure.StringToIPHookFunc(),
			mapstructure.StringToTimeHookFunc(time.RFC3339),
		),
	}
	decoder, err := mapstructure.NewDecoder(decoderCfg)
	if err != nil {
		return err
	}
	return decoder.Decode(nested)
}

// flattenMap converts nested maps into flattened map[string]any with dotted keys.
func flattenMap(prefix string, data map[string]any) map[string]any {
	out := make(map[string]any)
	for k, v := range data {
		full := prefix + "." + k
		switch val := v.(type) {
		case map[string]any:
			for fk, fv := range flattenMap(full, val) {
				out[fk] = fv
			}
		default:
			out[full] = v
		}
	}
	return out
}

// assignNested creates nested map from dotted keys (e.g. nested.key1 -> map[nested][key1])
func assignNested(dest map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	m := dest
	for i, part := range parts {
		if i == len(parts)-1 {
			m[part] = value
			return
		}
		if _, ok := m[part]; !ok {
			m[part] = make(map[string]any)
		}
		m = m[part].(map[string]any)
	}
}

// Inject is a convenience wrapper that will Unmarshal and then return the target filled.
// Example:
//
//	var ac AppConfig
//	if err := Inject("app", &ac); err != nil { ... }
func Inject(prefix string, target any) error {
	return Unmarshal(prefix, target)
}
