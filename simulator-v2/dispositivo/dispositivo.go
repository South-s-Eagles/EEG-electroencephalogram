// Struct do dispositivo
// tem todos os componentes do dispositivo
// tanto os sensores quando os dispositivo IOT si
package dispositivo

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/South-s-Eagles/EEG-electroencephalogram/commons"
)

// Dispositivo em si do EEG
// Ele é a "caixa preta" que vai ter os
// sensores, vai ser o dispositivo IOT de fato,
// que vai conseguir enviar os dados tratados
// conseguir ter a inteligência de tratar os dados
// recebidos pelo sensor e guardar para enviar com
// calma para o broker (IOT HUB)
type Dispositivo struct {
	Sensores   []Sensor `json:"sensores"`   // Sensores que existem no Dispositivo
	Bateria    float32  `json:"bateria"`    // Quantidade de bateria em porcentagem
	Frequencia float32  `json:"frequencia"` // Valor em hz da Frequencia da corrente elétrica
	Amplitude  float32  `json:"Amplitude"`  // Valor da Amplitude da onda da corrente elétrica
	Ligado     bool     `json:"ligado"`     // Estado do Dispositivo se está ligado ou desligado
}

// Lista todos os sensores no dispositivo
func (d *Dispositivo) ListarSensores() {
	for _, v := range d.Sensores {
		fmt.Println(v.toString())
	}
}

// Criar um novo dispositivo com a quantidade de sensor passada
func NewDispositivo(sensorQtd int8) (*Dispositivo, error) {
	if sensorQtd > 10 {
		return nil, errors.New("não é possível criar um dispositivo com mais de 10 sensores")
	}

	sensores := criarSensoresEmLote(int(sensorQtd))

	d := &Dispositivo{
		Sensores:   sensores,
		Bateria:    100.0,
		Frequencia: 0,
		Amplitude:  0,
		Ligado:     true,
	}

	return d, nil
}

// Envia os dados em heartz recebidos para o broker
// NOTE: Sera que faz sentido isso aqui ainda?
func (d *Dispositivo) enviarMensagem(arr []byte) {
	fmt.Println("dados enviados para o broker")
	d.reduzirBateria(0.1)
}

// Reduz a bateria
func (d *Dispositivo) reduzirBateria(value float32) {
	d.Bateria -= value
	err := d.validarBateria()
	if err != nil {
		fmt.Println("Desligando dispositvo")
		d.desligarDevice()
	}
}

// Validar a vida da bateria.
func (d *Dispositivo) validarBateria() error {
	if d.Bateria <= 0 {
		d.Ligado = false
		return errors.New("bateria descarregada")
	}
	return nil
}

// Desliga o dispositivo
func (d *Dispositivo) desligarDevice() {
	d.Ligado = false
}

// Simula o dispositivo ligado
func (d *Dispositivo) Run() {
	var wg sync.WaitGroup
	var lock sync.Mutex
	var totalData [][]int8

	for _, sensor := range d.Sensores {
		wg.Add(1)
		go func(s Sensor) {
			defer wg.Done()

			var data []int8
			for i := 0; i < 1000; i++ {
				s.gerarValor()
				data = append(data, s.Valor)
				time.Sleep(1 * time.Millisecond)
			}

			lock.Lock()
			defer lock.Unlock()
			totalData = append(totalData, data)
		}(sensor)
	}

	wg.Wait()

	for i, data := range totalData {
		freq, amp := commons.FrequenciaDominante(data)
		fmt.Printf("Frequência do sensor %d: %f\n", i, freq)
		fmt.Printf("Amplitude do sensor %d: %f\n", i, amp)
	}
}
