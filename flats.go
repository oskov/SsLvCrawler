package main

import (
	"database/sql"
	"fmt"
	"os"
)

type Flat struct {
	Id            int
	Text          string
	City          string
	District      string
	Street        string
	Rooms         int
	ApartmentArea int
	Floor         string
	HouseType     string
	Price         int
	Type          string
	Url           string
}

type FlatStorage interface {
	Put(flat Flat) bool
	GetAll() (flats []Flat)
	ToSql() (sql string, sqlParams []interface{})
	Save(db *sql.DB)
}

type flatStorage struct {
	flats []Flat
}

func NewFlatStorage() FlatStorage {
	return &flatStorage{flats: make([]Flat, 0)}
}

func (f *flatStorage) Put(flat Flat) bool {
	f.flats = append(f.flats, flat)
	return true
}

func (f *flatStorage) GetAll() []Flat {
	return f.flats
}

func (f *flatStorage) ToSql() (sql string, sqlParams []interface{}) {
	sql = "INSERT IGNORE INTO flats (" +
		"id_external, " +
		"text, " +
		"city, " +
		"district, " +
		"street, " +
		"rooms, " +
		"apartment_area, " +
		"floor, " +
		"house_type, " +
		"price, " +
		"type, " +
		"url, " +
		"added_dt) VALUES"
	sqlParams = make([]interface{}, 0)
	for _, v := range f.flats {
		sql = sql + "(?, ?, ? ,? ,? ,? ,? ,? , ?, ?, ?, ?, ?),"
		flat := []interface{}{
			v.Id,
			v.Text,
			v.City,
			v.District,
			v.Street,
			v.Rooms,
			v.ApartmentArea,
			v.Floor,
			v.HouseType,
			v.Price,
			v.Type,
			v.Url,
			CurrentDateTime()}
		sqlParams = append(sqlParams, flat...)
	}
	sql = sql[:len(sql)-1] + ";"
	return sql, sqlParams
}

func (f *flatStorage) Save(db *sql.DB) {
	if len(f.flats) == 0 {
		return
	}

	sqlQuery, sqlParams := f.ToSql()
	statement, err := db.Prepare(sqlQuery)
	if err != nil {
		fmt.Println("Unable to prepare statement")
		fmt.Println(err)
		os.Exit(-1)
	}
	res, err := statement.Exec(sqlParams...)
	if err != nil {
		fmt.Println("Unable to insert data")
		fmt.Println(err)
		os.Exit(-1)
	} else {
		fmt.Println(res)
		statement.Close()
	}
}
