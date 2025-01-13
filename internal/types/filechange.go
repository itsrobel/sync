package types

// TODO: when a file is change it can write a change log and then
// write to the file to update
type FileChange struct {
	Type      string // "add" or "remove"
	Content   string
	Position  int    // Line number where change occurred
	VersionId string `bson:"version_id"` // Unique ID for the file
}
