package util

import "uuid"

func IsValidUUID(str string) bool {
	_, err := uuid.Parse(str)
	return err == nil
}
