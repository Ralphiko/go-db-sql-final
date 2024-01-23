package main

import (
	"database/sql"
	"errors"
)

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	result, err := s.db.Exec("INSERT INTO parcel (client, status, address) VALUES (?, ?, ?)", p.Client, p.Status, p.Address)
	if err != nil {
		return 0, err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(lastID), nil
}

func (s ParcelStore) Get(number int) (Parcel, error) {
	row := s.db.QueryRow("SELECT * FROM parcel WHERE number = ?", number)

	var id, client int
	var status, address string

	err := row.Scan(&id, &client, &status, &address)
	if err != nil {
		return Parcel{}, err
	}

	p := Parcel{
		ID:      id,
		Client:  client,
		Status:  status,
		Address: address,
	}

	return p, nil
}

func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query("SELECT * FROM parcel WHERE client = ?", client)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []Parcel

	for rows.Next() {
		var id, client int
		var status, address string

		err := rows.Scan(&id, &client, &status, &address)
		if err != nil {
			return nil, err
		}

		p := Parcel{
			ID:      id,
			Client:  client,
			Status:  status,
			Address: address,
		}
		res = append(res, p)
	}

	return res, nil
}

func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", status, number)
	return err
}

func (s ParcelStore) SetAddress(number int, address string) error {
	var currentStatus string
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number).Scan(&currentStatus)
	if err != nil {
		return err
	}

	if currentStatus != "registered" {
		return errors.New("нельзя менять адрес для посылки со статусом " + currentStatus)
	}

	_, err = s.db.Exec("UPDATE parcel SET address = ? WHERE number = ?", address, number)
	return err
}

func (s ParcelStore) Delete(number int) error {
	var currentStatus string
	err := s.db.QueryRow("SELECT status FROM parcel WHERE number = ?", number).Scan(&currentStatus)
	if err != nil {
		return err
	}

	if currentStatus != "registered" {
		return errors.New("нельзя удалять посылку со статусом " + currentStatus)
	}

	_, err = s.db.Exec("DELETE FROM parcel WHERE number = ?", number)
	return err
}
