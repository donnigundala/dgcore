package mongo

// Helper function to join host strings for the URI
func joinHosts(hosts []string) string {
	var hostStr string
	for i, host := range hosts {
		if i > 0 {
			hostStr += ","
		}
		hostStr += host
	}
	return hostStr
}
