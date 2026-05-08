package state

import "github.com/google/uuid"

const uuidKey = "uuid"

func GetOrCreateUUID(s *State) (uuid.UUID, error) {
	value, err := s.GetOrComputeValue(uuidKey, func() (string, error) {
		return uuid.NewString(), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.Parse(value)
}
