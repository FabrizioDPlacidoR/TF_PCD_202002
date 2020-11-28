package main

import (
	"bufio"
	//"encoding/csv"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Planta struct {
	S_lenght float64 `json:"s_lenght"`
	S_width float64 `json:"s_width"`
	P_lenght float64 `json:"p_lenght"`
	P_width float64 `json:"p_width"`
	Plant_type string `json:"plant_type"`
}

var (
	remotehost []string
	wg         sync.WaitGroup
	wgTrain sync.WaitGroup
	chInfo chan map[string]int
	cSeto int
	cVersi int
	cVirgi int
	plant Planta
)

func enviar(plant Planta, nodo int) {
	conn, _ := net.Dial("tcp", remotehost[nodo])
	defer conn.Close()
	fmt.Println(plant)
	jsonBytes, _ := json.Marshal(plant)
	fmt.Fprintf(conn, "%s\n", string(jsonBytes))
}

func manejador(con net.Conn) {

	defer con.Close()
	r := bufio.NewReader(con)
	jsonString, _ := r.ReadString('\n')
	json.Unmarshal([]byte(jsonString), &plant)
	fmt.Println("Recibido: ", plant)
	if (plant.Plant_type=="setosa"){
		cSeto+=1
	}
	if (plant.Plant_type=="versicolor"){
		cVersi+=1
	}
	if (plant.Plant_type=="virginica"){
		cVirgi+=1
	}
	wg.Done()
}
func recepcion(con net.Conn){
	defer con.Close()
	r := bufio.NewReader(con)
	fmt.Print("Recibido: ")
	jsonString, _ := r.ReadString('\n')
	print(jsonString)
	var planta Planta
	json.Unmarshal([]byte(jsonString), &planta)
	for i := 0; i < 4; i++ {
		enviar(planta, i)
	}
}
func manejaTrain(con net.Conn, error_chan chan float64){
	defer con.Close()
	r:=bufio.NewReader(con)
	msg,_:=r.ReadString('\n')
	fmt.Printf("Recibido: %s\n",msg)
	val,_:=strconv.ParseFloat(msg,64)
	error_chan<-val
	wgTrain.Done()
}
func train(){
	fmt.Println("Esperando resultados de entrenamiento...")
	error_val:=0.0
	error_chan:=make(chan float64)
	wgTrain.Add(4)
	ln, _ := net.Listen("tcp", "localhost:8000")
	defer ln.Close()
	for i := 0; i < 4; i++ {
		con, _ := ln.Accept()
		go manejaTrain(con, error_chan)
	}
	e1,e2,e3,e4:=<-error_chan,<-error_chan,<-error_chan,<-error_chan
	error_val+=e1
	error_val+=e2
	error_val+=e3
	error_val+=e4
	wgTrain.Wait()
	fmt.Println("MSE promediado: ",float64(error_val/4.0))
}
func sendResponse(plant Planta){
	conn, _ := net.Dial("tcp", "localhost:3001")
	max:=0
	defer conn.Close()
	if (cSeto>=cVersi){
		if (cSeto>cVirgi){
			max=0
		}else{
			max=2
		}
	}else{
		if (cVersi>cVirgi){
			max=1
		}else{
			max=2
		}
	}
	if (max==0){
		plant.Plant_type="setosa"
	}
	if (max==1){
		plant.Plant_type="versicolor"
	}
	if (max==2){
		plant.Plant_type="virginica"
	}
	cSeto=0
	cVersi=0
	cVirgi=0
	fmt.Println(plant)
	jsonBytes, _ := json.Marshal(plant)
	fmt.Fprintf(conn, "%s\n", string(jsonBytes))
}
func main() {

	r := bufio.NewReader(os.Stdin)

	fmt.Print("Puerto escucha: ")
	port, _ := r.ReadString('\n')
	port = strings.TrimSpace(port)
	hostname := fmt.Sprintf("localhost:%s", port)

	for i := 0; i < 4; i++ {
		fmt.Printf("Puerto de nodo %d: ", i+1)
		port, _ := r.ReadString('\n')
		port = strings.TrimSpace(port)
		remotehost = append(remotehost, fmt.Sprintf("localhost:%s", port))
	}

	train()
	ln, _ := net.Listen("tcp", hostname)
	defer ln.Close()

	for {
		con,_:=ln.Accept()
		recepcion(con)
		
		wg.Add(4)
		for i := 0; i < 4; i++ {
			con, _ := ln.Accept()
			go manejador(con)
		}
		wg.Wait()
		sendResponse(plant)

	}
}
