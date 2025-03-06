// El paquete client contiene la lógica de interacción con el usuario
// así como de comunicación con el servidor.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"prac/pkg/api"
	"prac/pkg/ui"
)

// client estructura interna no exportada que controla
// el estado de la sesión (usuario, token) y logger.
type client struct {
	log              *log.Logger
	currentUser      string
	authToken        string
	currentSpecialty string //nuevo
	currentHospital  string //nuevo

}

// Run es la única función exportada de este paquete.
// Crea un client interno y ejecuta el bucle principal.
func Run() {
	// Creamos un logger con prefijo 'cli' para identificar
	// los mensajes en la consola.
	c := &client{
		log: log.New(os.Stdout, "[cli] ", log.LstdFlags),
	}
	c.runLoop()
}

// runLoop maneja la lógica del menú principal.
// Si NO hay usuario logueado, se muestran ciertas opciones;
// si SÍ hay usuario logueado, se muestran otras.
func (c *client) runLoop() {
	for {
		ui.ClearScreen()

		// Construimos un título que muestre el usuario logueado, si lo hubiera.
		var title string
		if c.currentUser == "" {
			title = "Menú"
		} else {
			title = fmt.Sprintf("Menú (%s)", c.currentUser)
		}

		// Generamos las opciones dinámicamente, según si hay un login activo.
		var options []string
		if c.currentUser == "" {
			// Usuario NO logueado: Registro, Login, Salir
			options = []string{
				"Registrar usuario",
				"Iniciar sesión",
				"Salir",
			}
		} else {
			// Usuario logueado: Ver datos, Actualizar datos, Logout, Salir
			options = []string{
				"Dar de alta paciente",
				"Ver historial del paciente",
				"Cerrar sesión",
				"Salir",
			}
		}

		// Mostramos el menú y obtenemos la elección del usuario.
		choice := ui.PrintMenu(title, options)

		// Hay que mapear la opción elegida según si está logueado o no.
		if c.currentUser == "" {
			// Caso NO logueado
			switch choice {
			case 1:
				c.registerUser()
			case 2:
				c.loginUser()
			case 3:
				// Opción Salir
				c.log.Println("Saliendo del cliente...")
				return
			}
		} else {
			// Caso logueado
			switch choice {
			case 1:
				c.darAltaPaciente()
			case 2:
				c.verHistorialPaciente()
			case 3:
				c.logoutUser()
			case 4:
				// Opción Salir
				c.log.Println("Saliendo del cliente...")
				return
			}
		}

		// Pausa para que el usuario vea resultados.
		ui.Pause("Pulsa [Enter] para continuar...")
	}
}

// registerUser pide credenciales y las envía al servidor para un registro.
// Si el registro es exitoso, se intenta el login automático.
func (c *client) registerUser() {
	ui.ClearScreen()
	fmt.Println("** Registro de usuario **")

	username := ui.ReadInput("Nombre de usuario")
	password := ui.ReadInput("Contraseña")
	apellido := ui.ReadInput("Apellido")
	especialidad := ui.ReadInput("ID de especialidad") //ID?
	hospital := ui.ReadInput("ID de hospital")         //ID???

	// Enviamos la acción al servidor
	res := c.sendRequest(api.Request{
		Action:       api.ActionRegister,
		Username:     username,
		Password:     password,
		Apellido:     apellido,
		Especialidad: especialidad,
		Hospital:     hospital,
	})

	// Mostramos resultado
	fmt.Println("Éxito:", res.Success)
	fmt.Println("Mensaje:", res.Message)

	// Si fue exitoso, probamos loguear automáticamente.
	if res.Success {
		c.log.Println("Registro exitoso; intentando login automático...")

		loginRes := c.sendRequest(api.Request{
			Action:   api.ActionLogin,
			Username: username,
			Password: password,
		})
		if loginRes.Success {
			c.currentUser = username
			c.authToken = loginRes.Token
			c.currentSpecialty = especialidad
			c.currentHospital = hospital
			fmt.Println("Login automático exitoso. Token guardado.")
		} else {
			fmt.Println("No se ha podido hacer login automático:", loginRes.Message)
		}
	}
}

// loginUser pide credenciales y realiza un login en el servidor.
func (c *client) loginUser() {
	ui.ClearScreen()
	fmt.Println("** Inicio de sesión **")

	username := ui.ReadInput("Nombre de usuario")
	password := ui.ReadInput("Contraseña")

	res := c.sendRequest(api.Request{
		Action:   api.ActionLogin,
		Username: username,
		Password: password,
	})

	fmt.Println("Éxito:", res.Success)
	fmt.Println("Mensaje:", res.Message)

	// Si login fue exitoso, guardamos currentUser y el token.
	if res.Success {
		c.currentUser = username
		c.authToken = res.Token

		var userData struct { //datos del usuario serv
			Especialidad string //id
			Hospital     string //id
		}
		json.Unmarshal([]byte(res.Data), &userData) //nuevo
		c.currentSpecialty = userData.Especialidad
		c.currentHospital = userData.Hospital
		fmt.Println("Sesión iniciada con éxito. Token guardado.")
	}
}

func (c *client) darAltaPaciente() {
	ui.ClearScreen()
	fmt.Println("** Dar de alta al paciente **")

	nombre := ui.ReadInput("Nombre: ")
	apellido := ui.ReadInput("Apellido: ")
	fecha_nacimiento := ui.ReadInput("Fecha de nacimiento (dd-MM-AAAA)")
	fecha := idstrings.split(fecha_nacimiento, "-")
	if len(fechaParts) != 3 {
		fmt.Println("Formato de fecha inválido")
		return
	}
	time.Date(fecha[2], fecha[0], fecha[1], 0, 0, 0, 0, time.Local)
	fecha.Format(time.DateOnly)

	sexo := ui.ReadInput("Sexo (H,M,O)")
	//hospital:=
	//historial:=
	//medico:=

	// Enviamos la acción al servidor
	res := c.sendRequest(api.Request{
		Action:     api.ActionAltaPaciente,
		Token:      c.authToken,
		Nombre:     nombre,
		Apellido:   apellido,
		Nacimiento: fecha,
		Sexo:       sexo,
		Hospital:   c.currentHospital,
		MedicoID:   c.currentUser,
	})

	// Mostramos resultado
	fmt.Println("Éxito:", res.Success)
	fmt.Println("Mensaje:", res.Message)
}

func (c *client) verHistorialPaciente() { //mirar
	ui.ClearScreen()
	fmt.Println("** Ver historial del paciente **")

	nombre := ui.ReadInput("Nombre del paciente")
	apellido := ui.ReadInput("Apellido del paciente")

	res := c.sendRequest(api.Request{
		Action:   api.ActionGetHistorial,
		Token:    c.authToken,
		Nombre:   nombre,
		Apellido: apellido,
	})

	if !res.Success {
		fmt.Println("Mensaje:", res.Message)
		if ui.Confirm("¿Desea dar de alta al paciente? (s/n)") { //mas novedoso que choices
			c.darAltaPaciente()
		}
		return
	}

	// Parsear expedientes del historial
	var historial struct {
		Expedientes []struct {
			ID            string
			Observaciones string
			Fecha         string
		}
	}
	err := json.Unmarshal([]byte(res.Data), &historial)
	if err != nil {
		fmt.Println("Error al procesar historial:", err)
		return
	}

	if len(historial.Expedientes) == 0 {
		fmt.Println("No hay expedientes para este paciente")
		return
	}

	for {
		ui.ClearScreen()
		fmt.Printf("Expedientes de %s %s:\n", nombre, apellido)
		options := make([]string, len(historial.Expedientes))
		for i, exp := range historial.Expedientes {
			options[i] = fmt.Sprintf("%d. %s - %s", i+1, exp.Fecha, exp.Observaciones[:min(20, len(exp.Observaciones))])
		}
		options = append(options, "Salir")

		choice := ui.PrintMenu("Seleccionar expediente", options)
		if choice == len(options) {
			return
		}

		selectedExp := historial.Expedientes[choice-1]
		c.manejarExpediente(selectedExp.ID, nombre, apellido)
	}
}

func (c *client) manejarExpedientes(expId, nombre, apellido string) {

	ui.ClearScreen()
	options := []string{
		"Ver Expedientes",
		"Modificar expediente",
		"Salir",
	}

	eleccion := ui.PrintMenu(fmt.Sprintf("Expediente de %s %s", nombre, apellido), options)

	switch eleccion {
	case 1:
		res := c.sendRequest(api.Request{
			Action:       api.ActionGetExpediente,
			Token:        c.authToken,
			ExpedienteID: expID,
		})
		fmt.Println("Éxito:", res.Success)
		fmt.Println("Detalles:", res.Data)
	case 2:
		observacion := ui.ReadInput("Nueva observación: ") //mirar lo de las fechas de modificacion
		res := c.sendRequest(api.Request{
			Action:        api.ActionModificarExpediente,
			Token:         c.authToken,
			ExpedienteID:  expID,
			Observaciones: observacion,
		})
		fmt.Println("Éxito:", res.Success)
		fmt.Println("Mensaje:", res.Message)
		if res.Success { //podemos poner un confirm
			fmt.Println("Edición confirmada")
		}
	case 3:
		return
	}
}

// fetchData pide datos privados al servidor.
// El servidor devuelve la data asociada al usuario logueado.
func (c *client) fetchData() {
	ui.ClearScreen()
	fmt.Println("** Obtener datos del usuario **")

	// Chequeo básico de que haya sesión
	if c.currentUser == "" || c.authToken == "" {
		fmt.Println("No estás logueado. Inicia sesión primero.")
		return
	}

	// Hacemos la request con ActionFetchData
	res := c.sendRequest(api.Request{
		Action:   api.ActionFetchData,
		Username: c.currentUser,
		Token:    c.authToken,
	})

	fmt.Println("Éxito:", res.Success)
	fmt.Println("Mensaje:", res.Message)

	// Si fue exitoso, mostramos la data recibida
	if res.Success {
		fmt.Println("Tus datos:", res.Data)
	}
}

// updateData pide nuevo texto y lo envía al servidor con ActionUpdateData.
func (c *client) updateData() {
	ui.ClearScreen()
	fmt.Println("** Actualizar datos del usuario **")

	if c.currentUser == "" || c.authToken == "" {
		fmt.Println("No estás logueado. Inicia sesión primero.")
		return
	}

	// Leemos la nueva Data
	newData := ui.ReadInput("Introduce el contenido que desees almacenar")

	// Enviamos la solicitud de actualización
	res := c.sendRequest(api.Request{
		Action:   api.ActionUpdateData,
		Username: c.currentUser,
		Token:    c.authToken,
		Data:     newData,
	})

	fmt.Println("Éxito:", res.Success)
	fmt.Println("Mensaje:", res.Message)
}

// logoutUser llama a la acción logout en el servidor, y si es exitosa,
// borra la sesión local (currentUser/authToken).
func (c *client) logoutUser() {
	ui.ClearScreen()
	fmt.Println("** Cerrar sesión **")

	if c.currentUser == "" || c.authToken == "" {
		fmt.Println("No estás logueado.")
		return
	}

	// Llamamos al servidor con la acción ActionLogout
	res := c.sendRequest(api.Request{
		Action:   api.ActionLogout,
		Username: c.currentUser,
		Token:    c.authToken,
	})

	fmt.Println("Éxito:", res.Success)
	fmt.Println("Mensaje:", res.Message)

	// Si fue exitoso, limpiamos la sesión local.
	if res.Success {
		c.currentUser = ""
		c.authToken = ""
	}
}

// sendRequest envía un POST JSON a la URL del servidor y
// devuelve la respuesta decodificada. Se usa para todas las acciones.
func (c *client) sendRequest(req api.Request) api.Response {
	jsonData, _ := json.Marshal(req)
	resp, err := http.Post("http://localhost:8080/api", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error al contactar con el servidor:", err)
		return api.Response{Success: false, Message: "Error de conexión"}
	}
	defer resp.Body.Close()

	// Leemos el body de respuesta y lo desempaquetamos en un api.Response
	body, _ := io.ReadAll(resp.Body)
	var res api.Response
	_ = json.Unmarshal(body, &res)
	return res
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
