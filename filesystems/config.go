package filesystems

// Config holds the configuration for the filesystem manager.
type Config struct {
	Default string
	Disks   map[string]Disk
}

// Disk represents the configuration for a storage disk.
type Disk struct {
	Driver string
	Config map[string]interface{}
}
