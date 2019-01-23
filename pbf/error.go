package pbf

type Error struct{}

func (e *Error) Error() string {
	return ""
}
