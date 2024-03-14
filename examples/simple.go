package main

import (
	"fmt"
	"github.com/Eitol/patentes_cl"
)

func main() {
	client := patentes_cl.NewClient()
	rut := "27029012" // Sustituye esto con un RUT real
	vehicles, err := client.GetByRut(rut)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Vehicles:", vehicles)
}
