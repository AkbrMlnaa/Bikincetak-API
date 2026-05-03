package config

import (
	"fmt"
	"os"

	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

var SnapClient snap.Client

func ConnectMidtrans() {
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	if serverKey == "" {
		fmt.Println("Peringatan: MIDTRANS_SERVER_KEY belum di-set di .env")
	}

	SnapClient.New(serverKey, midtrans.Sandbox)
	
	fmt.Println("Berhasil connect ke Midtrans snap client")
}