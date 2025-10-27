# Example
```
// ------------------------- Example usage (put in main.go or docs) -------------------------
/*
package main

import (
    "fmt"
    _ "your_project/configs" // blank import to run init() of each config file
    "your_project/pkg/config"
)

// AppConfig is an example struct you want injected
type AppConfig struct {
    Name     string `mapstructure:"name"`
    Env      string `mapstructure:"env"`
    Timezone string `mapstructure:"timezone"`
    Debug    bool   `mapstructure:"debug"`
}

func main() {
    // Load env + yaml
    config.Load()

    // Optional auto-discovery check (verifies configs/ exists)
    config.AutoDiscover("configs")

    // Inject config into struct
    var ac AppConfig
    if err := config.Inject("app", &ac); err != nil {
        panic(err)
    }

    fmt.Printf("App config: %+v\n", ac)

    // Print all registered keys
    config.PrintAll()
}
*/
```