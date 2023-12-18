package models

import (
	"fmt"
	"strings"

	upperDB "github.com/upper/db/v4"
)

// Models ...
type Models struct {
	Users    UsersModel
	Posts    PostsModel
	Comments CommentsModel
}

// NewModel ...
func NewModel(db upperDB.Session) Models {
	return Models{
		Users: UsersModel{
			db: db,
		},
		Posts: PostsModel{
			db: db,
		},
		Comments: CommentsModel{
			db: db,
		},
	}
}

func errHasDuplicate(err error, key string) bool {
	str := fmt.Sprintf(`duplicate key value violates unique constraint "%s"`, key)
	return strings.Contains(err.Error(), str)
}

func convertUpperIDToInt(id upperDB.ID) int {
	idType := fmt.Sprintf("%T", id)
	if idType == "int64" {
		return int(id.(int64))
	}
	return id.(int)
}
