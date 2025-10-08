package lib

import "github.com/google/uuid"

func GenerateUUID() []byte {
	id := uuid.New()
	return id[:]
}
