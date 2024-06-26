package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/South-s-Eagles/EEG-electroencephalogram/broker"
	"github.com/South-s-Eagles/EEG-electroencephalogram/database"
	"github.com/South-s-Eagles/EEG-electroencephalogram/dispositivo"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	sleepTime     = 2
	dataRetention = 5 * time.Second
	maxRetries    = 3
)

type ExternalPayload struct {
	DispositivoId     string `json:"dispositivoId"`
	Dispositivo       string `json:"dispositivo"`
	Valor             int16  `json:"valor"`
	UnidadeMedida     string `json:"unidadeMedida"`
	ConteudoAdicional string `json:"conteudoAdicional"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	databaseClient, err := database.NewMongoClient(ctx)
	if err != nil {
		panic(err)
	}
	defer databaseClient.Database().Client().Disconnect(ctx)

	azureBroker, err := broker.NewAzureBroker(ctx, "HostName=EEG-Simulator.azure-devices.net;DeviceId=eeg-simulator;SharedAccessKey=sxX+gQDWSpkpFbNfZ8xa1rifHRlO8n96aAIoTLnFe4I=")
	if err != nil {
		panic(err)
	}

	device, err := dispositivo.NewDispositivo(8)
	if err != nil {
		panic(err)
	}

	dataChan := make(chan *dispositivo.Dispositivo)

	go sendDataToDatabase(ctx, dataChan, databaseClient, azureBroker)

	for {
		device.Run()

		select {
		case <-ctx.Done():
			return
		case dataChan <- device:
		}

		time.Sleep(sleepTime * time.Second)
	}
}

func sendDataToDatabase(ctx context.Context, dataChan <-chan *dispositivo.Dispositivo, client *mongo.Collection, broker *broker.AzureBroker) {
	var dataBuffer []*dispositivo.Dispositivo
	timer := time.NewTimer(dataRetention)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case newData := <-dataChan:
			dataBuffer = append(dataBuffer, newData)
		case <-timer.C:
			fmt.Println("Enviando os dados para armazenamento")

			devicePayload, err := json.Marshal(dataBuffer)
			if err != nil {
				fmt.Println("Erro ao serializar devicePayload:", err)
				continue
			}

			payload := &ExternalPayload{
				DispositivoId:     dataBuffer[len(dataBuffer)-1].ID,
				Dispositivo:       "electroencephalogram",
				Valor:             10,
				UnidadeMedida:     "Hz",
				ConteudoAdicional: string(devicePayload),
			}

			finalPayload, err := json.Marshal(payload)
			if err != nil {
				fmt.Println("Erro ao serializar devicePayload:", err)
				continue
			}

			// client.InsertOne(ctx, payload)
			err = broker.SendMessage(finalPayload)
			fmt.Println(err)

			dataBuffer = nil
			timer.Reset(dataRetention)
		}
	}
}
