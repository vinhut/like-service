package models

import (
	"fmt"
	"time"

	"github.com/vinhut/like-service/helpers"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LikeDatabase interface {
	FindPost(string) (int, error)
	PostIsLiked(string, string) (bool, error)
	CreatePostLike(PostLike) (bool, error)
	DeletePostLike(string, string) (bool, error)
	FindComment(string) (int, error)
	CommentIsLiked(string, string) (bool, error)
	CreateCommentLike(CommentLike) (bool, error)
	DeleteCommentLike(string, string) (bool, error)
	FindUserLike(string) ([]string, error)
}

type likeDatabase struct {
	db helpers.DatabaseHelper
}

type PostLike struct {
	Likeid  primitive.ObjectID `bson:"_id, omitempty"`
	Uid     string
	Postid  string
	Created time.Time
}

type CommentLike struct {
	Likeid    primitive.ObjectID `bson:"_id, omitempty"`
	Uid       string
	Commentid string
	Created   time.Time
}

func NewLikeDatabase(db helpers.DatabaseHelper) LikeDatabase {
	return &likeDatabase{
		db: db,
	}
}

func (likedb *likeDatabase) FindPost(postid string) (int, error) {

	result, err := likedb.db.QueryAll("postlike", "postid", postid, PostLike{})
	if err != nil {
		fmt.Println("model find error ", err)
		return 0, err
	}

	result_count := len(result)
	return result_count, nil
}

func (likedb *likeDatabase) PostIsLiked(postid string, userid string) (bool, error) {
	query := map[string]string{
		"postid": postid,
		"uid":    userid,
	}
	postdata := PostLike{}
	query_err := likedb.db.Query("postlike", query, &postdata)
	if query_err != nil {
		return false, query_err
	}
	return true, nil

}

func (likedb *likeDatabase) CreatePostLike(post PostLike) (bool, error) {

	query := map[string]string{
		"postid": post.Postid,
		"userid": post.Uid,
	}
	err := likedb.db.Upsert("postlike", query, post)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (likedb *likeDatabase) DeletePostLike(postid string, userid string) (bool, error) {

	query := map[string]string{
		"postid": postid,
		"userid": userid,
	}

	err := likedb.db.Delete("postlike", query)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (likedb *likeDatabase) FindComment(commentid string) (int, error) {

	result, err := likedb.db.QueryAll("commentlike", "commentid", commentid, CommentLike{})
	if err != nil {
		fmt.Println("model find error ", err)
		return 0, err
	}

	result_count := len(result)
	return result_count, nil
}

func (likedb *likeDatabase) CommentIsLiked(commentid string, userid string) (bool, error) {
	query := map[string]string{
		"commentid": commentid,
		"uid":       userid,
	}

	commentdata := CommentLike{}
	query_err := likedb.db.Query("commentlike", query, &commentdata)
	if query_err != nil {
		return false, query_err
	}
	return true, nil

}

func (likedb *likeDatabase) CreateCommentLike(comment CommentLike) (bool, error) {

	query := map[string]string{
		"postid": comment.Commentid,
		"userid": comment.Uid,
	}
	err := likedb.db.Upsert("commentlike", query, comment)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (likedb *likeDatabase) DeleteCommentLike(commentid string, userid string) (bool, error) {

	query := map[string]string{
		"commentid": commentid,
		"userid":    userid,
	}

	err := likedb.db.Delete("commentlike", query)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (likedb *likeDatabase) FindUserLike(userid string) ([]string, error) {

	var result_str []string
	result, err := likedb.db.QueryAll("postlike", "userid", userid, PostLike{})
	if err != nil {
		fmt.Println("model find error ", err)
		return nil, err
	}

	results := make([]PostLike, len(result))

	for i, d := range result {
		if d == nil {
			fmt.Println("d is nil")
		}
		fmt.Println("d = ", d)
		results[i] = d.(PostLike)
	}

	for _, postlike := range results {
		result_str = append(result_str, postlike.Uid)
	}

	return result_str, nil
}
