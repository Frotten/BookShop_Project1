package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/models"
	"encoding/json"
)

func AddBook(p *models.AddBookParam) models.ResCode {
	tagsJSON, _ := json.Marshal(p.Tags)
	book := &models.Book{
		Title:      p.Title,
		Author:     p.Author,
		Publisher:  p.Publisher,
		Stock:      p.Stock,
		Sales:      0,
		Price:      p.Price,
		Score:      0,
		CoverImage: p.CoverImage,
		Tags:       tagsJSON,
	}
	err := mysql.AddBook(book)
	if err != nil {
		return models.CodeServerBusy
	}
	return models.CodeSuccess
}

func BookToCache(book *models.Book) (*models.BookCache, error) {
	var tags []string
	if len(book.Tags) > 0 {
		if err := json.Unmarshal(book.Tags, &tags); err != nil {
			return nil, err
		}
	}
	return &models.BookCache{
		BookID:     book.BookID,
		Title:      book.Title,
		Author:     book.Author,
		Publisher:  book.Publisher,
		Stock:      book.Stock,
		Sales:      book.Sales,
		Price:      book.Price,
		Score:      book.Score,
		CoverImage: book.CoverImage,
		Tags:       tags,
	}, nil
}
