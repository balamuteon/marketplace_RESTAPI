package models

import "time"

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=4,max=32"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type CreateAdRequest struct {
	Title       string  `json:"title" binding:"required,min=1,max=100"`
	Description string  `json:"description" binding:"required,max=1000"`
	Price       float64 `json:"price" binding:"required,gte=0"`
	ImageURL    string  `json:"image_url" binding:"omitempty,url"`
}

type CreateAdResponse struct {
	ID int64 `json:"id"`
}

type AdResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	ImageURL    string    `json:"image_url"`
	AuthorID    int64     `json:"author_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type AdsQuery struct {
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=10"`
	SortBy    string `form:"sort_by,default=created_at"` // 'created_at' or 'price'
	SortOrder string `form:"sort_order,default=desc"`    // 'asc' or 'desc'
}

type UpdateAdRequest struct {
	Title       *string  `json:"title,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty"`
}
