package models

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/golang-module/carbon/v2"
	upperDB "github.com/upper/db/v4"
)

const postsTitleIndex = "posts_title_key"

var (
	// ErrDuplicateTitle ...
	ErrDuplicateTitle = errors.New("Title already exists")
	// ErrDuplicateVote ...
	ErrDuplicateVote = errors.New("You have already voted for this post")
	// ErrDuplicatePost ...
	ErrDuplicatePost = errors.New("Post with same title already exists")

	queryTemplate = `
	SELECT COUNT(*) OVER() AS total_records, pq.*, u.username AS uname FROM (
		SELECT p.id, p.title, p.url, p.created_at, p.user_id as uid, COUNT(c.post_id) as comment_count, count(v.post_id) as votes
		FROM posts p
		LEFT JOIN comments c ON p.id = c.post_id
		LEFT JOIN votes v ON p.id = v.post_id
		#where#
		GROUP BY p.id
		#orderby#
	) AS pq
	LEFT JOIN users u ON u.id = uid
	#limit#
	`
)

// Posts is the struct for posts table in DB
type Posts struct {
	ID           int       `db:"id,omitempty"`
	Title        string    `db:"title"`
	URL          string    `db:"url"`
	CreatedAt    time.Time `db:"created_at"`
	UserID       int       `db:"user_id"`
	Username     string    `db:"username,omitempty"`
	CommentCount int       `db:"comment_count,omitempty"`
	TotalRecords int       `db:"total_records,omitempty"`
	Votes        int       `db:"votes,omitempty"`
}

// PostsModel ...
type PostsModel struct {
	db upperDB.Session
}

// Table ...
func (pm PostsModel) Table() string {
	return "posts"
}

// GetByID ...
func (pm PostsModel) GetByID(id int) (*Posts, error) {
	var post Posts
	query := strings.Replace(queryTemplate, "#where#", "WHERE p.id = $1", 1)
	query = strings.Replace(query, "#orderby#", "", 1)
	query = strings.Replace(query, "#limit#", "", 1)
	row, err := pm.db.SQL().Query(query, id)
	if err != nil {
		return nil, err
	}
	iter := pm.db.SQL().NewIterator(row)
	err = iter.One(&post)
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetPosts ...
func (pm PostsModel) GetPosts(f Filters) ([]Posts, MetaData, error) {
	var posts []Posts
	var rows *sql.Rows
	var err error
	var meta MetaData

	query := f.applyTemplate(queryTemplate)
	// fmt.Println("query to be run", query)
	if len(f.Query) > 0 {
		// if search post by title
		rows, err = pm.db.SQL().Query(query, "%"+strings.ToLower(f.Query)+"%", f.limit(), f.offset())
	} else {
		// else just display all posts
		rows, err = pm.db.SQL().Query(query, f.limit(), f.offset())
	}
	if err != nil {
		return nil, meta, err
	}

	iter := pm.db.SQL().NewIterator(rows)
	err = iter.All(&posts)
	if err != nil {
		return nil, meta, err
	}
	if len(posts) == 0 {
		// no rows returned
		return nil, meta, nil // if no posts, return an empty page
	}

	return posts, calculateMetaData(posts[0].TotalRecords, f.Page, f.PageSize), err
}

// AddVote to vote for a post by a user
func (pm PostsModel) AddVote(postID, userID int) error {
	col := pm.db.Collection("votes")
	_, err := col.Insert(map[string]int{
		"post_id": postID,
		"user_id": userID,
	})
	if err != nil {
		fmt.Println("DEBUG: error", err.Error())
		if errHasDuplicate(err, "votes_pkey") { // you can get this key name by describing the votes table in DB
			fmt.Println("DEBUG: returning dup error")
			return ErrDuplicateVote
		}
	}
	return nil
}

// Insert ...
func (pm PostsModel) Insert(title, url string, userID int) (*Posts, error) {
	post := Posts{
		CreatedAt: time.Now(),
		Title:     title,
		URL:       url,
		UserID:    userID,
	}
	col := pm.db.Collection(pm.Table())
	res, err := col.Insert(post)
	if err != nil {
		switch {
		case errHasDuplicate(err, postsTitleIndex):
			return nil, ErrDuplicatePost
		default:
			return nil, err
		}
	}
	post.ID = convertUpperIDToInt(res.ID())
	return &post, nil
}

// GetHumanPostDate gives posted date like "10 minutes ago"
func (p *Posts) GetHumanPostDate() string {
	return carbon.CreateFromStdTime(p.CreatedAt).DiffForHumans()
}

// GetHost returns the hostname of this post url
func (p *Posts) GetHost() string {
	u, err := url.Parse(p.URL)
	if err != nil {
		return ""
	}
	return u.Host
}
