package main

import (
	"math/rand"
	"testing"
	"time"

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
func TestAddGetDelete(t *testing.T) {
	db, err := tracker.db()
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.AddParcel(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	retrievedParcel, err := store.GetParcel(id)
	require.NoError(t, err)
	require.Equal(t, parcel.Client, retrievedParcel.Client)
	require.Equal(t, parcel.Status, retrievedParcel.Status)
	require.Equal(t, parcel.Address, retrievedParcel.Address)
	require.WithinDuration(t, time.Now().UTC(), retrievedParcel.CreatedAt, time.Second)

	err = store.DeleteParcel(id)
	require.NoError(t, err)

	_, err = store.GetParcel(id)
	require.Error(t, err)
}

func TestSetAddress(t *testing.T) {
	db, err := tracker.db()
	require.NoError(t, err)
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.AddParcel(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	newAddress := "new test address"
	err = store.SetAddress(id, newAddress)
	require.NoError(t, err)

	retrievedParcel, err := store.GetParcel(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, retrievedParcel.Address)
}

func TestSetStatus(t *testing.T) {
	db, err := tracker.db()
	require.NoError(t, err)
	store := NewParcelStore(db)
	parcel := getTestParcel()

	id, err := store.AddParcel(parcel)
	require.NoError(t, err)
	require.NotEmpty(t, id)

	newStatus := ParcelStatusDelivered
	err = store.SetStatus(id, newStatus)
	require.NoError(t, err)

	retrievedParcel, err := store.GetParcel(id)
	require.NoError(t, err)
	require.Equal(t, newStatus, retrievedParcel.Status)
}

func TestGetByClient(t *testing.T) {
	db, err := tracker.db()
	require.NoError(t, err)
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
		id, err := store.AddParcel(parcels[i])
		require.NoError(t, err)

		parcels[i].Number = id

		parcelMap[id] = parcels[i]
	}

	storedParcels, err := store.GetParcelsByClient(client)
	require.NoError(t, err)

	require.Equal(t, len(parcels), len(storedParcels))

	for _, storedParcel := range storedParcels {
		originalParcel, ok := parcelMap[storedParcel.Number]
		require.True(t, ok)

		require.Equal(t, originalParcel.Client, storedParcel.Client)
		require.Equal(t, originalParcel.Status, storedParcel.Status)
		require.Equal(t, originalParcel.Address, storedParcel.Address)
	}
}
