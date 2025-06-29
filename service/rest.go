package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gitlab.com/devskiller-tasks/rest-api-blog-golang/model"
	"gitlab.com/devskiller-tasks/rest-api-blog-golang/repository"
)

type RestApiService struct {
	postRepository    *repository.PostRepository
	commentRepository *repository.CommentRepository
}

type AckJsonResponse struct {
	Message string
	Status  int
}

func NewRestApiService() RestApiService {
	return RestApiService{postRepository: repository.NewPostRepository(), commentRepository: repository.NewCommentRepository()}
}

func (svc *RestApiService) ServeContent(port int) error {
	http.HandleFunc("POST /api/posts", handleAddPost(svc))
	http.HandleFunc("GET /api/posts/{postId}", handleGetPostByPostId(svc))
	http.HandleFunc("POST /api/comments", handleAddComment(svc))
	http.HandleFunc("GET /api/comments", handleGetCommentsByPostId(svc))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func handleAddPost(svc *RestApiService) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var post model.Post
		if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
			http.Error(w, "400 Bad Request", http.StatusBadRequest)
			return
		}
		if err := svc.postRepository.Insert(post); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		data, err := json.Marshal(&AckJsonResponse{Message: fmt.Sprintf("Post with id: %d successfully added", post.Id), Status: http.StatusOK})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func handleGetPostByPostId(svc *RestApiService) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Example: GET /api/posts/42

		// Every response should have the Content-Type=application/json header set.

		// If an invalid ID is given, the response should be in the format of `AckJsonResponse` with a status of 400 and a message:
		// { "Message": "Wrong id path variable: PATH_VARIABLE", "Status": 400 }
		// The HTTP response code should also be set to 400.

		// If the given postID does not exist, the response should be in the format of `AckJsonResponse` with a status of 404 and a message:
		// { "Message": "Post with id: [POST_ID] does not exist", "Status": 404 }
		// The HTTP response code should also be set to 404.

		// If the post with the given ID exists, the response should be a valid JSON representation of the post entity:
		// { "Id": 2, "Title": "test title", "Content": "this is a post content", "CreationDate": "1970-01-01T03:46:40+01:00" }
		w.Header().Set("Content-Type", "application/json")

		// Extract postId from URL path
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(AckJsonResponse{
				Message: "Wrong id path variable: ",
				Status:  http.StatusBadRequest,
			})
			return
		}

		idStr := parts[3]
		postId, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(AckJsonResponse{
				Message: fmt.Sprintf("Wrong id path variable: %s", idStr),
				Status:  http.StatusBadRequest,
			})
			return
		}

		post, err := svc.postRepository.GetById(postId)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(AckJsonResponse{
				Message: fmt.Sprintf("Post with id: %d does not exist", postId),
				Status:  http.StatusNotFound,
			})
			return
		}

		json.NewEncoder(w).Encode(post)
	}
}

func handleGetCommentsByPostId(svc *RestApiService) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		postIdStr := r.URL.Query().Get("postId")
		postid, err := strconv.ParseInt(postIdStr, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(AckJsonResponse{
				Message: fmt.Sprintf("Wrong id path variable: %d", postid),
				Status:  http.StatusBadRequest,
			})
		}
		comments := svc.commentRepository.GetAllByPostId(uint64(postid))
		json.NewEncoder(w).Encode(comments)
	}
}

func handleAddComment(svc *RestApiService) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var comment model.Comment
		if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(AckJsonResponse{
				Message: "Could not deserialize comment JSON payload",
				Status:  http.StatusBadRequest,
			})
			return
		}
		if comment.Author == "" || comment.Comment == "" || comment.PostId == 0 || comment.Id == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(AckJsonResponse{
				Message: "Could not deserialize comment JSON payload",
				Status:  http.StatusBadRequest,
			})
			return
		}
		if err := svc.commentRepository.Insert(comment); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(AckJsonResponse{
				Message: err.Error(),
				Status:  http.StatusBadRequest,
			})
			return
		}
		json.NewEncoder(w).Encode(AckJsonResponse{
			Message: fmt.Sprintf("Comment with id: %d successfully added", comment.Id),
			Status:  http.StatusOK,
		})
	}
}
