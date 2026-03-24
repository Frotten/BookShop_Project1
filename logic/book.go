package logic

import (
	"Project1_Shop/dao/mysql"
	"Project1_Shop/models"
	"encoding/json"
)

func AddBook(p *models.AddBookParam) models.ResCode {
	tagsJSON, _ := json.Marshal(p.Tags)
	exist := mysql.ExistBookByInfo(p.Title, p.Author, p.Publisher)
	if exist {
		return models.CodeBookExist
	}
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

func BookListToCache(book *models.Book) (*models.ListBook, error) {
	var tags []string
	if len(book.Tags) > 0 {
		if err := json.Unmarshal(book.Tags, &tags); err != nil {
			return nil, err
		}
	}
	return &models.ListBook{
		BookID: book.BookID,
		Title:  book.Title,
		Sales:  book.Sales,
		Score:  float64(book.Score) / 100,
		Tags:   tags,
	}, nil
}

func GetBookByID(ID int64) (*models.Book, error) {
	book, err := mysql.GetBookByID(ID)
	if err != nil {
		return nil, err
	}
	return book, nil
}

func DeleteBook(ID int64) models.ResCode {
	exist := mysql.ExistBook(ID)
	if !exist {
		return models.CodeBookNotExist
	}
	_ = mysql.DeleteBook(ID)
	exist = mysql.ExistBook(ID)
	if !exist {
		return models.CodeSuccess
	}
	return models.CodeServerBusy
}

func UpdateBook(B *models.UpdateBookParam) models.ResCode {
	tagsJSON, _ := json.Marshal(B.Tags)
	book := &models.Book{
		BookID:     B.BookID,
		Title:      B.Title,
		Author:     B.Author,
		Publisher:  B.Publisher,
		Stock:      B.Stock,
		Price:      B.Price,
		CoverImage: B.CoverImage,
		Tags:       tagsJSON,
	}
	err := mysql.UpdateBook(book)
	if err != nil {
		return models.CodeServerBusy
	}
	return models.CodeSuccess
}
