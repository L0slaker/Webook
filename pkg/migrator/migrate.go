package migrator

type Entity interface {
	ID() int64
	CompareTo(Entity) bool
}
