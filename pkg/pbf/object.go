package pbf

type Object interface {
	ObjectID() ObjectID
}

type ObjectID int64
