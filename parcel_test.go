package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	randSource = rand.NewSource(time.Now().UnixNano())
	randRange  = rand.New(randSource)
)

func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func ConnectDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "tracker.db")
	return db, err
}

func TestAddGetDelete(t *testing.T) {
	db, err := ConnectDb()
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	parcel.Number, err = store.Add(parcel)

	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)
	stored, err := store.Get(parcel.Number)

	require.NoError(t, err)
	assert.Equal(t, parcel, stored)
	err = store.Delete(parcel.Number)
	require.NoError(t, err)

	_, err = store.Get(parcel.Number)
	require.Equal(t, sql.ErrNoRows, err)
}

func TestSetAddress(t *testing.T) {
	db, err := ConnectDb()
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	parcel.Number, err = store.Add(parcel)

	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)
	newAddress := "new test address"
	err = store.SetAddress(parcel.Number, newAddress)
	require.NoError(t, err)

	stored, err := store.Get(parcel.Number)

	require.NoError(t, err)
	assert.Equal(t, newAddress, stored.Address)
}

func TestSetStatus(t *testing.T) {
	db, err := ConnectDb()
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)
	parcel := getTestParcel()

	parcel.Number, err = store.Add(parcel)

	require.NoError(t, err)
	require.NotEmpty(t, parcel.Number)
	err = store.SetStatus(parcel.Number, ParcelStatusDelivered)
	require.NoError(t, err)

	stored, err := store.Get(parcel.Number)

	require.NoError(t, err)
	assert.Equal(t, ParcelStatusDelivered, stored.Status)
}

func TestGetByClient(t *testing.T) {
	db, err := ConnectDb()
	if err != nil {
		require.NoError(t, err)
	}
	defer db.Close()

	store := NewParcelStore(db)

	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])

		require.NoError(t, err)
		require.NotEmpty(t, id)

		parcels[i].Number = id
		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Equal(t, len(storedParcels), len(parcels))

	for _, parcel := range storedParcels {
		assert.Equal(t, parcel, parcelMap[parcel.Number])
	}
}
