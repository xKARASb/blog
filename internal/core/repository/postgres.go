package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/xkarasb/blog/internal/core/dto"
	"github.com/xkarasb/blog/pkg/db/postgres"
	"github.com/xkarasb/blog/pkg/errors"
	"github.com/xkarasb/blog/pkg/types"
)

type PostgresRepository struct {
	DB *postgres.DB
}

func NewBlogRepository(db *postgres.DB) *PostgresRepository {
	return &PostgresRepository{
		DB: db,
	}
}

func (rep *PostgresRepository) AddNewUser(email, password_hash, role, refreshToken string) (*dto.UserDB, error) {
	user := &dto.UserDB{}

	query := `INSERT INTO users (email, password_hash, role, refresh_token, refresh_token_expiry_time) VALUES ($1, $2, $3, $4, $5) RETURNING *;`
	refreshTokenExpire := time.Now().Add(time.Duration(time.Hour * 24 * 7))

	err := rep.DB.Get(user, query, email, password_hash, role, refreshToken, refreshTokenExpire)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			switch pgErr.Code {
			case "23505":
				return nil, errors.ErrorRepositoryUserAlreadyExsist
			case "23514":
				return nil, errors.ErrorRepositoryBadRole
			}
		}

		return nil, err
	}
	return user, nil
}

func (rep *PostgresRepository) GetUserByEmail(email string) (*dto.UserDB, error) {
	user := &dto.UserDB{}

	query := `SELECT * FROM users WHERE email = $1;`
	err := rep.DB.Get(user, query, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (rep *PostgresRepository) GetUserById(id uuid.UUID) (*dto.UserDB, error) {
	user := &dto.UserDB{}

	query := `SELECT * FROM users WHERE user_id = $1;`
	err := rep.DB.Get(user, query, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (rep *PostgresRepository) UpdateRefreshToken(id uuid.UUID, refreshToken string) (*dto.UserDB, error) {
	user := &dto.UserDB{}

	query := `UPDATE users SET refresh_token = $2 WHERE user_id = $1 RETURNING *;`
	err := rep.DB.Get(user, query, id, refreshToken)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (rep *PostgresRepository) GetRefreshToken(id uuid.UUID) (string, error) {
	var token string
	query := `SELECT refresh_token FROM users WHERE user_id = $1;`
	err := rep.DB.Get(&token, query, id)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (rep *PostgresRepository) GetPostByIdempotencyKey(idempotencyKey string) (*dto.PostDB, error) {
	post := &dto.PostDB{}

	query := `SELECT * FROM posts WHERE idempotency_key = $1;`
	err := rep.DB.Get(post, query, idempotencyKey)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (rep *PostgresRepository) CreatePost(
	authorId uuid.UUID, idempotencyKey, title, content string) (*dto.PostDB, error) {
	post := &dto.PostDB{}

	query := `INSERT INTO posts (author_id, idempotency_key, title, content) VALUES ($1, $2, $3, $4) RETURNING *;`

	err := rep.DB.Get(post, query, authorId, idempotencyKey, title, content)
	if err != nil {
		pgErr, ok := err.(*pq.Error)
		if ok && pgErr.Code == "23505" {
			return nil, errors.ErrorRepositoryUserAlreadyExsist
		}
		return nil, err
	}
	return post, nil
}

func (rep *PostgresRepository) GetPostById(id uuid.UUID) (*dto.PostDB, error) {
	post := &dto.PostDB{}

	query := `SELECT * FROM posts WHERE post_id = $1;`
	err := rep.DB.Get(post, query, id)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (rep *PostgresRepository) UpdatePost(id uuid.UUID, title, content string, status types.PostStatus) (*dto.PostDB, error) {
	post := &dto.PostDB{}
	query := `UPDATE posts SET title = $2, content = $3, status = $4 WHERE post_id = $1 RETURNING *;`
	err := rep.DB.Get(post, query, id, title, content, status)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (rep *PostgresRepository) CreateImage(imageId, postId uuid.UUID, imageUrl string) (*dto.ImageDB, error) {
	image := &dto.ImageDB{}

	query := `INSERT INTO images (image_id, post_id, image_url) VALUES ($1, $2, $3) RETURNING *;`

	err := rep.DB.Get(image, query, imageId, postId, imageUrl)
	if err != nil {
		pgErr, ok := err.(*pq.Error)
		if ok && pgErr.Code == "23505" {
			return nil, errors.ErrorRepositoryUserAlreadyExsist
		}
		return nil, err
	}
	return image, nil
}

func (rep *PostgresRepository) DeleteImage(imageId uuid.UUID) (*dto.ImageDB, error) {
	image := &dto.ImageDB{}
	query := `DELETE FROM images WHERE image_id = $1 RETURNING *`
	err := rep.DB.Get(image, query, imageId)
	if err != nil {
		return nil, err
	}
	return image, nil

}

func (rep *PostgresRepository) GetPostImages(postId uuid.UUID) ([]*dto.ImageDB, error) {
	var images []*dto.ImageDB

	query := `SELECT * FROM images WHERE post_id = $1;`
	err := rep.DB.Select(&images, query, postId)

	if err != nil {
		return nil, err
	}
	return images, nil
}

func (rep *PostgresRepository) GetPublishedPosts() ([]*dto.PostUserDB, error) {
	var posts []*dto.PostUserDB

	query := `SELECT p.*, u.* FROM posts p
LEFT JOIN users u ON u.user_id = p.author_id
WHERE p.status = 'published';`
	err := rep.DB.Select(&posts, query)

	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (rep *PostgresRepository) GetUserPosts(userId uuid.UUID) ([]*dto.PostUserDB, error) {
	var posts []*dto.PostUserDB

	query := `SELECT p.*, u.* FROM posts p
LEFT JOIN users u ON u.user_id = p.author_id
WHERE p.author_id = $1;`
	err := rep.DB.Select(&posts, query, userId)

	if err != nil {
		return nil, err
	}
	return posts, nil
}
