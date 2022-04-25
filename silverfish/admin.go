package silverfish

import (
	"silverfish/silverfish/entity"
)

// Admin export
type Admin struct {
	userInf *entity.MongoInf
}

// NewAdmin export
func NewAdmin(userInf *entity.MongoInf) *Admin {
	a := new(Admin)
	a.userInf = userInf
	return a
}
