package main

import (
	"encoding/json"
	"log"
	"net/http"
	"fmt"
	"net"
	"github.com/gorilla/mux"
	"sync"
	"bufio"
)

type Contact struct {
	Name string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}
type Planta struct {
	S_lenght float64 `json:"s_lenght"`
	S_width float64 `json:"s_width"`
	P_lenght float64 `json:"p_lenght"`
	P_width float64 `json:"p_width"`
	Plant_type string `json:"plant_type"`
}

var contacts []Contact
var wg sync.WaitGroup
func main(){
	r := mux.NewRouter()

	contacts = append(contacts, Contact{Name: "Friend_1", Phone: "989999999", Email: "123@gmail.com"})
	contacts = append(contacts, Contact{Name: "Friend_2", Phone: "979999999", Email: "456@gmail.com"})
	contacts = append(contacts, Contact{Name: "Friend_3", Phone: "969999999", Email: "789@gmail.com"})

	r.HandleFunc("/contacts", getContacts).Methods("GET")
	r.HandleFunc("/contacts/{name}", getContact).Methods("GET")
	r.HandleFunc("/contacts", createContact).Methods("POST")
	r.HandleFunc("/contacts/{name}", updateContact).Methods("PUT")
	r.HandleFunc("/contacts/{name}", deleteContact).Methods("DELETE")
	r.HandleFunc("/plant", enviarAPuerto8000).Methods("POST")

	log.Fatal(http.ListenAndServe(":3000", r))
}

//Listar todos los contactos
func getContacts(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contacts)
}

//Buscar un contacto por nombre
func getContact(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for _, item := range contacts {
		if item.Name == params["name"]{
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	json.NewEncoder(w).Encode(&Contact{})
}

//Agregar un nuevo contacto
func createContact(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	var contact Contact
	_ = json.NewDecoder(r.Body).Decode(&contact)
	contacts = append(contacts, contact)
	json.NewEncoder(w).Encode(contact)
}

func enviarAPuerto8000(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	var plant Planta
	_ = json.NewDecoder(r.Body).Decode(&plant)
	conn, _ := net.Dial("tcp", "localhost:8000")
	defer conn.Close()
	fmt.Println(plant)
	jsonBytes, _ := json.Marshal(plant)
	fmt.Fprintf(conn, "%s\n", string(jsonBytes))

	reccon,_:=net.Listen("tcp","localhost:3001")
	defer reccon.Close()
	for i:=0;i<1;i++{
		accept,_:=reccon.Accept()
		defer accept.Close()
		reader:=bufio.NewReader(accept)
		jsonString, _ := reader.ReadString('\n')
		var plant Planta
		json.Unmarshal([]byte(jsonString), &plant)
		fmt.Println("Recibido: ", plant)
		json.NewEncoder(w).Encode(plant)
	}
}

//Eliminar un contacto
func deleteContact(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for idx, item := range contacts {
		if item.Name == params["name"] {
			contacts = append(contacts[:idx], contacts[idx+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(contacts)
}

//Actualizar contacto
func updateContact(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for idx, item := range contacts {
		if item.Name == params["name"] {
			contacts = append(contacts[:idx], contacts[idx+1:]...)
			var contact Contact
			_ = json.NewDecoder(r.Body).Decode(&contact)
			contact.Name = params["name"]
			contacts = append(contacts, contact)
			json.NewEncoder(w).Encode(contact)
			return
		}
	}
}