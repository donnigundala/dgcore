package mongo

// Migration represents a migration in the database.
type Migration struct {
	Name  string `bson:"name"`
	Batch int    `bson:"batch"`
}
