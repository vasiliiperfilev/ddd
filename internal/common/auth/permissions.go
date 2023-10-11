package auth

import "golang.org/x/exp/slices"

type Role int

const (
	TrainerRole Role = iota
	AttendeeRole
)

type Permission int

const (
	ChangeTrainerHoursPermission Permission = iota
	SignForTrainingPermission
)

type permissionsMap map[Role][]Permission

var PermissionsMap permissionsMap = permissionsMap{
	TrainerRole:  {ChangeTrainerHoursPermission},
	AttendeeRole: {SignForTrainingPermission},
}

func (pm permissionsMap) HasPermission(r Role, p Permission) bool {
	return slices.Contains(pm[r], p)
}
