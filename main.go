package main

import (
	"github.com/gin-gonic/gin"
	opentracing "github.com/opentracing/opentracing-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
	"github.com/vinhut/like-service/helpers"
	"github.com/vinhut/like-service/models"
	"github.com/vinhut/like-service/services"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"encoding/json"
	"os"
	"strconv"
	"time"
)

var SERVICE_NAME = "like-service"

type UserAuthData struct {
	Uid     string
	Email   string
	Role    string
	Created string
}

func checkUser(authservice services.AuthService, token string) (*UserAuthData, error) {

	data := &UserAuthData{}
	user_data, auth_error := authservice.Check(SERVICE_NAME, token)
	if auth_error != nil {
		return data, auth_error
	}

	if err := json.Unmarshal([]byte(user_data), data); err != nil {
		panic(err)
		return data, err
	}

	return data, nil

}

func setupRouter(likedb models.LikeDatabase, authservice services.AuthService) *gin.Engine {

	var JAEGER_COLLECTOR_ENDPOINT = os.Getenv("JAEGER_COLLECTOR_ENDPOINT")
	cfg := jaegercfg.Configuration{
		ServiceName: "like-service",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:          true,
			CollectorEndpoint: JAEGER_COLLECTOR_ENDPOINT,
		},
	}
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory
	tracer, _, _ := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	opentracing.SetGlobalTracer(tracer)

	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "OK")
	})

	router.GET(SERVICE_NAME+"/postcount", func(c *gin.Context) {

		span := tracer.StartSpan("get postlike count")

		value, err := c.Cookie("token")
		post_id, _ := c.GetQuery("postid")
		if err != nil {
			panic("failed get token")
		}
		_, check_err := checkUser(authservice, value)
		if check_err != nil {
			panic("error check user")
		}

		like_count, find_err := likedb.FindPost(post_id)
		if find_err != nil {
			panic(find_err)
		}
		c.String(200, strconv.Itoa(like_count))
		span.Finish()

	})

	router.GET(SERVICE_NAME+"/post", func(c *gin.Context) {

		span := tracer.StartSpan("get post like")

		value, err := c.Cookie("token")
		post_id, _ := c.GetQuery("postid")
		if err != nil {
			panic("failed get token")
		}

		user_data, check_err := checkUser(authservice, value)
		if check_err != nil {
			panic("error check user")
		}

		isliked, query_err := likedb.PostIsLiked(post_id, user_data.Uid)
		if query_err != nil {
			panic(query_err)
		}
		c.String(200, strconv.FormatBool(isliked))
		span.Finish()

	})

	router.POST(SERVICE_NAME+"/post", func(c *gin.Context) {
		span := tracer.StartSpan("like post")

		value, err := c.Cookie("token")
		post_id, _ := c.GetQuery("postid")
		if err != nil {
			panic("failed get token")
		}
		user_data, check_err := checkUser(authservice, value)
		if check_err != nil {
			panic("error check user")
		}

		new_post_like := models.PostLike{
			Likeid:  primitive.NewObjectIDFromTimestamp(time.Now()),
			Uid:     user_data.Uid,
			Postid:  post_id,
			Created: time.Now(),
		}

		_, create_err := likedb.CreatePostLike(new_post_like)
		if create_err != nil {
			panic(create_err)
		}
		c.String(200, "Liked")
		span.Finish()
	})

	router.DELETE(SERVICE_NAME+"/post", func(c *gin.Context) {
		span := tracer.StartSpan("unlike post")

		value, err := c.Cookie("token")
		post_id, _ := c.GetQuery("postid")
		if err != nil {
			panic("failed get token")
		}
		user_data, check_err := checkUser(authservice, value)
		if check_err != nil {
			panic("error check user")
		}

		_, delete_err := likedb.DeletePostLike(post_id, user_data.Uid)
		if delete_err != nil {
			panic(delete_err)
		}
		c.String(200, "deleted")
		span.Finish()
	})

	router.GET(SERVICE_NAME+"/commentcount", func(c *gin.Context) {

		span := tracer.StartSpan("get comment like count")

		value, err := c.Cookie("token")
		comment_id, _ := c.GetQuery("commentid")
		if err != nil {
			panic("failed get token")
		}
		_, check_err := checkUser(authservice, value)
		if check_err != nil {
			panic("error check user")
		}

		like_count, query_err := likedb.FindComment(comment_id)
		if query_err != nil {
			panic(query_err)
		}
		c.String(200, strconv.Itoa(like_count))
		span.Finish()

	})

	router.GET(SERVICE_NAME+"/comment", func(c *gin.Context) {

		span := tracer.StartSpan("get comment like")

		value, err := c.Cookie("token")
		comment_id, _ := c.GetQuery("commentid")
		if err != nil {
			panic("failed get token")
		}
		user_data, check_err := checkUser(authservice, value)
		if check_err != nil {
			panic("error check user")
		}

		isLiked, query_err := likedb.CommentIsLiked(comment_id, user_data.Uid)
		if query_err != nil {
			panic(query_err)
		}
		c.String(200, strconv.FormatBool(isLiked))
		span.Finish()

	})

	router.POST(SERVICE_NAME+"/comment", func(c *gin.Context) {
		span := tracer.StartSpan("like comment")

		value, err := c.Cookie("token")
		comment_id, _ := c.GetQuery("commentid")
		if err != nil {
			panic("failed get token")
		}
		user_data, check_err := checkUser(authservice, value)
		if check_err != nil {
			panic("error check user")
		}

		new_comment_like := models.CommentLike{

			Likeid:    primitive.NewObjectIDFromTimestamp(time.Now()),
			Uid:       user_data.Uid,
			Commentid: comment_id,
			Created:   time.Now(),
		}

		_, create_err := likedb.CreateCommentLike(new_comment_like)
		if create_err != nil {
			panic(create_err)
		}
		c.String(200, "Liked")
		span.Finish()
	})

	router.DELETE(SERVICE_NAME+"/comment", func(c *gin.Context) {
		span := tracer.StartSpan("unlike comment")

		value, err := c.Cookie("token")
		comment_id, _ := c.GetQuery("commentid")
		if err != nil {
			panic("failed get token")
		}
		user_data, check_err := checkUser(authservice, value)
		if check_err != nil {
			panic("error check user")
		}

		_, delete_err := likedb.DeleteCommentLike(comment_id, user_data.Uid)
		if delete_err != nil {
			panic(delete_err)
		}
		c.String(200, "deleted")
		span.Finish()
	})

	router.GET(SERVICE_NAME+"/user", func(c *gin.Context) {

		span := tracer.StartSpan("get user like")

		value, err := c.Cookie("token")
		if err != nil {
			panic("failed get token")
		}
		user_data, check_err := checkUser(authservice, value)
		if check_err != nil {
			panic("error check user")
		}

		user_like, find_err := likedb.FindUserLike(user_data.Uid)
		if find_err != nil {
			panic(find_err)
		}
		result, marshal_err := json.Marshal(user_like)
		if marshal_err != nil {
			panic(marshal_err)
		}
		c.String(200, string(result))
		span.Finish()

	})

	// internal endpoint

	router.POST("internal/post", func(c *gin.Context) {
		span := tracer.StartSpan("internal generate like post")

		uid, _ := c.GetQuery("uid")
		post_id, _ := c.GetQuery("postid")

		new_post_like := models.PostLike{
			Likeid:  primitive.NewObjectIDFromTimestamp(time.Now()),
			Uid:     uid,
			Postid:  post_id,
			Created: time.Now(),
		}

		_, create_err := likedb.CreatePostLike(new_post_like)
		if create_err != nil {
			panic(create_err)
		}
		c.String(200, "Liked")
		span.Finish()
	})

	return router
}

func main() {

	mongo_layer := helpers.NewMongoDatabase()
	likedb := models.NewLikeDatabase(mongo_layer)
	authservice := services.NewUserAuthService()
	router := setupRouter(likedb, authservice)
	router.Run(":8080")

}
