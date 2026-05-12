package models

import "time"

type Article struct {
	Id       int
	Title    string
	Anons    string
	FullText string
	UserId   int
}

type User struct {
	Id int
}

type Session struct {
	UserId    int
	CreatedAt time.Time
	ExpiresAt time.Time
}

type ArticleWithAuthor struct {
	Article
	AuthorName string
}