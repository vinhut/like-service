package main

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	mocks_models "github.com/vinhut/like-service/mocks_models"
	mocks_services "github.com/vinhut/like-service/mocks_services"

	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestPing(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_like := mocks_models.NewMockLikeDatabase(ctrl)
	mock_auth := mocks_services.NewMockAuthService(ctrl)

	router := setupRouter(mock_like, mock_auth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestCheckUser(t *testing.T) {

	now := time.Now()
	token := "852a37a34b727c0e0b331806-7af4bdfdcc60990d427f383efecc8529289d040dd67e0753b9e2ee5a1e938402186f28324df23f6faa4e2bbf43f584ae228c55b00143866215d6e92805d470a1cc2a096dcca4d43527598122313be412e17fbefdcdab2fae02e06a405791d936862d4fba688b3c7fd784d4"
	user_data := "{\"uid\": \"1\", \"email\": \"test@email.com\", \"role\": \"standard\", \"created\": \"" + now.Format("2006-01-02T15:04:05") + "\"}"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_auth := mocks_services.NewMockAuthService(ctrl)
	mock_auth.EXPECT().Check(gomock.Any(), gomock.Any()).Return(user_data, nil)

	data, _ := checkUser(mock_auth, token)
	test_data := &UserAuthData{}

	if err := json.Unmarshal([]byte(user_data), test_data); err != nil {
		fmt.Println(err)
	}

	assert.Equal(t, test_data, data)
}

func TestGetPostLikeCount(t *testing.T) {

	now := time.Now()
	token := "852a37a34b727c0e0b331806-7af4bdfdcc60990d427f383efecc8529289d040dd67e0753b9e2ee5a1e938402186f28324df23f6faa4e2bbf43f584ae228c55b00143866215d6e92805d470a1cc2a096dcca4d43527598122313be412e17fbefdcdab2fae02e06a405791d936862d4fba688b3c7fd784d4"
	user_data := "{\"uid\": \"1\", \"email\": \"test@email.com\", \"role\": \"standard\", \"created\": \"" + now.Format("2006-01-02T15:04:05") + "\"}"
	postid := "1"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_like := mocks_models.NewMockLikeDatabase(ctrl)
	mock_auth := mocks_services.NewMockAuthService(ctrl)

	mock_auth.EXPECT().Check(gomock.Any(), gomock.Any()).Return(user_data, nil)
	mock_like.EXPECT().FindPost(gomock.Any()).Return(1, nil)

	router := setupRouter(mock_like, mock_auth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", SERVICE_NAME+"/postcount?postid="+postid, nil)
	req.Header.Set("Cookie", "token="+token+";")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

}

func TestGetPostLikeStatus(t *testing.T) {

	now := time.Now()
	token := "852a37a34b727c0e0b331806-7af4bdfdcc60990d427f383efecc8529289d040dd67e0753b9e2ee5a1e938402186f28324df23f6faa4e2bbf43f584ae228c55b00143866215d6e92805d470a1cc2a096dcca4d43527598122313be412e17fbefdcdab2fae02e06a405791d936862d4fba688b3c7fd784d4"
	user_data := "{\"uid\": \"1\", \"email\": \"test@email.com\", \"role\": \"standard\", \"created\": \"" + now.Format("2006-01-02T15:04:05") + "\"}"
	postid := "1"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_like := mocks_models.NewMockLikeDatabase(ctrl)
	mock_auth := mocks_services.NewMockAuthService(ctrl)

	mock_auth.EXPECT().Check(gomock.Any(), gomock.Any()).Return(user_data, nil)
	mock_like.EXPECT().PostIsLiked(gomock.Any(), gomock.Any()).Return(true, nil)

	router := setupRouter(mock_like, mock_auth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", SERVICE_NAME+"/post?postid="+postid, nil)
	req.Header.Set("Cookie", "token="+token+";")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

}

func TestCreatePostLike(t *testing.T) {

	now := time.Now()
	token := "852a37a34b727c0e0b331806-7af4bdfdcc60990d427f383efecc8529289d040dd67e0753b9e2ee5a1e938402186f28324df23f6faa4e2bbf43f584ae228c55b00143866215d6e92805d470a1cc2a096dcca4d43527598122313be412e17fbefdcdab2fae02e06a405791d936862d4fba688b3c7fd784d4"
	user_data := "{\"uid\": \"1\", \"email\": \"test@email.com\", \"role\": \"standard\", \"created\": \"" + now.Format("2006-01-02T15:04:05") + "\"}"
	postid := "1"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_like := mocks_models.NewMockLikeDatabase(ctrl)
	mock_auth := mocks_services.NewMockAuthService(ctrl)

	mock_auth.EXPECT().Check(gomock.Any(), gomock.Any()).Return(user_data, nil)
	mock_like.EXPECT().CreatePostLike(gomock.Any()).Return(true, nil)

	router := setupRouter(mock_like, mock_auth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", SERVICE_NAME+"/post?postid="+postid, nil)
	req.Header.Set("Cookie", "token="+token+";")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

}

func TestDeletePostLike(t *testing.T) {

	now := time.Now()
	token := "852a37a34b727c0e0b331806-7af4bdfdcc60990d427f383efecc8529289d040dd67e0753b9e2ee5a1e938402186f28324df23f6faa4e2bbf43f584ae228c55b00143866215d6e92805d470a1cc2a096dcca4d43527598122313be412e17fbefdcdab2fae02e06a405791d936862d4fba688b3c7fd784d4"
	user_data := "{\"uid\": \"1\", \"email\": \"test@email.com\", \"role\": \"standard\", \"created\": \"" + now.Format("2006-01-02T15:04:05") + "\"}"
	postid := "1"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_like := mocks_models.NewMockLikeDatabase(ctrl)
	mock_auth := mocks_services.NewMockAuthService(ctrl)

	mock_auth.EXPECT().Check(gomock.Any(), gomock.Any()).Return(user_data, nil)
	mock_like.EXPECT().DeletePostLike(gomock.Any(), gomock.Any()).Return(true, nil)

	router := setupRouter(mock_like, mock_auth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", SERVICE_NAME+"/post?postid="+postid, nil)
	req.Header.Set("Cookie", "token="+token+";")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

}

func TestGetCommentLikeCount(t *testing.T) {

	now := time.Now()
	token := "852a37a34b727c0e0b331806-7af4bdfdcc60990d427f383efecc8529289d040dd67e0753b9e2ee5a1e938402186f28324df23f6faa4e2bbf43f584ae228c55b00143866215d6e92805d470a1cc2a096dcca4d43527598122313be412e17fbefdcdab2fae02e06a405791d936862d4fba688b3c7fd784d4"
	user_data := "{\"uid\": \"1\", \"email\": \"test@email.com\", \"role\": \"standard\", \"created\": \"" + now.Format("2006-01-02T15:04:05") + "\"}"
	commentid := "1"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_like := mocks_models.NewMockLikeDatabase(ctrl)
	mock_auth := mocks_services.NewMockAuthService(ctrl)

	mock_auth.EXPECT().Check(gomock.Any(), gomock.Any()).Return(user_data, nil)
	mock_like.EXPECT().FindComment(gomock.Any()).Return(1, nil)

	router := setupRouter(mock_like, mock_auth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", SERVICE_NAME+"/commentcount?commentid="+commentid, nil)
	req.Header.Set("Cookie", "token="+token+";")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

}

func TestGetCommentLikeStatus(t *testing.T) {

	now := time.Now()
	token := "852a37a34b727c0e0b331806-7af4bdfdcc60990d427f383efecc8529289d040dd67e0753b9e2ee5a1e938402186f28324df23f6faa4e2bbf43f584ae228c55b00143866215d6e92805d470a1cc2a096dcca4d43527598122313be412e17fbefdcdab2fae02e06a405791d936862d4fba688b3c7fd784d4"
	user_data := "{\"uid\": \"1\", \"email\": \"test@email.com\", \"role\": \"standard\", \"created\": \"" + now.Format("2006-01-02T15:04:05") + "\"}"
	commentid := "1"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_like := mocks_models.NewMockLikeDatabase(ctrl)
	mock_auth := mocks_services.NewMockAuthService(ctrl)

	mock_auth.EXPECT().Check(gomock.Any(), gomock.Any()).Return(user_data, nil)
	mock_like.EXPECT().CommentIsLiked(gomock.Any(), gomock.Any()).Return(true, nil)

	router := setupRouter(mock_like, mock_auth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", SERVICE_NAME+"/comment?commentid="+commentid, nil)
	req.Header.Set("Cookie", "token="+token+";")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

}

func TestPostCommentLike(t *testing.T) {

	now := time.Now()
	token := "852a37a34b727c0e0b331806-7af4bdfdcc60990d427f383efecc8529289d040dd67e0753b9e2ee5a1e938402186f28324df23f6faa4e2bbf43f584ae228c55b00143866215d6e92805d470a1cc2a096dcca4d43527598122313be412e17fbefdcdab2fae02e06a405791d936862d4fba688b3c7fd784d4"
	user_data := "{\"uid\": \"1\", \"email\": \"test@email.com\", \"role\": \"standard\", \"created\": \"" + now.Format("2006-01-02T15:04:05") + "\"}"
	commentid := "1"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_like := mocks_models.NewMockLikeDatabase(ctrl)
	mock_auth := mocks_services.NewMockAuthService(ctrl)

	mock_auth.EXPECT().Check(gomock.Any(), gomock.Any()).Return(user_data, nil)
	mock_like.EXPECT().CreateCommentLike(gomock.Any()).Return(true, nil)

	router := setupRouter(mock_like, mock_auth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", SERVICE_NAME+"/comment?commentid="+commentid, nil)
	req.Header.Set("Cookie", "token="+token+";")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

}

func TestGetUserLike(t *testing.T) {

	now := time.Now()
	token := "852a37a34b727c0e0b331806-7af4bdfdcc60990d427f383efecc8529289d040dd67e0753b9e2ee5a1e938402186f28324df23f6faa4e2bbf43f584ae228c55b00143866215d6e92805d470a1cc2a096dcca4d43527598122313be412e17fbefdcdab2fae02e06a405791d936862d4fba688b3c7fd784d4"
	user_data := "{\"uid\": \"1\", \"email\": \"test@email.com\", \"role\": \"standard\", \"created\": \"" + now.Format("2006-01-02T15:04:05") + "\"}"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_like := mocks_models.NewMockLikeDatabase(ctrl)
	mock_auth := mocks_services.NewMockAuthService(ctrl)

	mock_auth.EXPECT().Check(gomock.Any(), gomock.Any()).Return(user_data, nil)
	mock_like.EXPECT().FindUserLike(gomock.Any()).Return(make([]string, 1), nil)

	router := setupRouter(mock_like, mock_auth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", SERVICE_NAME+"/user", nil)
	req.Header.Set("Cookie", "token="+token+";")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

}

func TestDeleteCommentLike(t *testing.T) {

	now := time.Now()
	token := "852a37a34b727c0e0b331806-7af4bdfdcc60990d427f383efecc8529289d040dd67e0753b9e2ee5a1e938402186f28324df23f6faa4e2bbf43f584ae228c55b00143866215d6e92805d470a1cc2a096dcca4d43527598122313be412e17fbefdcdab2fae02e06a405791d936862d4fba688b3c7fd784d4"
	user_data := "{\"uid\": \"1\", \"email\": \"test@email.com\", \"role\": \"standard\", \"created\": \"" + now.Format("2006-01-02T15:04:05") + "\"}"
	commentid := "1"

	os.Setenv("KEY", "12345678901234567890123456789012")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock_like := mocks_models.NewMockLikeDatabase(ctrl)
	mock_auth := mocks_services.NewMockAuthService(ctrl)

	mock_auth.EXPECT().Check(gomock.Any(), gomock.Any()).Return(user_data, nil)
	mock_like.EXPECT().DeleteCommentLike(gomock.Any(), gomock.Any()).Return(true, nil)

	router := setupRouter(mock_like, mock_auth)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", SERVICE_NAME+"/comment?commentid="+commentid, nil)
	req.Header.Set("Cookie", "token="+token+";")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

}
