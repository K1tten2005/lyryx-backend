package storage

import "github.com/lib/pq"

func isUniqueViolation(err error, constraint string) bool {
	pqErr, ok := err.(*pq.Error)
	if !ok {
		return false
	}

	if pqErr.Code != "23505" {
		return false
	}

	return pqErr.Constraint == constraint
}
