package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

const (
	ParcelStatusRegistered = "registered"
	ParcelStatusSent       = "sent"
	ParcelStatusDelivered  = "delivered"
)

type Parcel struct {
	Number    int
	Client    int
	Status    string
	Address   string
	CreatedAt string
}

type ParcelStore struct {
	db *sql.DB
}

func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

func (s ParcelStore) Add(p Parcel) (int, error) {
	result, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		p.Client, p.Status, p.Address, p.CreatedAt)
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
	var status, address, createdAt string

	err := row.Scan(&id, &client, &status, &address, &createdAt)
	if err != nil {
		return Parcel{}, err
	}

	p := Parcel{
		Number:    id,
		Client:    client,
		Status:    status,
		Address:   address,
		CreatedAt: createdAt,
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
		var status, address, createdAt string

		err := rows.Scan(&id, &client, &status, &address, &createdAt)
		if err != nil {
			return nil, err
		}

		p := Parcel{
			Number:    id,
			Client:    client,
			Status:    status,
			Address:   address,
			CreatedAt: createdAt,
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

type ParcelService struct {
	store ParcelStore
}

func NewParcelService(store ParcelStore) ParcelService {
	return ParcelService{store: store}
}

func (s ParcelService) Register(client int, address string) (Parcel, error) {
	parcel := Parcel{
		Client:    client,
		Status:    ParcelStatusRegistered,
		Address:   address,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	id, err := s.store.Add(parcel)
	if err != nil {
		return parcel, err
	}

	parcel.Number = id

	fmt.Printf("Новая посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s\n",
		parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt)

	return parcel, nil
}

func (s ParcelService) ChangeAddress(number int, address string) error {
	return s.store.SetAddress(number, address)
}

func (s ParcelService) NextStatus(number int) error {
	return s.store.SetStatus(number, ParcelStatusSent)
}

func (s ParcelService) PrintClientParcels(client int) error {
	parcels, err := s.store.GetByClient(client)
	if err != nil {
		return err
	}

	fmt.Printf("Посылки клиента %d:\n", client)
	for _, parcel := range parcels {
		fmt.Printf("Посылка № %d на адрес %s от клиента с идентификатором %d зарегистрирована %s, статус %s\n",
			parcel.Number, parcel.Address, parcel.Client, parcel.CreatedAt, parcel.Status)
	}
	fmt.Println()

	return nil
}

func (s ParcelService) Delete(number int) error {
	return s.store.Delete(number)
}

func main() {
	db, err := sql.Open("sqlite", "tracker.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	service := NewParcelService(store)

	client := 1
	address := "Псков, д. Пушкина, ул. Колотушкина, д. 5"
	p, err := service.Register(client, address)
	if err != nil {
		log.Println(err)
		return
	}

	newAddress := "Саратов, д. Верхние Зори, ул. Козлова, д. 25"
	err = service.ChangeAddress(p.Number, newAddress)
	if err != nil {
		log.Println(err)
		return
	}

	err = service.NextStatus(p.Number)
	if err != nil {
		log.Println(err)
		return
	}

	err = service.PrintClientParcels(client)
	if err != nil {
		log.Println(err)
		return
	}

	err = service.Delete(p.Number)
	if err != nil {
		log.Println(err)
		return
	}

	err = service.PrintClientParcels(client)
	if err != nil {
		log.Println(err)
		return
	}

	p, err = service.Register(client, address)
	if err != nil {
		log.Println(err)
		return
	}

	err = service.Delete(p.Number)
	if err != nil {
		log.Println(err)
		return
	}

	err = service.PrintClientParcels(client)
	if err != nil {
		log.Println(err)
		return
	}
}
