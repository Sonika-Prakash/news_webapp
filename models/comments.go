package models

import (
	"time"

	"github.com/golang-module/carbon/v2"
	"github.com/upper/db/v4"
)

// Comments ...
type Comments struct {
	ID        int       `db:"comment_id,omitempty"`
	CreatedAt time.Time `db:"comment_created_at,omitempty"`
	Body      string    `db:"body"`
	PostID    int       `db:"post_id"`
	UserID    int       `db:"user_id"`
	Users     `db:",inline"`
}

// CommentsModel ...
type CommentsModel struct {
	db db.Session
}

// Table ...
func (cm CommentsModel) Table() string {
	return "comments"
}

// GetCommentsForPost ...
func (cm CommentsModel) GetCommentsForPost(postID int) ([]Comments, error) {
	var comments []Comments
	query := cm.db.SQL().Select("c.id AS comment_id", "c.created_at AS comment_created_at", "*").From(cm.Table() + " AS c").Join("users AS u").On("c.user_id = u.id").Where(db.Cond{"c.post_id": postID}).OrderBy("c.created_at DESC")

	err := query.All(&comments)
	if err != nil {
		return nil, err
	}

	return comments, nil
}

// Insert ...
func (cm CommentsModel) Insert(body string, postID int, userID int) error {
	_, err := cm.db.Collection(cm.Table()).Insert(map[string]interface{}{
		"created_at": time.Now(),
		"body":       body,
		"user_id":    userID,
		"post_id":    postID,
	})
	if err != nil {
		return err
	}
	return nil
}

// GetHumanCommentDate ...
func (c *Comments) GetHumanCommentDate() string {
	return carbon.CreateFromStdTime(c.CreatedAt).DiffForHumans()
}
