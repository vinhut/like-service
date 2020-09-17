package main

import (
	"github.com/gin-gonic/gin"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	transport "github.com/uber/jaeger-client-go/transport/zipkin"
	"github.com/uber/jaeger-client-go/zipkin"
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
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
	trsport, _ := transport.NewHTTPTransport(
		JAEGER_COLLECTOR_ENDPOINT,
		transport.HTTPLogger(jaeger.StdLogger),
	)
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
	cfg.InitGlobalTracer(
		"like-service",
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
		jaegercfg.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		jaegercfg.ZipkinSharedRPCSpan(true),
		jaegercfg.Reporter(jaeger.NewRemoteReporter(trsport)),
	)
	tracer := opentracing.GlobalTracer()

	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "OK")
	})

	router.GET(SERVICE_NAME+"/postcount", func(c *gin.Context) {

		span := tracer.StartSpan("get postlike count")

		value, cookie_err := c.Cookie("token")
		post_id, _ := c.GetQuery("postid")
		if cookie_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "Unauthorized"})
			return
		}
		_, check_err := checkUser(authservice, value)
		if check_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "Unauthorized"})
			return
		}

		like_count, find_err := likedb.FindPost(post_id)
		if find_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(404, gin.H{"reason": "like not found"})
			return
		}
		c.String(200, strconv.Itoa(like_count))
		span.Finish()

	})

	router.GET(SERVICE_NAME+"/post", func(c *gin.Context) {

		span := tracer.StartSpan("get post like")

		value, cookie_err := c.Cookie("token")
		post_id, _ := c.GetQuery("postid")
		if cookie_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "Unauthorized"})
			return
		}

		user_data, check_err := checkUser(authservice, value)
		if check_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "Unauthorized"})
			return
		}

		isliked, query_err := likedb.PostIsLiked(post_id, user_data.Uid)
		if query_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(404, gin.H{"reason": "like not found"})
			return
		}
		c.String(200, strconv.FormatBool(isliked))
		span.Finish()

	})

	router.POST(SERVICE_NAME+"/post", func(c *gin.Context) {
		span := tracer.StartSpan("like post")

		value, cookie_err := c.Cookie("token")
		post_id, _ := c.GetQuery("postid")
		if cookie_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}
		user_data, check_err := checkUser(authservice, value)
		if check_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}

		new_post_like := models.PostLike{
			Likeid:  primitive.NewObjectIDFromTimestamp(time.Now()),
			Uid:     user_data.Uid,
			Postid:  post_id,
			Created: time.Now(),
		}

		_, create_err := likedb.CreatePostLike(new_post_like)
		if create_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(500, gin.H{"reason": "create like error"})
			return
		}
		c.String(200, "Liked")
		span.Finish()
	})

	router.DELETE(SERVICE_NAME+"/post", func(c *gin.Context) {
		span := tracer.StartSpan("unlike post")

		value, cookie_err := c.Cookie("token")
		post_id, _ := c.GetQuery("postid")
		if cookie_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}
		user_data, check_err := checkUser(authservice, value)
		if check_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}

		_, delete_err := likedb.DeletePostLike(post_id, user_data.Uid)
		if delete_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(500, gin.H{"reason": "delete like error"})
			return
		}
		c.String(200, "deleted")
		span.Finish()
	})

	router.GET(SERVICE_NAME+"/commentcount", func(c *gin.Context) {

		span := tracer.StartSpan("get comment like count")

		value, cookie_err := c.Cookie("token")
		comment_id, _ := c.GetQuery("commentid")
		if cookie_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}
		_, check_err := checkUser(authservice, value)
		if check_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}

		like_count, query_err := likedb.FindComment(comment_id)
		if query_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(404, gin.H{"reason": "comment like not found"})
			return
		}
		c.String(200, strconv.Itoa(like_count))
		span.Finish()

	})

	router.GET(SERVICE_NAME+"/comment", func(c *gin.Context) {

		span := tracer.StartSpan("get comment like")

		value, cookie_err := c.Cookie("token")
		comment_id, _ := c.GetQuery("commentid")
		if cookie_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}
		user_data, check_err := checkUser(authservice, value)
		if check_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}

		isLiked, query_err := likedb.CommentIsLiked(comment_id, user_data.Uid)
		if query_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(404, gin.H{"reason": "comment like not found"})
			return
		}
		c.String(200, strconv.FormatBool(isLiked))
		span.Finish()

	})

	router.POST(SERVICE_NAME+"/comment", func(c *gin.Context) {
		span := tracer.StartSpan("like comment")

		value, cookie_err := c.Cookie("token")
		comment_id, _ := c.GetQuery("commentid")
		if cookie_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}
		user_data, check_err := checkUser(authservice, value)
		if check_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}

		new_comment_like := models.CommentLike{

			Likeid:    primitive.NewObjectIDFromTimestamp(time.Now()),
			Uid:       user_data.Uid,
			Commentid: comment_id,
			Created:   time.Now(),
		}

		_, create_err := likedb.CreateCommentLike(new_comment_like)
		if create_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(500, gin.H{"reason": "create like error"})
			return
		}
		c.String(200, "Liked")
		span.Finish()
	})

	router.DELETE(SERVICE_NAME+"/comment", func(c *gin.Context) {
		span := tracer.StartSpan("unlike comment")

		value, cookie_err := c.Cookie("token")
		comment_id, _ := c.GetQuery("commentid")
		if cookie_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}
		user_data, check_err := checkUser(authservice, value)
		if check_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}

		_, delete_err := likedb.DeleteCommentLike(comment_id, user_data.Uid)
		if delete_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(500, gin.H{"reason": "error delete like"})
			return
		}
		c.String(200, "deleted")
		span.Finish()
	})

	router.GET(SERVICE_NAME+"/user", func(c *gin.Context) {

		span := tracer.StartSpan("get user like")

		value, cookie_err := c.Cookie("token")
		if cookie_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}
		user_data, check_err := checkUser(authservice, value)
		if check_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(401, gin.H{"reason": "unauthorized"})
			return
		}

		user_like, find_err := likedb.FindUserLike(user_data.Uid)
		if find_err != nil {
			span.Finish()
			c.AbortWithStatusJSON(404, gin.H{"reason": "not found"})
			return
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
			span.Finish()
			c.AbortWithStatusJSON(500, gin.H{"reason": "error create post like"})
			return
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
